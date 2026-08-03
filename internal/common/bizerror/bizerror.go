package bizerror

import "fmt"

type BizError struct {
	Code    int
	Message string
}

var (
	ErrOK       = &BizError{Code: 0, Message: "ok"}
	ErrInternal = &BizError{Code: 50000, Message: "内部错误"}
)

func New(code *BizError, msg ...string) *BizError {
	e := &BizError{Code: code.Code, Message: code.Message}
	if len(msg) > 0 {
		e.Message = msg[0]
	}
	return e
}

func (e *BizError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
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
