package product

import (
	"github.com/dextea-v3/dextea-customer/api/internal/infra/config"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/productrpc"
)

func NewModule(cfg *config.Config) (*Handler, error) {
	client, err := productrpc.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return NewHandler(NewService(client)), nil
}
