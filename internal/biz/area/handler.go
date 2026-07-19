package area

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
	area := r.Group("/api/v1/area")
	area.GET("/regeo", h.ReverseGeocode)
	area.GET("/cities", h.GetCities)
}

// ReverseGeocode 逆地址编码接口
// GET /api/v1/area/regeo?longitude=xxx&latitude=xxx
func (h *Handler) ReverseGeocode(c *gin.Context) {
	var req ReverseGeocodeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("[WARN] area regeo invalid request params: %+v", err)
		response.FailBiz(c, bizerror.New(bizerror.CodeValidationFail))
		return
	}

	result, err := h.svc.ReverseGeocode(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, result)
}

// GetCities 获取城市列表
// GET /api/v1/area/cities
func (h *Handler) GetCities(c *gin.Context) {
	result, err := h.svc.GetCities(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, result)
}
