// Package middleware 提供 gin 全局中间件。目前包含「全局异常拦截器」，
// 统一捕获并转换系统抛出的各类异常（数据库/网络/运行异常，以及自定义 BizError）。
package middleware

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/common/response"
)

// ExceptionInterceptor 返回一个 gin 中间件，作为「全局异常拦截器」。
//
// 它通过 recover 在请求处理链中捕获以下异常：
//   - 业务异常（*bizerror.BizError）：按原样透传业务码与提示（可 panic 抛出，也会被正确转换）；
//   - 数据库/网络/运行异常：统一清洗为对外的通用提示文案，绝不把底层细节
//     （SQL、堆栈、系统错误等）回显给前端；原始错误与堆栈仅记录在服务器日志中供排查。
//
// 无论何种异常，最终都通过 response 包写出统一的 APIResponse 结构。
func ExceptionInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				err := recoverToError(r)

				// 内部记录真实错误与堆栈，前端看不到。
				// 用 %+v 输出错误链，便于排查根因。
				log.Printf("[PANIC] recovered: %+v\n%s", err, debug.Stack())

				// 若响应尚未写出（如 handler 中途 panic），则写出统一的错误响应；
				// 已写出则不再覆盖，避免重复 WriteHeader 报错。
				if !c.Writer.Written() {
					response.WriteError(c, err)
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}

// recoverToError 将 recover 捕获的任意值规范化为 error。
func recoverToError(r interface{}) error {
	switch v := r.(type) {
	case error:
		return v
	case string:
		return fmt.Errorf("%s", v)
	default:
		return fmt.Errorf("%v", v)
	}
}
