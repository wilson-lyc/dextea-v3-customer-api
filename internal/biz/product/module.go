package product

import "github.com/jmoiron/sqlx"

func NewModule(database *sqlx.DB) *Handler {
	return NewHandler(NewService(NewRepository(database)))
}
