package bangumi_test

import (
	"context"
	"os"
	"testing"

	"github.com/21ess/frozeel/provider"
	"github.com/21ess/frozeel/provider/bangumi"
)

func TestBmProvider_GetRandomCharacter(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		opts    []provider.SubjectOption
		want    *provider.Character
		wantErr bool
	}{
		// {
		// 	"empty options",
		// 	[]provider.SubjectOption{},
		// 	nil,
		// 	false,
		// },
		{
			name: "with year range and tags",
			opts: []provider.SubjectOption{
				provider.WithYearRange(2024, 2026),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bangumi.BmProvider{
				Token: os.Getenv("BM_TOKEN"),
			}
			got, gotErr := b.GetRandomCharacter(context.Background(), tt.opts...)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetRandomCharacter() failed: %v", gotErr)
				}
				return
			}
			t.Logf("got character: %+v", got)
		})
	}
}
