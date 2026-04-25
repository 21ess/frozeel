// Package provider define common structure that all kinds of provider
// can use and interfaces that need to match
package provider

import (
	"context"
	"time"
)

// Range is a [l, u) range => >=l && <u
type Range[T comparable] struct {
	Lower T
	Upper T
}

type SubjectOption interface {
	Apply(query *SubjectQuery)
}

type SubjectOptionFunc func(query *SubjectQuery)

func (f SubjectOptionFunc) Apply(query *SubjectQuery) {
	f(query)
}

type SubjectQuery struct {
	YearRange Range[int]
	Indices   []int
	Tags      []string
	Type      []int
	Nsfw      bool
	Rating    Range[float32]
	Rank      Range[int]
	Sort      string
}

func WithYearRange(startYear, endYear int) SubjectOption {
	return SubjectOptionFunc(func(q *SubjectQuery) {
		q.YearRange = Range[int]{startYear, endYear}
	})
}

func WithIndices(indices ...int) SubjectOption {
	return SubjectOptionFunc(func(q *SubjectQuery) {
		q.Indices = append([]int(nil), indices...)
	})
}

func WithTags(tags ...string) SubjectOption {
	return SubjectOptionFunc(func(q *SubjectQuery) {
		q.Tags = append([]string(nil), tags...)
	})
}

func WithTypes(types ...int) SubjectOption {
	return SubjectOptionFunc(func(q *SubjectQuery) {
		q.Type = append([]int(nil), types...)
	})
}

func WithNSFW(v bool) SubjectOption {
	return SubjectOptionFunc(func(q *SubjectQuery) {
		q.Nsfw = v
	})
}

func WithRatings(l, u float32) SubjectOption {
	return SubjectOptionFunc(func(q *SubjectQuery) {
		q.Rating = Range[float32]{l, u}
	})
}

func WithRanks(l, u int) SubjectOption {
	return SubjectOptionFunc(func(q *SubjectQuery) {
		q.Rank = Range[int]{l, u}
	})
}

func WithSort(sortBy string) SubjectOption {
	return SubjectOptionFunc(func(q *SubjectQuery) {
		q.Sort = sortBy
	})
}

type AnimeProvider interface {
	// GetRandomSubject using options
	GetRandomSubject(ctx context.Context, opts ...SubjectOption) (*Subject, error)

	// Compatibility helpers
	// GetRandomSubjectByYear(ctx context.Context, startYear, endYear int) (*Subject, error)
	// GetRandomSubjectByIndices(ctx context.Context, indices []int) (*Subject, error)

	// GetRandomCharacter
	GetRandomCharacter(ctx context.Context, opts ...SubjectOption) (*Character, error)

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
	Relation string   `json:"relation"` // only useful to subjects, character's relation in a specific subject
	Tags     []string `json:"tags"`     // TODO: how to set this?
	// Nsfw     bool     `json:"nsfw"`     // R18 or not / hentai or regular
}

// Subject is a set of characters, e.g. an anime, a manga...
type Subject struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`  // chinese name first, if empty fallback to other
	Alias       []string     `json:"alias"` // names in other langs
	PublishDate time.Time    `json:"publish_date"`
	Characters  []*Character `json:"characters"`
	Tags        []string     `json:"tags"`
	Image       string       `json:"image"`
	Nsfw        bool         `json:"nsfw"` // R18 or not / hentai or regular
}
