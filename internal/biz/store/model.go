package store

import (
	"encoding/json"
	"time"
)

type Store struct {
	ID            int64           `db:"id"             json:"id"`
	Name          string          `db:"name"           json:"name"`
	RegionCode    string          `db:"region_code"    json:"region_code"`
	// RegionNames 地区名称层级，格式如 ["广东省","广州市","番禺区"]
	RegionNames   json.RawMessage `db:"region_names"   json:"region_names"`
	Address       string          `db:"address"        json:"address"`
	Status        int             `db:"status"         json:"status"`
	BusinessHours string          `db:"business_hours" json:"business_hours"`
	Phone         string          `db:"phone"          json:"phone"`
	Longitude     float64         `db:"longitude"      json:"longitude"`
	Latitude      float64         `db:"latitude"       json:"latitude"`
	Account       string          `db:"account"        json:"account"`
	Password      string          `db:"password"       json:"-"`
	Email         string          `db:"email"          json:"email"`
	CreatedAt     time.Time       `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"     json:"updated_at"`
}
