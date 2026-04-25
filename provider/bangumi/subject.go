// Package bangumi
package bangumi

import (
	"context"
	"encoding/json"
	"fmt"
	rand2 "math/rand/v2"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/21ess/frozeel/provider"
)

var totOffset = 999

// default query settings
const (
	DefaultSort      = "heat"
	DefaultYearRange = 30
)

func (b *BmProvider) GetRandomSubject(ctx context.Context, opts ...provider.SubjectOption) (*provider.Subject, error) {
	option := mergeSubjectQuery(opts...)
	filters := buildSubjectFilters(option)

	// get totle offset
	first, err := b.searchSubjects(ctx, 0, 1, filters)
	if err != nil {
		return nil, err
	}

	if first.Total <= 0 || len(first.Data) == 0 {
		return nil, fmt.Errorf("no subject found with given filters")
	}

	offset := 0
	if first.Total > 1 {
		offset = rand2.IntN(first.Total)
	}

	resp, err := b.searchSubjects(ctx, offset, 1, filters)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("failed to fetch random subject")
	}

	return convertSubject(resp.Data[0]), nil
}

func mergeSubjectQuery(opts ...provider.SubjectOption) *provider.SubjectQuery {
	merged := &provider.SubjectQuery{}
	for _, opt := range opts {
		if opt != nil {
			opt.Apply(merged)
		}
	}

	return merged
}

func buildSubjectFilters(opt *provider.SubjectQuery) map[string]any {
	filters := map[string]any{
		"nsfw": opt.Nsfw, // default to false
	}
	filters["sort"] = DefaultSort
	if opt.Sort != "" {
		filters["sort"] = opt.Sort
	}

	if airDateFilter := buildAirDateFilter(opt.YearRange); len(airDateFilter) > 0 {
		filters["air_date"] = airDateFilter
	}
	if len(opt.Indices) > 0 {
		indices := dedupeInts(opt.Indices)
		filters["indices"] = indices
	}
	if len(opt.Tags) > 0 {
		filters["tag"] = append([]string(nil), opt.Tags...)
	}
	if len(opt.Type) > 0 {
		filters["type"] = append([]int(nil), opt.Type...)
	}
	if ratingFilter := buildFloatRangeFilter(opt.Rating); len(ratingFilter) > 0 {
		filters["rating"] = ratingFilter
	}
	if rankFilter := buildIntRangeFilter(opt.Rank); len(rankFilter) > 0 {
		filters["rank"] = rankFilter
	}

	return filters
}

func buildAirDateFilter(yearRange provider.Range[int]) []string {
	startYear, endYear := yearRange.Lower, yearRange.Upper
	curYear := time.Now().Year()
	if startYear <= 0 && endYear <= 0 {
		// recent {DefaultYearRange} years
		startYear, endYear = curYear-DefaultYearRange, curYear
	}

	if startYear <= 0 {
		return []string{fmt.Sprintf("<%d-01-01", endYear)}
	}
	if endYear <= 0 {
		return []string{fmt.Sprintf(">=%d-01-01", startYear)}
	}
	if startYear > endYear {
		startYear, endYear = endYear, startYear
	}

	filter := make([]string, 0, 2)
	filter = append(filter, fmt.Sprintf(">=%d-01-01", startYear))
	filter = append(filter, fmt.Sprintf("<%d-01-01", endYear))
	return filter
}

func buildIntRangeFilter(r provider.Range[int]) []string {
	filter := make([]string, 0, 2)
	if r.Lower > 0 {
		filter = append(filter, fmt.Sprintf(">=%d", r.Lower))
	}
	if r.Upper > 0 {
		filter = append(filter, fmt.Sprintf("<%d", r.Upper))
	}
	return filter
}

func buildFloatRangeFilter(r provider.Range[float32]) []string {
	filter := make([]string, 0, 2)
	if r.Lower > 0 {
		filter = append(filter, ">="+strconv.FormatFloat(float64(r.Lower), 'f', -1, 32))
	}
	if r.Upper > 0 {
		filter = append(filter, "<"+strconv.FormatFloat(float64(r.Upper), 'f', -1, 32))
	}
	return filter
}

func dedupeInts(in []int) []int {
	s := slices.Clone(in)
	slices.Sort(s)

	out := make([]int, 0, len(s))
	var prev int
	seen := false
	for _, value := range s {
		if !seen || value != prev {
			out = append(out, value)
			prev = value
			seen = true
		}
	}
	return out
}

func (b *BmProvider) searchSubjects(ctx context.Context, offset, limit int, filters map[string]any) (*bangumiSubjectResponse, error) {
	url := fmt.Sprintf("%s/v0/search/subjects?limit=%d&offset=%d", BaseUrl, limit, offset)
	bodyMap := map[string]any{
		"keyword": "",
		"filter":  filters,
	}
	if len(bodyMap["filter"].(map[string]any)) == 0 {
		bodyMap["filter"] = map[string]any{}
	}

	bodyByte, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subject request body: %w", err)
	}

	var result bangumiSubjectResponse
	if err := provider.DoHTTPJSON(ctx, "POST", url, bodyByte, b.Token, &result); err != nil {
		return nil, fmt.Errorf("search subjects: %w", err)
	}

	return &result, nil
}

func convertSubject(sub bangumiSubject) *provider.Subject {
	return &provider.Subject{
		ID:          sub.ID,
		Name:        pickSubjectName(sub),
		Alias:       []string{sub.NameCN, sub.Name, sub.NameJP},
		PublishDate: parseAirDateToTime(sub.AirDate),
		Characters:  nil,
		Tags:        parseSubjectTags(sub),
		Image:       pickFirstImage(sub.Images),
		Nsfw:        sub.NSFW,
	}
}

func pickSubjectName(s bangumiSubject) string {
	if s.NameCN != "" {
		return s.NameCN
	}
	if s.Name != "" {
		return s.Name
	}
	if s.NameJP != "" {
		return s.NameJP
	}
	return ""
}

func parseSubjectTags(s bangumiSubject) []string {
	tags := make([]string, 0)
	if len(s.Tags) > 0 {
		tags = append(tags, s.Tags...)
	}

	if len(s.MetaTags) > 0 {
		meta := make([]string, 0, len(s.MetaTags))
		for _, raw := range s.MetaTags {
			switch v := raw.(type) {
			case string:
				meta = append(meta, v)
			case map[string]any:
				if t, ok := v["name"].(string); ok {
					meta = append(meta, t)
				}
			}
		}
		for _, tag := range meta {
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	// remove duplicates with stable order
	seen := make(map[string]struct{}, len(tags))
	uniq := make([]string, 0, len(tags))
	for _, t := range tags {
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		uniq = append(uniq, t)
	}

	sort.Strings(uniq)
	return uniq
}

func parseAirDateToTime(airDate string) time.Time {
	formats := []string{"2006-01-02", "2006-01", "2006"}
	for _, f := range formats {
		t, err := time.Parse(f, airDate)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}
