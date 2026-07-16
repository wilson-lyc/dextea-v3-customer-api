package customer

import (
	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/response"
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
		response.FailBiz(c, bizerror.New(bizerror.CodeBadRequest, err.Error()))
		return
	}

	if !req.Platform.Valid() {
		response.FailBiz(c, bizerror.New(CodePlatformInvalid))
		return
	}

	result, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, result)
}
