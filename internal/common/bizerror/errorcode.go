package bizerror

var (
	// ErrMysqlDisabled mysql(数据库)未启用时返回的统一错误。
	ErrMysqlDisabled = &BizError{Code: 50300, Message: "数据库未启用"}
	// ErrRedisDisabled redis 未启用时返回的统一错误。
	ErrRedisDisabled = &BizError{Code: 50301, Message: "Redis 未启用"}
)
