package customer

import "github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"

var (
	ErrPlatformInvalid     = &bizerror.BizError{Code: 41001, Message: "不支持的登录平台"}
	ErrPlatformNotSupported = &bizerror.BizError{Code: 41002, Message: "该登录平台暂未开放"}
	ErrAlipayNotConfigured = &bizerror.BizError{Code: 41003, Message: "支付宝登录未配置"}
	ErrAlipayAuthFailed    = &bizerror.BizError{Code: 41004, Message: "支付宝授权失败"}
)
