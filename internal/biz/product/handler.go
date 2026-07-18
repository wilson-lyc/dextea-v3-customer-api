package product

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
	g := r.Group("/api/v1/products")
	g.GET("/detail", h.GetDetail)
}

func (h *Handler) GetDetail(c *gin.Context) {
	var req GetProductDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("[WARN] product detail invalid request params: %+v", err)
		response.FailBiz(c, bizerror.New(bizerror.CodeValidationFail))
		return
	}

	result, err := h.svc.GetDetail(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, result)
}
