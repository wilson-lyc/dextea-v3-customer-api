package gallery

import "time"

// Gallery 图库，存储图片资源（url 为可访问地址，object_key 为对象存储 key）
type Gallery struct {
	ID        int64     `db:"id"         json:"id"`
	URL       string    `db:"url"        json:"url"`
	ObjectKey string    `db:"object_key" json:"object_key"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Name      string    `db:"name"       json:"name"`
}
