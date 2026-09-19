package store

import "github.com/dextea-v3/dextea-customer/api/internal/infra/storerpc"

func NewModule(client *storerpc.Client) *Handler {
	return NewHandler(NewService(client))
}
