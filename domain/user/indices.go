// Package game
package user

import (
	"time"

	"github.com/21ess/frozeel/domain/im"
	"github.com/21ess/frozeel/provider"
)

// Collection 收藏集
type Collection struct {
	ID         int64               `json:"id,omitempty"` // 本地数据库 ID
	Title      string              `json:"title,omitempty"`
	Desc       string              `json:"desc,omitempty"`
	Total      int64               `json:"total,omitempty"`
	Stat       Stat                `json:"stat,omitempty"`
	CreatedAt  time.Time           `json:"created_at,omitempty"`
	UpdatedAt  time.Time           `json:"updated_at,omitempty"`
	Creator    Creator             `json:"creator,omitempty"`
	NSFW       bool                `json:"nsfw,omitempty"`
	MetaData   map[string]any      `json:"meta_data,omitempty"` // 比如 data_src：Bangumi，bangumi_indices_id: 29
	Subjects   []*provider.Subject `json:"subjects,omitempty"`
	Popularity int64               `json:"popularity,omitempty"`
}

// Stat Bangumi的收藏集的统计数据
type Stat struct {
	Comments int `json:"comments"` // 评论
	Collects int `json:"collects"` // 收藏
}

type Creator struct {
	IMSrc  im.IMType `json:"im_src"`  // 区分不同来源
	UserId int64     `json:"user_id"` // 唯一 ID
}
