package order

import "github.com/dextea-v3/dextea-customer/api/internal/infra/config"

func NewModule(cfg *config.Config) *Handler {
	return NewHandler(NewService(cfg))
}
