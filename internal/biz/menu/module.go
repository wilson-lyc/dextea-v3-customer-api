package menu

import (
	"github.com/dextea-v3/dextea-customer/api/internal/infra/config"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/productrpc"
	"github.com/jmoiron/sqlx"
)

func NewModule(database *sqlx.DB, cfg *config.Config) (*Handler, error) {
	client, err := productrpc.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return NewHandler(NewService(NewRepository(database), client)), nil
}
