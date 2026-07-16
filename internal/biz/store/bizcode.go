package store

import "github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"

var (
	CodeRedisDisabled = bizerror.BizErrorCode{Code: 42001, Message: "Redis 未启用"}
	CodeDBDisabled    = bizerror.BizErrorCode{Code: 42002, Message: "数据库未启用"}
)
