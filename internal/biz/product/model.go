package product

import "time"

type Product struct {
	ID          int64     `db:"id"           json:"id"`
	Name        string    `db:"name"         json:"name"`
	Brief       string    `db:"brief"        json:"brief"`
	Description string    `db:"description"  json:"description"`
	Status      int       `db:"status"       json:"status"`
	Price       float64   `db:"price"        json:"price"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updated_at"`
}

type ProductImage struct {
	ProductID int64     `db:"product_id" json:"product_id"`
	ImageID   int64     `db:"image_id"   json:"image_id"`
	Type      int       `db:"type"       json:"type"`
	Sort      int       `db:"sort"       json:"sort"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type ProductStoreStatus struct {
	ProductID int64     `db:"product_id" json:"product_id"`
	StoreID   int64     `db:"store_id"   json:"store_id"`
	Status    int       `db:"status"     json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CustomizationItem struct {
	ID        int64     `db:"id"          json:"id"`
	ProductID int64     `db:"product_id"  json:"product_id"`
	Name      string    `db:"name"        json:"name"`
	Sort      int       `db:"sort"        json:"sort"`
	Status    int       `db:"status"      json:"status"`
	CreatedAt time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt time.Time `db:"updated_at"  json:"updated_at"`
}

type CustomizationOption struct {
	ID                 int64     `db:"id"                   json:"id"`
	ItemID             int64     `db:"item_id"              json:"item_id"`
	Name               string    `db:"name"                 json:"name"`
	Price              float64   `db:"price"                json:"price"`
	Sort               int       `db:"sort"                 json:"sort"`
	Status             int       `db:"status"               json:"status"`
	IngredientID       *int64    `db:"ingredient_id"        json:"ingredient_id"`
	IngredientQuantity *float64  `db:"ingredient_quantity"  json:"ingredient_quantity"`
	CreatedAt          time.Time `db:"created_at"           json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"           json:"updated_at"`
}

type CustomizationOptionStoreStatus struct {
	OptionID  int64     `db:"option_id"  json:"option_id"`
	StoreID   int64     `db:"store_id"   json:"store_id"`
	Status    int       `db:"status"     json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
