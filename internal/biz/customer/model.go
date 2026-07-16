package customer

import "time"

type Customer struct {
	ID           int64     `db:"id"             json:"id"`
	Name         string    `db:"name"           json:"name"`
	WeixinOpenID string    `db:"weixin_open_id"  json:"-"`
	AlipayOpenID string    `db:"alipay_open_id"  json:"-"`
	Email        string    `db:"email"          json:"email"`
	Phone        string    `db:"phone"          json:"phone"`
	Password     string    `db:"password"       json:"-"`
	Status       int       `db:"status"         json:"status"`
	Platform     int       `db:"platform"       json:"platform"`
	CreatedAt    time.Time `db:"created_at"     json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"     json:"updated_at"`
}

type Platform string

const (
	PlatformWeixin Platform = "weixin"
	PlatformAlipay Platform = "alipay"
)

func (p Platform) Valid() bool {
	switch p {
	case PlatformWeixin, PlatformAlipay:
		return true
	default:
		return false
	}
}

// DBValue 返回平台在数据库 customers.platform 列中的 int 值。
// 约定：支付宝=1，微信=0（与表结构一致，platform 列无默认值，必须显式写入）。
func (p Platform) DBValue() int {
	switch p {
	case PlatformAlipay:
		return 1
	case PlatformWeixin:
		return 0
	default:
		return 0
	}
}
