// Package bizerror 定义业务错误码（BizErrorCode）与业务异常（BizError）。
//
// BizErrorCode 是「错误码 + 默认消息」的配对；BizError 是可在业务层抛出/返回的
// 异常，它引用一个 BizErrorCode，并允许用一个新 message 覆盖其默认消息。
package bizerror

import "fmt"

// BizErrorCode 错误码与默认消息的配对。
type BizErrorCode struct {
	Code    int
	Message string
}

// 预定义的业务错误码。新增错误码时在此登记即可。
var (
	CodeOK         = BizErrorCode{Code: 0, Message: "ok"}
	CodeBadRequest = BizErrorCode{Code: 40000, Message: "参数错误"}
	CodeNotFound   = BizErrorCode{Code: 40400, Message: "资源不存在"}
	CodeDBDisabled = BizErrorCode{Code: 50300, Message: "数据库未启用"}
	CodeInternal   = BizErrorCode{Code: 50000, Message: "内部错误"}
)

// BizError 业务异常，可携带一个 BizErrorCode 与可选的覆盖消息。
type BizError struct {
	code    BizErrorCode
	message string // 覆盖消息；为空时回退到 code.Message
}

// New 基于 BizErrorCode 构造业务异常。
// 可选的 msg 用于覆盖该错误码的默认消息；不传则使用默认消息。
func New(code BizErrorCode, msg ...string) *BizError {
	e := &BizError{code: code}
	if len(msg) > 0 {
		e.message = msg[0]
	}
	return e
}

// Error 实现 error 接口，便于作为异常抛出或返回。
func (e *BizError) Error() string {
	return fmt.Sprintf("[%d] %s", e.code.Code, e.Message())
}

// Code 返回最终错误码的整数值（始终等于所引用 BizErrorCode 的 Code）。
func (e *BizError) Code() int {
	return e.code.Code
}

// BizCode 返回所引用的 BizErrorCode 对象。
func (e *BizError) BizCode() BizErrorCode {
	return e.code
}

// Message 返回最终消息：传入了覆盖消息则用覆盖，否则用默认消息。
func (e *BizError) Message() string {
	if e.message != "" {
		return e.message
	}
	return e.code.Message
}

// As 从 err 中提取 *BizError，便于按错误码分支处理。
// 非 *BizError 或 nil 时返回 (nil, false)。
func As(err error) (*BizError, bool) {
	if err == nil {
		return nil, false
	}
	if b, ok := err.(*BizError); ok {
		return b, true
	}
	return nil, false
}
