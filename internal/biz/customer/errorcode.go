package customer

import "github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"

// 客服模块错误码。所有错误统一在 bizerror 错误中心风格下定义，业务代码只引用模板。
var (
	ErrPlatformInvalid     = &bizerror.BizError{Code: 41001, Message: "不支持的登录平台", Kind: bizerror.KindValidation}
	ErrPlatformNotSupported = &bizerror.BizError{Code: 41002, Message: "该登录平台暂未开放", Kind: bizerror.KindValidation}
	ErrAlipayNotConfigured = &bizerror.BizError{Code: 41003, Message: "支付宝登录未配置", Kind: bizerror.KindDownstream}
	ErrAlipayAuthFailed    = &bizerror.BizError{Code: 41004, Message: "支付宝授权失败", Kind: bizerror.KindDownstream}
	// ErrCustomerDataEmpty 客服数据为空（下游未返回内容）。码值从 42001 调整为 44001，
	// 避免与订单模块 42001 冲突，并归入业务类 4xxxxx 段。
	ErrCustomerDataEmpty = &bizerror.BizError{Code: 44001, Message: "客服数据为空", Kind: bizerror.KindDownstream}
)
