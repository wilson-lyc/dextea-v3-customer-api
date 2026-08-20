package order

import "github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"

// 订单模块错误码（下游订单中台相关）。所有码值在 bizerror 错误中心统一定义，
// 业务代码只引用模板，禁止散落手写文案。
var (
	// ErrOrderServiceNotConfigured 订单服务未配置（Nacos/静态地址皆不可用）。
	ErrOrderServiceNotConfigured = &bizerror.BizError{Code: 42001, Message: "订单服务未配置", Kind: bizerror.KindDownstream}
	// ErrOrderServiceUnavailable 订单服务暂不可用（网络/超时等瞬时失败）。
	ErrOrderServiceUnavailable = &bizerror.BizError{Code: 42002, Message: "订单服务暂不可用", Kind: bizerror.KindDownstream}
	// ErrOrderServiceError 订单服务处理失败（下游返回非预期响应）。
	ErrOrderServiceError = &bizerror.BizError{Code: 42003, Message: "订单服务处理失败", Kind: bizerror.KindDownstream}
)
