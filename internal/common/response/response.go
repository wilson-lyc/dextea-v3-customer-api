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

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
)

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var (
	CodeOK         = bizerror.CodeOK.Code
	CodeBadRequest = bizerror.CodeBadRequest.Code
	CodeNotFound   = bizerror.CodeNotFound.Code
	CodeDBDisabled = bizerror.CodeDBDisabled.Code
	CodeInternal   = bizerror.CodeInternal.Code
)

func Success(c *gin.Context, httpStatus int, data interface{}) {
	c.JSON(httpStatus, APIResponse{
		Code:    CodeOK,
		Message: "ok",
		Data:    data,
	})
}

func OK(c *gin.Context, data interface{}) {
	Success(c, http.StatusOK, data)
}

func Created(c *gin.Context, data interface{}) {
	Success(c, http.StatusCreated, data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Fail(c *gin.Context, httpStatus, bizCode int, message string) {
	c.JSON(httpStatus, APIResponse{
		Code:    bizCode,
		Message: message,
		Data:    nil,
	})
}

func FailError(c *gin.Context, httpStatus, bizCode int, err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	Fail(c, httpStatus, bizCode, msg)
}

func FailBiz(c *gin.Context, err *bizerror.BizError) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    err.Code(),
		Message: err.Message(),
		Data:    nil,
	})
}

func FailBadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, CodeBadRequest, message)
}

func FailNotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, CodeNotFound, message)
}

func FailServiceUnavailable(c *gin.Context, message string) {
	Fail(c, http.StatusServiceUnavailable, CodeDBDisabled, message)
}

func FailInternal(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, CodeInternal, message)
}

func WriteError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if b, ok := bizerror.As(err); ok {
		FailBiz(c, b)
		return
	}
	status, code, msg := classifySystemError(err)
	Fail(c, status, code, msg)
}

func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if _, ok := bizerror.As(err); !ok {
		log.Printf("[ERROR] handler error: %+v", err)
	}
	WriteError(c, err)
}

func classifySystemError(err error) (httpStatus, bizCode int, message string) {
	var mysqlErr *mysql.MySQLError
	var netErr net.Error

	switch {
	case errors.Is(err, driver.ErrBadConn),
		errors.Is(err, sql.ErrConnDone),
		errors.Is(err, sql.ErrTxDone),
		errors.As(err, &mysqlErr):
		return http.StatusInternalServerError, CodeInternal, "数据库服务异常，请稍后重试"

	case errors.As(err, &netErr),
		errors.Is(err, syscall.ECONNREFUSED),
		errors.Is(err, syscall.ETIMEDOUT):
		return http.StatusInternalServerError, CodeInternal, "网络连接异常，请稍后重试"

	default:
		return http.StatusInternalServerError, CodeInternal, "服务器内部错误，请稍后重试"
	}
}
