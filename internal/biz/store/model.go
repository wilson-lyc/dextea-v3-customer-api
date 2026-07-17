package store

import (
	"time"
)

type Store struct {
	ID            int64     `db:"id"             json:"id"`
	Name          string    `db:"name"           json:"name"`
	Province      string    `db:"province"       json:"province"`
	City          string    `db:"city"           json:"city"`
	District      string    `db:"district"       json:"district"`
	Address       string    `db:"address"        json:"address"`
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

type StoreMenu struct {
	StoreID   int64     `db:"store_id"   json:"store_id"`
	MenuID    int64     `db:"menu_id"    json:"menu_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
