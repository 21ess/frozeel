package provider

import "context"

type AnimeProvider interface {
	GetRandomCharacter(ctx context.Context) (*Character, error)
	GetCharacterByName(ctx context.Context, name string) (*Character, error)
	// GetCharacterTags(ctx context.Context, name string) ([]string, error)
}

// Character each game's answer, randomly picked from provider list
type Character struct {
	Name     string   `json:"name"`
	Summary  string   `json:"summary"`
	Image    string   `json:"image"` // default to large medium images
	Gender   string   `json:"gender"`
	Birthday string   `json:"birthday"`
	Nsfw     bool     `json:"nsfw"` // R18 or not / hentai or regular
	Tags     []string `json:"tags"` // TODO: how to set this?
}
