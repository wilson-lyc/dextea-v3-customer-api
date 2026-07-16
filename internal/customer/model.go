package customer

import "time"

type Customer struct {
	ID           int64     `db:"id"             json:"id"`
	Name         string    `db:"name"           json:"name"`
	WeixinOpenID string    `db:"weixin_open_id"  json:"weixin_open_id"`
	AlipayOpenID string    `db:"alipay_open_id"  json:"alipay_open_id"`
	Email        string    `db:"email"          json:"email"`
	Phone        string    `db:"phone"          json:"phone"`
	Password     string    `db:"password"       json:"-"`
	CreatedAt    time.Time `db:"created_at"     json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"     json:"updated_at"`
}
