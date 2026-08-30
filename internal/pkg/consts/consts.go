package consts

const (
	CustomerIDHeader = "X-Customer-Id"
	APIPrefixV1 = "/api/v1"

	// DownstreamAPIPrefix 是下游订单服务的固定路径前缀（如 /api/v1/customer/orders）。
	// 该前缀不在 env 的 ORDER_SERVICE_BASE_URL 中配置，统一在此固定，
	// 配置只需提供 host:port（如 example.com:3000）；转发时剥离本地订单路由前缀
	// （APIPrefixV1 + /orders）后，再拼接该前缀得到下游真实地址。
	DownstreamAPIPrefix = "/api/v1/customer/orders"
)
