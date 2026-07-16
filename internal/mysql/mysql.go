// Package mysql 负责打开并管理 MySQL 数据库连接（基于 database/sql 与 sqlx）。
//
// 该包把「数据库」这一层收敛为 MySQL 专用实现：驱动固定为 go-sql-driver/mysql，
// 调用方只需传入拼接好的 DSN，无需再关心驱动注册与驱动名。
package mysql

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql" // 匿名导入，注册 mysql 驱动供 sqlx 使用
)

// New 打开一个 MySQL 连接并返回 *sqlx.DB。
//
// dsn 为 MySQL 连接串（形如 user:pass@tcp(host:port)/db?charset=utf8mb4&parseTime=true）。
// 若 dsn 为空表示不启用数据库，返回 (nil, nil)，由上层决定降级行为。
// 内部使用 sqlx.Connect，会在打开后立即 Ping 校验连通性。
func New(dsn string) (*sqlx.DB, error) {
	if dsn == "" {
		return nil, nil
	}
	// 驱动已在包内匿名导入注册，driver 固定为 "mysql"。
	return sqlx.Connect("mysql", dsn)
}
