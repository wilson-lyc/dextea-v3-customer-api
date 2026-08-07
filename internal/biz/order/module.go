package order

import "github.com/dextea-v3/dextea-customer/api/internal/config"

func NewModule(cfg *config.Config) *Handler {
	return NewHandler(NewService(cfg))
}
