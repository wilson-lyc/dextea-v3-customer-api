package response

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
)

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    bizerror.ErrOK.Code,
		Message: "ok",
		Data:    data,
	})
}

func Error(c *gin.Context, httpStatus, bizCode int, message string) {
	c.JSON(httpStatus, APIResponse{
		Code:    bizCode,
		Message: message,
		Data:    nil,
	})
}

func ErrorBiz(c *gin.Context, err *bizerror.BizError) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    err.Code,
		Message: err.Message,
		Data:    nil,
	})
}

func ErrorOf(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if b, ok := bizerror.As(err); ok {
		ErrorBiz(c, b)
		return
	}
	log.Printf("[ERROR] handler error: %+v", err)
	Error(c, http.StatusInternalServerError, bizerror.ErrInternal.Code, "服务异常，请稍后重试")
}
