package area

// NewModule 组装 area 模块的全部依赖并返回 HTTP Handler。
func NewModule(amapAPIKey string) *Handler {
	return NewHandler(NewService(amapAPIKey))
}
