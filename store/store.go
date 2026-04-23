package store

import (
	"context"

	"github.com/21ess/frozeel/domain/game"
	"github.com/21ess/frozeel/provider"
)

type GameDB interface {
	// Collection 相关
	GetCollection(ctx context.Context, id int64) (*game.Collection, error)               // 获取数据库的作品集合
	ListCollections(ctx context.Context, count int) ([]*game.Collection, error)          // 获取数据库的作品集合 (默认按人气排序)
	CreateCollection(ctx context.Context, collection *game.Collection) error             // 添加集合
	BatchAddSubjectToCollection(ctx context.Context, subjects []*provider.Subject) error // 添加作品到集合
	IncrCollectionPopularity(ctx context.Context, collectionID int64) error              // 增加集合的人气

}

type AgentDB interface{}
