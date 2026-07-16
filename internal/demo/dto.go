package demo

// HealthResponse 健康检查接口的响应结构。
//
// 该文件属于「接口数据提取」层：集中定义请求/响应的数据形态，
// 并承载 gin 的绑定（ShouldBind）与校验（binding tag），
// 让 handler 只关心流程编排，不掺杂数据结构细节。
type HealthResponse struct {
	Status  string `json:"status"  binding:"required"` // 服务整体状态
	Service string `json:"service"`                    // 服务名
	DB      string `json:"db"`                         // 数据库连通状态：ok / down
}
