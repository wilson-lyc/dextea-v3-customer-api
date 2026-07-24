package middleware

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

// customerIDHeader 是注入到请求头、供后续业务读取的 customer 标识头。
const customerIDHeader = "X-Customer-Id"

// customerIDContextKey 是写入 gin context 的 key，便于 handler 直接取用。
const customerIDContextKey = "customerID"

// Auth 返回 gin 鉴权中间件。
//
// 命中 whitelist 的路径直接放行（如登录接口）；其余请求须携带
// `Authorization: Bearer <token>`，校验通过后将 token 中的 customerid（uid claim）
// 以字符串形式写入请求头 customerIDHeader 再放行，校验失败返回 401。
func Auth(cfg *config.Config, whitelist []string) gin.HandlerFunc {
	skip := make(map[string]struct{}, len(whitelist))
	for _, p := range whitelist {
		skip[strings.TrimSpace(p)] = struct{}{}
	}
	secret := cfg.JWTSecret

	return func(c *gin.Context) {
		// 白名单路径直接放行，不做 token 校验。
		if _, ok := skip[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		// 校验 Authorization: Bearer <token> 格式。
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

		// 验签 + 过期校验。
		claims, err := jwt.Parse(secret, token)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, bizerror.CodeUnauthorized.Code, bizerror.CodeUnauthorized.Message)
			c.Abort()
			return
		}

		// 校验通过：将 customerid 注入请求头，供后续业务读取。
		c.Request.Header.Set(customerIDHeader, strconv.FormatInt(claims.UID, 10))
		c.Set(customerIDContextKey, claims.UID)

		c.Next()
	}
}
