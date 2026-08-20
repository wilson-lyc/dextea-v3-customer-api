// Package log 提供结合 OpenTelemetry 的日志能力：
//
//   - 本地用 zap 输出结构化 JSON（stderr），每条日志自动携带 trace_id / span_id；
//   - 远端通过 OTLP 把日志作为 LogRecord 上报到采集后端（如 Grafana Loki / Tempo /
//     Jaeger），与同一条链路的 trace 在后端按 trace_id 关联；
//   - 调用方须传入 context.Context，日志自动从链路上下文提取 trace_id；
//     无链路上下文时 trace_id 为空，不影响正常输出。
package log

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 全局 logger：本地结构化输出。远端上报由 LoggerProvider（otel.Setup 初始化）负责。
var local *zap.Logger

// provider 保存由 otel.Setup 注入的 LoggerProvider，用于把日志上报到 OTLP。
// 未注入时为 nil，emit 时安全跳过（仅本地输出）。
var (
	providerMu sync.RWMutex
	provider   otellog.LoggerProvider
)

// SetProvider 由 otel.Setup 在初始化 LoggerProvider 后调用，使日志具备远端上报能力。
func SetProvider(lp otellog.LoggerProvider) {
	providerMu.Lock()
	defer providerMu.Unlock()
	provider = lp
}

func init() {
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "ts"
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	enc := zapcore.NewJSONEncoder(cfg)
	core := zapcore.NewCore(enc, zapcore.Lock(os.Stderr), zapcore.DebugLevel)
	local = zap.New(core, zap.AddCallerSkip(2), zap.AddStacktrace(zapcore.ErrorLevel))
}

// TraceID 从 context 中提取当前链路的 trace id（16 进制字符串），无链路时返回空串。
func TraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}
	return span.SpanContext().TraceID().String()
}

// SpanID 从 context 中提取当前 span id。
func SpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}
	return span.SpanContext().SpanID().String()
}

// ctxFields 返回注入到本地日志的链路字段。
func ctxFields(ctx context.Context) []zap.Field {
	tid := TraceID(ctx)
	if tid == "" {
		return nil
	}
	return []zap.Field{zap.String("trace_id", tid), zap.String("span_id", SpanID(ctx))}
}

// Info 记录一条 info 级别日志，ctx 用于注入 trace_id。
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	local.Info(msg, append(ctxFields(ctx), fields...)...)
	emit(ctx, otellog.SeverityInfo, msg, fields)
}

// Warn 记录一条 warn 级别日志，ctx 用于注入 trace_id。
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	local.Warn(msg, append(ctxFields(ctx), fields...)...)
	emit(ctx, otellog.SeverityWarn, msg, fields)
}

// Error 记录一条 error 级别日志，ctx 用于注入 trace_id。
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	local.Error(msg, append(ctxFields(ctx), fields...)...)
	emit(ctx, otellog.SeverityError, msg, fields)
}

// Fatal 记录一条 error 级别日志后退出进程（仅用于启动期致命错误）。
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	local.Fatal(msg, append(ctxFields(ctx), fields...)...)
}

// emit 把日志作为 OTLP LogRecord 上报（若 LoggerProvider 已初始化）。
func emit(ctx context.Context, severity otellog.Severity, msg string, fields []zap.Field) {
	// 未初始化 LoggerProvider 时仅本地输出，不 panic。
	providerMu.RLock()
	lp := provider
	providerMu.RUnlock()
	if lp == nil {
		return
	}
	lg := lp.Logger("dextea-customer-api")

	attrs := make([]attribute.KeyValue, 0, len(fields))
	for _, f := range fields {
		enc := zapcore.NewMapObjectEncoder()
		f.AddTo(enc)
		for k, v := range enc.Fields {
			attrs = append(attrs, attribute.String(k, fmt.Sprintf("%v", v)))
		}
	}

	lr := otellog.Record{}
	lr.SetTimestamp(time.Now())
	lr.SetSeverity(severity)
	lr.SetBody(attribute.StringValue(msg))
	lr.AddAttributes(attrs...)
	lg.Emit(ctx, lr)
}
