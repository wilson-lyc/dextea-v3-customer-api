package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// TraceIDHeader 是每个请求响应中注入的 trace id 头名称，
// 便于前端/网关在出现问题时携带该值定位链路。
const TraceIDHeader = "X-Trace-Id"

// Trace 是一个 gin 中间件，为每个请求在响应头注入当前链路的 trace id。
//
// 说明：
//   - 入站请求的链路 span 由 otelhttp 在 router 层创建（见 router.Setup），
//     此处只需从请求 context 中取出 span 并写入 trace id；
//   - 即便请求未被 OTel 插桩（例如未配置采集端），也仍会从 context 取得
//     一个无操作的 span，trace id 为空字符串，不影响正常响应。
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		span := trace.SpanFromContext(c.Request.Context())
		if span == nil {
			return
		}
		traceID := span.SpanContext().TraceID().String()
		if traceID == "" || traceID == "00000000000000000000000000000000" {
			return
		}
		c.Header(TraceIDHeader, traceID)
	}
}

// Tracer 返回本项目全局 tracer，供业务代码手动打点。
func Tracer() trace.Tracer {
	return otel.Tracer("dextea-customer-api")
}
