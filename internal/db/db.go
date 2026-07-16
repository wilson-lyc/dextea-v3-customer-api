package db

import "github.com/jmoiron/sqlx"

// New 打开一个数据库连接并返回 *sqlx.DB。
//
// driverName 为已在别处匿名导入注册的驱动名，例如 "sqlite"、"mysql"、"postgres"。
// 调用方（如 main）需导入对应驱动以完成注册，例如：
//
//	import _ "modernc.org/sqlite" // 纯 Go 的 SQLite，无需 CGO
//
// 若 dsn 为空表示不启用数据库，返回 (nil, nil)，由上层决定降级行为。
// 使用 Connect 会在打开后立即 Ping 校验连通性。
func New(driverName, dsn string) (*sqlx.DB, error) {
	if dsn == "" {
		return nil, nil
	}
	return sqlx.Connect(driverName, dsn)
}
