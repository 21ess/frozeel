package bangumi

import (
	"reflect"
	"testing"

	"github.com/21ess/frozeel/provider"
)

func TestBuildSubjectFiltersUsesHalfOpenRanges(t *testing.T) {
	t.Parallel()

	filters := buildSubjectFilters(&provider.SubjectQuery{
		YearRange: provider.Range[int]{Lower: 2001, Upper: 2005},
		Indices:   []int{3, 1, 3},
		Tags:      []string{"mecha", "school"},
		Type:      []int{2, 6},
		Nsfw:      true,
		Rating:    provider.Range[float32]{Lower: 7.5, Upper: 9},
		Rank:      provider.Range[int]{Lower: 10, Upper: 100},
		Sort:      "rank",
	})

	want := map[string]any{
		"air_date": []string{">=2001-01-01", "<2005-01-01"},
		"indices":  []int{1, 3},
		"tag":      []string{"mecha", "school"},
		"type":     []int{2, 6},
		"nsfw":     true,
		"rating":   []string{">=7.5", "<9"},
		"rank":     []string{">=10", "<100"},
		"sort":     "rank",
	}

	if !reflect.DeepEqual(filters, want) {
		t.Fatalf("buildSubjectFilters() = %#v, want %#v", filters, want)
	}
}

func TestBuildSubjectFiltersOmitsEmptyRanges(t *testing.T) {
	t.Parallel()

	filters := buildSubjectFilters(&provider.SubjectQuery{
		Rating: provider.Range[float32]{Upper: 8.5},
		Rank:   provider.Range[int]{Lower: 50},
	})

	if got, want := filters["sort"], DefaultSort; got != want {
		t.Fatalf("default sort = %v, want %v", got, want)
	}

	if got, want := filters["rating"], []string{"<8.5"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("rating filter = %#v, want %#v", got, want)
	}

	if got, want := filters["rank"], []string{">=50"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("rank filter = %#v, want %#v", got, want)
	}
}
