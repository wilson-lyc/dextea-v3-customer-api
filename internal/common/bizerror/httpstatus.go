package bizerror

import "net/http"

// HTTPStatus 根据错误类别推导默认 HTTP 状态码。
// 约定：校验/业务类使用对应客户端错误码；下游/系统/限流类默认 200 + 业务码，
// 由前端按 code 判断（保持网关对下游透传的契约），同时限流暴露 429 便于边缘缓存识别。
func (e *BizError) HTTPStatus() int {
	if e.httpStatus != 0 {
		return e.httpStatus
	}
	switch e.Kind {
	case KindValidation:
		return http.StatusBadRequest
	case KindFlowControl:
		return http.StatusTooManyRequests
	case KindDownstream:
		return http.StatusOK
	case KindSystem:
		return http.StatusOK
	case KindBusiness:
		return http.StatusOK
	default:
		return http.StatusOK
	}
}
