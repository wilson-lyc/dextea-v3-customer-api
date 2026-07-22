package order

import (
	"io"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/common/response"
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

func (h *Handler) handle(c *gin.Context, fn func(context.Context, []byte) (*ForwardResult, error)) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[WARN] order forward read request body failed: %+v", err)
		response.FailBadRequest(c, "读取请求体失败")
		return
	}

	result, err := fn(c.Request.Context(), body)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Data(result.StatusCode, result.ContentType, result.Body)
}
