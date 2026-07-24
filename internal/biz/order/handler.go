package order

import (
	"context"
	"io"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/common/response"
	"github.com/dextea-v3/dextea-customer/api/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
	g := r.Group("/api/v1/orders")
	g.POST("", func(c *gin.Context) { h.handle(c, h.svc.Create) })
	g.POST("/pre-build", func(c *gin.Context) { h.handle(c, h.svc.PreBuild) })
}

func (h *Handler) handle(c *gin.Context, fn func(context.Context, int64, []byte) (*ForwardResult, error)) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[WARN] order forward read request body failed: %+v", err)
		response.FailBadRequest(c, "读取请求体失败")
		return
	}

	// 取出鉴权中间件写入的已认证顾客 id（来自 token，不可被请求体伪造）。
	customerID, err := strconv.ParseInt(c.Request.Header.Get(middleware.CustomerIDHeader), 10, 64)
	if err != nil {
		response.Fail(c, 401, bizerror.CodeUnauthorized.Code, bizerror.CodeUnauthorized.Message)
		return
	}

	result, err := fn(c.Request.Context(), customerID, body)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Data(result.StatusCode, result.ContentType, result.Body)
}
