package menu

import (
	"net/http"
	"strconv"

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
	g := r.Group("/api/v1/stores")
	g.GET("/:storeId/menu", h.GetStoreMenu)
}

// GetStoreMenu 获取门店菜单，门店 ID 取自路径参数。
func (h *Handler) GetStoreMenu(c *gin.Context) {
	storeID, err := strconv.ParseInt(c.Param("storeId"), 10, 64)
	if err != nil || storeID <= 0 {
		response.Error(c, http.StatusBadRequest, 40001, "门店ID不合法")
		return
	}

	result, err := h.svc.GetStoreMenu(c.Request.Context(), storeID)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	response.OK(c, result)
}
