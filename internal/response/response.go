// Package response 定义统一的 HTTP 接口响应结构，供各业务模块的 handler 复用，
// 避免响应形态（成功/失败字段、业务码）散落在各处。
package response

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"log"
	"net"
	"net/http"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"

	"github.com/dextea-v3/dextea-customer/api/internal/bizerror"
)

// APIResponse 统一接口响应结构。
//
// 所有接口都应通过本包的辅助函数返回该结构：
//   - code：业务状态码，0 表示成功（与 HTTP 状态码解耦，便于前端按码分支）；
//   - message：可读提示文案，成功时通常为 "ok"；
//   - data：业务数据；无数据时传 nil。
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// 业务状态码常量（单一来源为 bizerror 包）。与 HTTP 状态码区分，便于前端按码分支。
var (
	CodeOK         = bizerror.CodeOK.Code
	CodeBadRequest = bizerror.CodeBadRequest.Code
	CodeNotFound   = bizerror.CodeNotFound.Code
	CodeDBDisabled = bizerror.CodeDBDisabled.Code
	CodeInternal   = bizerror.CodeInternal.Code
)

// Success 以指定 HTTP 状态码返回成功响应。
func Success(c *gin.Context, httpStatus int, data interface{}) {
	c.JSON(httpStatus, APIResponse{
		Code:    CodeOK,
		Message: "ok",
		Data:    data,
	})
}

// OK 等价于 Success(c, 200, data)。
func OK(c *gin.Context, data interface{}) {
	Success(c, http.StatusOK, data)
}

// Created 等价于 Success(c, 201, data)。
func Created(c *gin.Context, data interface{}) {
	Success(c, http.StatusCreated, data)
}

// NoContent 返回 204 无内容。
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Fail 返回错误响应。bizCode 为业务错误码，message 为提示文案。
func Fail(c *gin.Context, httpStatus, bizCode int, message string) {
	c.JSON(httpStatus, APIResponse{
		Code:    bizCode,
		Message: message,
		Data:    nil,
	})
}

// FailError 以 err.Error() 作为提示返回错误响应。
// err 为 nil 时提示为空字符串。
func FailError(c *gin.Context, httpStatus, bizCode int, err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	Fail(c, httpStatus, bizCode, msg)
}

// FailBiz 以 *bizerror.BizError 返回错误响应，自动取其中的错误码与最终消息。
// 这是推荐用法：业务层用 bizerror.New(...) 抛出异常，handler 直接透传。
func FailBiz(c *gin.Context, httpStatus int, err *bizerror.BizError) {
	c.JSON(httpStatus, APIResponse{
		Code:    err.Code(),
		Message: err.Message(),
		Data:    nil,
	})
}

// FailBadRequest 返回 400 错误响应。
func FailBadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, CodeBadRequest, message)
}

// FailNotFound 返回 404 错误响应。
func FailNotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, CodeNotFound, message)
}

// FailServiceUnavailable 返回 503 错误响应（如数据库未启用）。
func FailServiceUnavailable(c *gin.Context, message string) {
	Fail(c, http.StatusServiceUnavailable, CodeDBDisabled, message)
}

// FailInternal 返回 500 错误响应。
func FailInternal(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, CodeInternal, message)
}

// WriteError 将任意 error 转换为统一的错误响应，供 handler 与全局异常拦截器共用。
//
//   - *bizerror.BizError：透传其业务码与最终消息（业务异常按原样返回，不清洗）；
//   - 其余错误（数据库/网络/运行异常等）：一律清洗为对外的通用提示文案，
//     绝不把 SQL、堆栈、系统错误等给程序员看的底层细节回显给前端。
//
// 注意：本函数只负责「对外的响应内容」。原始错误（含 SQL、堆栈等）应由调用方
// 在写出响应之前记录到服务器日志，避免信息泄露。
func WriteError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if b, ok := bizerror.As(err); ok {
		status := http.StatusInternalServerError
		if b.Code() == bizerror.CodeDBDisabled.Code {
			status = http.StatusServiceUnavailable
		}
		FailBiz(c, status, b)
		return
	}
	status, code, msg := classifySystemError(err)
	Fail(c, status, code, msg)
}

// HandleError 处理并写出错误响应，同时把非业务异常（数据库/网络/运行异常）
// 记录到服务器日志，避免信息泄露到前端；业务异常（*bizerror.BizError）只透传不记录日志。
// handler 层统一调用本函数，无需各业务模块重复编写 writeError 辅助函数。
//
// 注意：本函数只负责「对外的响应内容」与「非业务异常的日志记录」。
// 若业务模块需要额外的上下文日志，可在调用前自行补充。
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if _, ok := bizerror.As(err); !ok {
		log.Printf("[ERROR] handler error: %+v", err)
	}
	WriteError(c, err)
}

// classifySystemError 将非业务异常分类为对外的通用提示。
// 数据库、网络、运行等任何非业务异常都不应把底层细节回显给前端，
// 而是给出友好的「稍后重试」类提示；具体的错误类型仅用于区分文案方向。
func classifySystemError(err error) (httpStatus, bizCode int, message string) {
	var mysqlErr *mysql.MySQLError
	var netErr net.Error

	switch {
	// 数据库异常：连接断开、事务状态、MySQL 驱动层错误等。
	case errors.Is(err, driver.ErrBadConn),
		errors.Is(err, sql.ErrConnDone),
		errors.Is(err, sql.ErrTxDone),
		errors.As(err, &mysqlErr):
		return http.StatusInternalServerError, CodeInternal, "数据库服务异常，请稍后重试"

	// 网络异常：连接被拒、超时、DNS 等网络层错误。
	case errors.As(err, &netErr),
		errors.Is(err, syscall.ECONNREFUSED),
		errors.Is(err, syscall.ETIMEDOUT):
		return http.StatusInternalServerError, CodeInternal, "网络连接异常，请稍后重试"

	// 其余运行异常：统一兜底为服务器内部错误。
	default:
		return http.StatusInternalServerError, CodeInternal, "服务器内部错误，请稍后重试"
	}
}
