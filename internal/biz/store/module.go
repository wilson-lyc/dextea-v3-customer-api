package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func NewModule(database *sqlx.DB, rdb *redis.Client) *Handler {
	return NewHandler(NewService(NewRepository(database), rdb))
}
