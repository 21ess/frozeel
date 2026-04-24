// Package game
package game

import (
	"time"

	"github.com/21ess/frozeel/provider"
)

type Stat struct {
	Comments int `json:"comments"`
	Collects int `json:"collects"`
}

type IMType int32

const (
	Telegram IMType = iota + 1
)

type Creator struct {
	IMSrc  IMType `json:"im_src"`  // 区分不同来源
	UserId int64  `json:"user_id"` // 唯一 ID
}

// Collection 进存储
type Collection struct {
	ID         int64               `json:"id"` // 本地数据库 ID
	Title      string              `json:"title"`
	Desc       string              `json:"desc"`
	Total      int64               `json:"total"`
	Stat       Stat                `json:"stat"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
	Creator    Creator             `json:"creator"`
	NSFW       bool                `json:"nsfw,omitempty"`
	MetaData   map[string]any      `json:"meta_data,omitempty"` // 比如 data_src：Bangumi，bangumi_indices_id: 29
	Subjects   []*provider.Subject `json:"subjects,omitempty"`
	Popularity int64               `json:"popularity,omitempty"`
}
