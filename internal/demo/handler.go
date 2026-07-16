package demo

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/response"
)

// Handler 属于「接口定义」层，负责把业务能力暴露为 HTTP 接口，
// 并完成请求上下文的传递与响应写出。响应统一走 response 包。
type Handler struct {
	svc *Service
}

// NewHandler 创建 Handler 实例。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 将本模块的路由注册到 gin 引擎。
// 每个业务模块自行声明自己的接口，避免路由散落在全局 router 中。
func (h *Handler) Register(r *gin.Engine) {
	r.GET("/health", h.Health)
	r.POST("/products", h.CreateProduct)
}

// Health 健康检查接口。
//
//	@Summary	健康检查
//	@Tags		demo
//	@Produce	json
//	@Success	200	{object}	HealthResponse
//	@Router		/health [get]
func (h *Handler) Health(c *gin.Context) {
	resp, err := h.svc.CheckHealth(c.Request.Context())
	if err != nil {
		response.FailInternal(c, err.Error())
		return
	}
	response.OK(c, resp)
}

// writeError 把 service/repo 层错误转换为统一的错误响应。
// 直接复用 response.WriteError：业务异常透传，系统/数据库/网络异常被清洗为通用提示。
// 对于非业务异常，先记录原始错误到服务器日志，再写出清洗后的对外提示。
func writeError(c *gin.Context, err error) {
	if _, ok := bizerror.As(err); !ok {
		log.Printf("[ERROR] handler error: %+v", err)
	}
	response.WriteError(c, err)
}

// CreateProduct 创建商品（demo 模块保留的唯一业务接口，用于联调测试）。
func (h *Handler) CreateProduct(c *gin.Context) {
	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 以 BizError 抛出业务异常，并用校验错误详情覆盖默认消息。
		response.FailBiz(c, http.StatusBadRequest,
			bizerror.New(bizerror.CodeBadRequest, err.Error()))
		return
	}

	p := &Product{Name: req.Name, Price: req.Price}
	created, err := h.svc.CreateProduct(c.Request.Context(), p)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Created(c, created)
}
