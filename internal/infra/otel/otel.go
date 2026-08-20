// Package otel 封装 OpenTelemetry 初始化逻辑：全局文本传播器（W3C TraceContext +
// Baggage）与基于 OTLP 的 TracerProvider。初始化遵循标准 OTel 环境变量
// （OTEL_EXPORTER_OTLP_ENDPOINT 等），无需硬编码上报地址。
package otel

import (
	"context"
	"log"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	applog "github.com/dextea-v3/dextea-customer/api/internal/infra/log"
)

// Setup 初始化全局 propagator 与 TracerProvider，返回 shutdown 函数用于优雅退出。
//
// serviceName 用于设置 resource 的服务名，上报到后端的 trace 会带上该标识；
// 若未配置 OTEL_EXPORTER_OTLP_ENDPOINT，autoexport 会回退到无操作 exporter，
// 不会因缺少采集端而启动失败。
func Setup(ctx context.Context, serviceName, serviceVersion, environment string) (func(context.Context) error, error) {
	// 1. 全局文本传播器：负责跨服务边界注入/提取 traceparent 等头。
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	// 2. Resource：标识服务身份，便于在后端区分来源。
	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.DeploymentEnvironment(environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// 3. Span exporter：优先使用 autoexport（尊重 OTEL_EXPORTER_OTLP_* 环境变量），
	//    回退到 OTLP gRPC 默认端点（localhost:4317），保证任意环境都能启动。
	spanExporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		log.Printf("[otel] autoexport span exporter failed, fallback to otlptracegrpc: %v", err)
		spanExporter, err = otlptracegrpc.New(ctx)
		if err != nil {
			return nil, err
		}
	}

	// 4. TracerProvider：批量上报以降低开销。
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(spanExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	// 5. LoggerProvider：将日志通过 OTLP 上报，与 trace 在后端按 trace_id 关联。
	//    优先使用 autoexport（尊重 OTEL_LOGS_EXPORTER 环境变量），未配置时回退到
	//    无操作 exporter，不会因缺少日志采集端而启动失败。
	loggerProvider, err := newLoggerProvider(ctx, res)
	if err != nil {
		return nil, err
	}
	applog.SetProvider(loggerProvider)

	return func(ctx context.Context) error {
		var err error
		if e := tracerProvider.Shutdown(ctx); e != nil {
			err = e
		}
		if e := loggerProvider.Shutdown(ctx); e != nil && err == nil {
			err = e
		}
		return err
	}, nil
}

// newLoggerProvider 创建 LoggerProvider：优先 autoexport，回退 otlploggrpc 默认端点。
func newLoggerProvider(ctx context.Context, res *sdkresource.Resource) (*sdklog.LoggerProvider, error) {
	logExporter, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		log.Printf("[otel] autoexport log exporter failed, fallback to otlploggrpc: %v", err)
		// 回退到 OTLP gRPC 默认端点（localhost:4317）。
		logExporter, err = otlploggrpc.New(ctx)
		if err != nil {
			return nil, err
		}
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)
	return lp, nil
}
