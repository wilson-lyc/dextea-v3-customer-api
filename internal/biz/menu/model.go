package menu

import "time"

type Menu struct {
	ID          int64     `db:"id"          json:"id"`
	Name        string    `db:"name"        json:"name"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"  json:"updated_at"`
}

type MenuGroup struct {
	ID        int64     `db:"id"        json:"id"`
	MenuID    int64     `db:"menu_id"   json:"menu_id"`
	Name      string    `db:"name"      json:"name"`
	Sort      int       `db:"sort"      json:"sort"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type MenuProduct struct {
	GroupID   int64     `db:"group_id"   json:"group_id"`
	ProductID int64     `db:"product_id" json:"product_id"`
	Sort      int       `db:"sort"       json:"sort"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
