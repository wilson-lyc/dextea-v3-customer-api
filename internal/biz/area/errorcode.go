package area

import "github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"

var (
	ErrAmapUnavailable = &bizerror.BizError{Code: 43001, Message: "高德地图服务不可用", Kind: bizerror.KindDownstream}
	ErrAmapError       = &bizerror.BizError{Code: 43002, Message: "高德地图服务返回错误", Kind: bizerror.KindDownstream}
	ErrInvalidLocation = &bizerror.BizError{Code: 43003, Message: "经纬度参数无效", Kind: bizerror.KindValidation}
)
