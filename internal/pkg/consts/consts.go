package consts

const (
	CustomerIDHeader = "X-Customer-Id"
	APIPrefixV1 = "/api/v1"

	// DownstreamAPIPrefix 是下游订单服务的固定路径前缀（如 /api/v1）。
	// 该前缀不在 env 的 ORDER_SERVICE_BASE_URL 中配置，统一在此固定，
	// 配置只需提供 host:port（如 example.com:3000）。
	DownstreamAPIPrefix = "/api/v1"
)
