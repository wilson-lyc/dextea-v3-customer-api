package order

import "github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"

var (
	CodeOrderServiceNotConfigured = bizerror.BizErrorCode{Code: 42001, Message: "订单服务未配置"}
	CodeOrderServiceUnavailable   = bizerror.BizErrorCode{Code: 42002, Message: "订单服务暂不可用"}
	CodeOrderServiceError         = bizerror.BizErrorCode{Code: 42003, Message: "订单服务处理失败"}
)
