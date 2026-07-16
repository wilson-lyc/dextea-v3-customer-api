package customer

import "github.com/dextea-v3/dextea-customer/api/internal/bizerror"

var (
	CodePlatformInvalid = bizerror.BizErrorCode{Code: 41001, Message: "不支持的登录平台"}
)
