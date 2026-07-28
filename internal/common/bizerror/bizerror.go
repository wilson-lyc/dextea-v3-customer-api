package bizerror

import "fmt"

type BizErrorCode struct {
	Code    int
	Message string
}

var (
	CodeOK         = BizErrorCode{Code: 0, Message: "ok"}
	CodeBadRequest = BizErrorCode{Code: 40000, Message: "参数错误"}
	CodeNotFound   = BizErrorCode{Code: 40400, Message: "资源不存在"}
	CodeDBDisabled     = BizErrorCode{Code: 50300, Message: "数据库未启用"}
	CodeInternal       = BizErrorCode{Code: 50000, Message: "内部错误"}
	CodeValidationFail = BizErrorCode{Code: 40001, Message: "请求参数不合法"}
	CodeUnauthorized   = BizErrorCode{Code: 40100, Message: "未登录或登录已过期"}
)

type BizError struct {
	code    BizErrorCode
	message string
}

func New(code BizErrorCode, msg ...string) *BizError {
	e := &BizError{code: code}
	if len(msg) > 0 {
		e.message = msg[0]
	}
	return e
}

func (e *BizError) Error() string {
	return fmt.Sprintf("[%d] %s", e.code.Code, e.Message())
}

func (e *BizError) Code() int {
	return e.code.Code
}

func (e *BizError) BizCode() BizErrorCode {
	return e.code
}

func (e *BizError) Message() string {
	if e.message != "" {
		return e.message
	}
	return e.code.Message
}

func As(err error) (*BizError, bool) {
	if err == nil {
		return nil, false
	}
	if b, ok := err.(*BizError); ok {
		return b, true
	}
	return nil, false
}
