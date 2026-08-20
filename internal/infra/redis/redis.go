// Package redis 负责打开并管理 Redis 连接（基于 go-redis/v9）。
package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// New 打开一个 Redis 连接并返回 *redis.Client。
//
// addr 形如 "127.0.0.1:6379"；password 为空表示无密码；db 为逻辑库编号（默认 0）。
// 若 addr 为空表示不启用 Redis，返回 (nil, nil)，由上层决定降级行为。
// 打开后会立即 Ping 校验连通性，失败则关闭连接并返回错误。
func New(addr, password string, db int) (*redis.Client, error) {
	if addr == "" {
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 打开后立即 Ping 校验连通性，避免把不可用的连接交给上层。
	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
