package order

import "github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"

var (
	ErrOrderServiceNotConfigured = &bizerror.BizError{Code: 42001, Message: "订单服务未配置"}
	ErrOrderServiceUnavailable   = &bizerror.BizError{Code: 42002, Message: "订单服务暂不可用"}
	ErrOrderServiceError         = &bizerror.BizError{Code: 42003, Message: "订单服务处理失败"}
)
