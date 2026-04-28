// Package game
package game

import "github.com/21ess/frozeel/domain/user"

const (
	WaitForDesc = iota + 1
)

// UserPerf 构建 Guess 的用户偏好
type UserPerf struct {
	Collection  user.Collection
	StartYear   int32
	StartSeason int32
	EndYear     int32
	EndSeason   int32
	Desc        string
	WaitFor     int
}

type BuildSession struct {
	BuildFormMap map[int64]*UserPerf
	FormMsgIDMap map[int64]int
}
