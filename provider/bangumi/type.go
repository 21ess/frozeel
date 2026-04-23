// Package bangumi provides a provider for Bangumi characters.
package bangumi

type bangumiSearchResponse struct {
	Data   []bangumiCharacter `json:"data"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

type bangumiCharacter struct {
	BirthMon int     `json:"birth_mon"`
	BirthDay int     `json:"birth_day"`
	Gender   string  `json:"gender"`
	Images   images  `json:"images"`
	Summary  string  `json:"summary"`
	Name     string  `json:"name"`
	Infobox  []boxKV `json:"infobox"`
	NSFW     bool    `json:"nsfw"`
}

type boxKV struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type images struct {
	Small  string `json:"small"`
	Grid   string `json:"grid"`
	Large  string `json:"large"`
	Medium string `json:"medium"`
}
