package bizerror

// 基础设施错误（系统类）。
var (
	// ErrMysqlDisabled mysql(数据库)未启用时返回的统一错误。
	ErrMysqlDisabled = &BizError{Code: 50300, Message: "数据库未启用", Kind: KindSystem}
	// ErrRedisDisabled redis 未启用时返回的统一错误。
	ErrRedisDisabled = &BizError{Code: 50301, Message: "Redis 未启用", Kind: KindSystem}
)
