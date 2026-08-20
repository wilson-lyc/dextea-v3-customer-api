// Package server 定义应用层与 HTTP 引擎之间的装配契约。
//
// 通过 Registerable 接口把 router 与具体业务 handler 解耦：
// router 只认接口，业务模块只要实现 Register(*gin.Engine) 即可被装载，
// 新增模块无需改动 router 与 main，满足开闭原则。
package server

import "github.com/gin-gonic/gin"

// Registerable 是业务模块对外暴露的路由注册契约。
// customer.Handler 已隐式实现该接口，无需任何改动即可接入。
type Registerable interface {
	Register(r *gin.Engine)
}
