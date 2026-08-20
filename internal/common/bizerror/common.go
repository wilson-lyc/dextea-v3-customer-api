package bizerror

import "net/http"

// 通用错误中心：所有模块共享的系统 / 鉴权 / 参数类错误。
// 业务代码只引用这些模板，禁止在 handler 中手写文案。
//
// 错误码分段规范（详见 docs/异常处理机制重构方案.md）：
//   1xxxxx 系统 | 2xxxxx 业务 | 3xxxxx 下游依赖 | 4xxxxx 参数/校验 | 5xxxxx 限流/幂等
// 迁移期保留历史码值（如 50000、40100、40001、40301）以兼容前端契约，新增错误使用新分段。

// 系统错误（历史码 50000 保留）
var (
	ErrInternal = &BizError{Code: 50000, Message: "内部错误", Kind: KindSystem}
)

// 成功
var (
	ErrOK = &BizError{Code: 0, Message: "ok", Kind: KindBusiness}
)

// 鉴权 / 参数（历史码保留，新分段以 4xxxxx 表达）
var (
	ErrUnauthorized = &BizError{Code: 40100, Message: "未登录或登录已过期", Kind: KindValidation, httpStatus: http.StatusUnauthorized}
	ErrForbidden    = &BizError{Code: 40301, Message: "无权限访问", Kind: KindValidation, httpStatus: http.StatusForbidden}
	ErrBadRequest   = &BizError{Code: 40001, Message: "参数不合法", Kind: KindValidation}
	ErrParamInvalid = &BizError{Code: 40000, Message: "参数错误", Kind: KindValidation}
)

// 限流 / 熔断（新分段 5xxxxx）
var (
	ErrTooManyRequests = &BizError{Code: 50001, Message: "请求过于频繁，请稍后重试", Kind: KindFlowControl}
	ErrCircuitBreaker  = &BizError{Code: 50002, Message: "服务繁忙，请稍后重试", Kind: KindFlowControl}
	ErrDuplicateSubmit = &BizError{Code: 50003, Message: "请勿重复提交", Kind: KindFlowControl}
)
