package area

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
)

const cityColumns = `city`

// Repository 负责 area 模块的数据访问。
type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// GetDistinctCities 获取去重后的城市列表，按城市名排序。
func (r *Repository) GetDistinctCities(ctx context.Context) ([]string, error) {
	if r.db == nil {
		return nil, bizerror.ErrMysqlDisabled
	}

	query := `SELECT DISTINCT ` + cityColumns + ` FROM stores WHERE city != '' ORDER BY city`
	var cities []string
	if err := r.db.SelectContext(ctx, &cities, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return cities, nil
}
