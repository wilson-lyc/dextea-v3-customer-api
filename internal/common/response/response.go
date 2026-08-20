package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	applog "github.com/dextea-v3/dextea-customer/api/internal/infra/log"
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
		Message: bizerror.ErrOK.Message,
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

// ErrorOf 统一处理任意 error：若为 *BizError（含被 %w 包装）按错误分类映射 HTTP 状态与业务码；
// 否则按未知系统异常兜底（文案收敛，不泄露内部细节；根因仅在日志出现）。
func ErrorOf(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if b, ok := bizerror.As(err); ok {
		// 系统类错误需告警并记录根因/上下文；其余记录 warn 便于排障。
		if b.Alertable() {
			applog.Error(c.Request.Context(), "biz error(alert)",
				zap.Int("code", b.Code), zap.String("message", b.Message),
				zap.Any("fields", b.Fields), zap.Error(b.Cause))
		} else {
			applog.Warn(c.Request.Context(), "biz error",
				zap.Int("code", b.Code), zap.String("message", b.Message),
				zap.Any("fields", b.Fields))
		}
		Error(c, b.HTTPStatus(), b.Code, b.Message)
		return
	}
	// 未知错误：兜底 + 告警 + 仅日志保留根因。
	applog.Error(c.Request.Context(), "unhandled error", zap.Error(err))
	Error(c, http.StatusInternalServerError, bizerror.ErrInternal.Code, "服务异常，请稍后重试")
}
