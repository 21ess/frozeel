// Package bangumi provides a provider for Bangumi characters.
package bangumi

const (
	BaseUrl = "https://api.bgm.tv"
)

type BmProvider struct {
	Token string
}

type bangumiSearchResponse struct {
	Data   []bangumiCharacter `json:"data"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

type bangumiSubjectResponse struct {
	Data   []bangumiSubject `json:"data"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
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

type bangumiSubject struct {
	ID       int                `json:"id"`
	Name     string             `json:"name"`
	NameCN   string             `json:"name_cn"`
	NameJP   string             `json:"name_jp"`
	AirDate  string             `json:"air_date"`
	Summary  string             `json:"summary"`
	Images   images             `json:"images"`
	Tags     []string           `json:"tags"`
	MetaTags []any              `json:"meta_tags"`
	NSFW     bool               `json:"nsfw"`
	Infobox  []bangumiInfoboxKV `json:"infobox"`
}

type bangumiSubjectCharacterResponse struct {
	Data   []bangumiSubjectCharacter `json:"data"`
	Total  int                       `json:"total"`
	Limit  int                       `json:"limit"`
	Offset int                       `json:"offset"`
}

type bangumiSubjectCharacter struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	NameCN   string   `json:"name_cn"`
	NameJP   string   `json:"name_jp"`
	Actor    string   `json:"actor"`
	Relation string   `json:"relation"`
	Role     string   `json:"role"`
	Images   images   `json:"images"`
	Tags     []string `json:"tags"`
}

type bangumiInfoboxKV struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}
