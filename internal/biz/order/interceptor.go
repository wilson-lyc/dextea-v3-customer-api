package order

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/common/response"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/dextea-v3/dextea-customer/api/internal/jwt"
)

// CustomerIDHeader 是 Order 模块转发时，写入请求头、供下游 Java 订单服务读取的
// 固定字段名。所有 Order 模块的转发请求都必须通过该拦截器统一写入此头。
const CustomerIDHeader = "X-Customer-Id"

// customIDContextKey 是写入 gin.Context 的 key，便于 handler 直接取用已解析的顾客 id。
const customIDContextKey = "customID"

// ForwardInterceptor 是 Order 模块统一的请求拦截中间件。
//
// 对所有进入 Order 模块的转发请求，强制按如下顺序处理：
//  1. 从原始请求的 Authorization: Bearer <token> 中提取 Token；
//  2. 使用 JWT 解析出其中的 Customer ID（即 uid claim）；
//  3. 将 Customer ID 写入固定请求头 X-Custom-Id，供后续转发链路向下游透传。
//
// 任何一步失败（缺头、格式非法、签名/过期校验不通过）均直接返回 401，
// 不会进入后续 handler，确保未经过身份解析的请求绝不会被转发。
func ForwardInterceptor(cfg *config.Config) gin.HandlerFunc {
	secret := cfg.JWTSecret

	return func(c *gin.Context) {
		// 1. 提取 Token：仅接受 Bearer 方案。
		const bearerPrefix = "Bearer "
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, bearerPrefix) {
			response.Fail(c, http.StatusUnauthorized, bizerror.CodeUnauthorized.Code, bizerror.CodeUnauthorized.Message)
			c.Abort()
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(auth, bearerPrefix))
		if token == "" {
			response.Fail(c, http.StatusUnauthorized, bizerror.CodeUnauthorized.Code, bizerror.CodeUnauthorized.Message)
			c.Abort()
			return
		}

		// 2. 解析 Token，得到 Customer ID（uid）。
		claims, err := jwt.Parse(secret, token)
		if err != nil {
			// 不把底层错误细节回显给前端，统一为未登录提示。
			response.Fail(c, http.StatusUnauthorized, bizerror.CodeUnauthorized.Code, bizerror.CodeUnauthorized.Message)
			c.Abort()
			return
		}

		// 3. 将 Customer ID 统一写入固定请求头 X-Customer-Id。
		customerID := strconv.FormatInt(claims.UID, 10)
		c.Request.Header.Set(CustomerIDHeader, customerID)
		c.Set(customIDContextKey, claims.UID)

		c.Next()
	}
}
