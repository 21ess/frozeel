// Package game
package game

import (
	"time"
)

type Stat struct {
	Comments int `json:"comments"`
	Collects int `json:"collects"`
}

type Creator struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// Collection 进存储
type Collection struct {
	ID        int64          `json:"id"` // 本地数据库 ID
	Title     string         `json:"title"`
	Desc      string         `json:"desc"`
	Total     int            `json:"total"`
	Stat      Stat           `json:"stat"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Creator   Creator        `json:"creator"`
	NSFW      bool           `json:"nsfw"`
	MetaData  map[string]any `json:"meta_data"` // 比如 data_src：Bangumi，bangumi_indices_id: 29
}
