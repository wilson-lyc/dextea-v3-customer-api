package bizerror

// Kind 表示错误的宏观类别，决定 HTTP 映射、是否可重试、是否需要告警。
// 对照微服务错误分类（用户错误 vs 服务端错误）与阿里错误码分层。
type Kind int

const (
	KindSystem     Kind = 1 // 系统错误：基础设施、未知 panic
	KindBusiness   Kind = 2 // 业务错误：领域规则不满足
	KindDownstream Kind = 3 // 下游依赖错误：调用订单中台等失败
	KindValidation Kind = 4 // 参数 / 校验错误
	KindFlowControl Kind = 5 // 限流 / 幂等 / 熔断
)
