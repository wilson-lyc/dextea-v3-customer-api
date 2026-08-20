package bizerror

import (
	"errors"
	"fmt"
)

// BizError 表示业务错误，统一承载业务码、文案、分类、根因与排障上下文。
//
// 设计要点（对标 Uber Go Style Guide / 阿里错误码规约）：
//   - 实现标准 error 接口并支持 Unwrap，可被 errors.Is / errors.As 识别，
//     即使被 fmt.Errorf("xxx: %w", err) 包装也能还原业务码。
//   - Kind 决定 HTTP 映射、是否可重试、是否需要告警，由错误中心统一定义，
//     业务代码只引用已定义模板，禁止散落手写文案。
type BizError struct {
	Code    int                    // 业务/系统错误码（0 表示成功）
	Message string                 // 对外提示文案
	Kind    Kind                   // 宏观错误类别
	Cause   error                  // 根因（可空），保留链路
	Fields  map[string]interface{} // 排障上下文，如 orderId、customerId（不回显给客户端）
	httpStatus int                 // 可选：显式指定 HTTP 状态码（非 0 时优先于 Kind 推导）
}

// New 兼容旧签名：基于已定义错误模板创建实例，可选覆盖对外文案。
// 保留此签名以兼容存量调用 bizerror.New(ErrXxx, "文案")。
func New(code *BizError, msg ...string) *BizError {
	e := &BizError{
		Code:       code.Code,
		Message:    code.Message,
		Kind:       code.Kind,
		httpStatus: code.httpStatus,
	}
	if len(msg) > 0 {
		e.Message = msg[0]
	}
	return e
}

// NewWith 新风格构造：基于模板创建实例并附加根因 / 上下文。
func NewWith(code *BizError, opts ...Option) *BizError {
	e := &BizError{
		Code:       code.Code,
		Message:    code.Message,
		Kind:       code.Kind,
		httpStatus: code.httpStatus,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Option 用于构造 BizError 时附加信息。
type Option func(*BizError)

// WithMessage 覆盖对外提示文案。
func WithMessage(msg string) Option {
	return func(e *BizError) { e.Message = msg }
}

// WithCause 绑定根因，保留错误链路（供 errors.Is/As 与日志使用）。
func WithCause(cause error) Option {
	return func(e *BizError) { e.Cause = cause }
}

// WithFields 附加排障上下文。
func WithFields(fields map[string]interface{}) Option {
	return func(e *BizError) {
		if e.Fields == nil {
			e.Fields = make(map[string]interface{}, len(fields))
		}
		for k, v := range fields {
			e.Fields[k] = v
		}
	}
}

// WithField 附加单个排障上下文字段。
func WithField(key string, value interface{}) Option {
	return func(e *BizError) {
		if e.Fields == nil {
			e.Fields = make(map[string]interface{}, 1)
		}
		e.Fields[key] = value
	}
}

// Error 实现 error 接口。
func (e *BizError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d]%s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d]%s", e.Code, e.Message)
}

// Unwrap 返回根因，使 errors.Is/As 可穿透包装。
func (e *BizError) Unwrap() error {
	return e.Cause
}

// Retryable 表示该错误是否可重试（下游瞬时失败 / 限流可退避重试）。
func (e *BizError) Retryable() bool {
	return e.Kind == KindDownstream || e.Kind == KindFlowControl
}

// Alertable 表示该错误是否需要告警（系统类错误需立即通知）。
func (e *BizError) Alertable() bool {
	return e.Kind == KindSystem
}

// As 从 error 中提取 *BizError。兼容直接 *BizError 与经 fmt.Errorf("%w") 包装的情形。
func As(err error) (*BizError, bool) {
	if err == nil {
		return nil, false
	}
	var b *BizError
	if errors.As(err, &b) {
		return b, true
	}
	return nil, false
}
