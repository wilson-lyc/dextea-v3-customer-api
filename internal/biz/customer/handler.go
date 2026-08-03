package customer

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/common/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	v1.POST("/customers/login", h.Login)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[WARN] 客户登录请求参数无效: %+v", err)
		response.ErrorBiz(c, bizerror.New(&bizerror.BizError{Code: 40001, Message: "请求参数不合法"}))
		return
	}

	if !req.Platform.Valid() {
		response.ErrorBiz(c, bizerror.New(ErrPlatformInvalid))
		return
	}

	result, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	response.OK(c, result)
}
