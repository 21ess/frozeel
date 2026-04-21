// Package bangumi provides a provider for Bangumi characters.
package bangumi

import (
	"context"

	"github.com/21ess/frozeel/provider"
)

type Provider struct{}

func (b *Provider) GetRandomCharacter(ctx context.Context) (*provider.Character, error) {
	// send post req
	return nil, nil
}

func (b *Provider) GetCharacterByName(ctx context.Context, name string) (*provider.Character, error) {
	return nil, nil
}
