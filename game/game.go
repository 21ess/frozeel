// Package game provides the main logic for the game
package game

import (
	"context"

	"github.com/21ess/frozeel/provider"
)

// Game new from each `/start` command, end after match some conditions ...
type Game struct {
	provider.AnimeProvider
	// TODO: add some status if needed
	Answer *provider.Character
}

func NewGame(p provider.AnimeProvider) *Game {
	return &Game{
		AnimeProvider: p,
	}
}

func (g *Game) HandleStart(ctx context.Context) error {
	c, err := g.GetRandomCharacter(ctx)
	if err != nil {
		return err
	}
	g.Answer = c
	return nil
}

// HandleGuess handles each guess
// TODO: clear guess operation:
// 1. @robot
// 2. using `/guess` like command => not friendly for users who using mobile maybe?
func (g *Game) HandleGuess() {
}
