// Package game
package game

const (
	WaitForDesc = iota + 1
)

type BuildForm struct {
	Collection  Collection
	StartYear   int32
	StartSeason int32
	EndYear     int32
	EndSeason   int32
	Desc        string
	WaitFor     int
}

type BuildSession struct {
	BuildFormMap map[int64]*BuildForm
	FormMsgIDMap map[int64]int
}
