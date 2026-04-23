package provider

import (
	"context"
	"time"
)

type AnimeProvider interface {
	// GetRandomSubject get a random subject with optional time duration  TODO: add indices
	GetRandomSubject(ctx context.Context, start, end time.Time) (*Subject, error)

	// GetRandomCharacter
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
	Actor    string   `json:"actor"` // seiyu or voice actor
	Birthday string   `json:"birthday"`
	Relation string   `json:"relation"` // only useful to subjects, character's relation in a specifix subject
	Tags     []string `json:"tags"`     // TODO: how to set this?
	// Nsfw     bool     `json:"nsfw"`     // R18 or not / hentai or regular
}

// Subject is a set of characters, e.g. an anime, a manga...
type Subject struct {
	Name        string       `json:"name"`
	PublishDate time.Time    `json:"publish_date"`
	Characters  []*Character `json:"characters"`
	Tags        []string     `json:"tags"`
	Image       string       `json:"image"`
	Nsfw        bool         `json:"nsfw"` // R18 or not / hentai or regular
}
