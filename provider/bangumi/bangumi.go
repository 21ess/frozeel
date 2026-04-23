// Package bangumi
package bangumi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	rand2 "math/rand/v2"
	"net/http"
	"time"

	"github.com/21ess/frozeel/provider"
)

type BmProvider struct {
	Token string
}

func (b *BmProvider) GetRandomSubject(ctx context.Context, start, end time.Time) (*provider.Subject, error) {
	// TODO implement me
	panic("implement me")
}

var TotOffset = 999

func (b *BmProvider) GetRandomCharacter(ctx context.Context) (*provider.Character, error) {
	offset := rand2.N(TotOffset)
	resp, err := b.searchCharacters(ctx, offset, 1, "")
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("failed to fetch random character")
	}

	if resp.Total != 0 && resp.Total != TotOffset {
		TotOffset = resp.Total // update total offset
	}

	char := resp.Data[0]
	return &provider.Character{
		Name:     char.Name,
		Summary:  char.Summary,
		Image:    pickFirstImage(char.Images),
		Gender:   char.Gender,
		Birthday: parseBirthday(char),
		// Nsfw:     char.NSFW,
		Tags: parseTags(char.Infobox),
	}, nil
}

func (b *BmProvider) GetCharacterByName(ctx context.Context, name string) (*provider.Character, error) {
	resp, err := b.searchCharacters(ctx, 0, 1, name)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("failed to search character: %s", name)
	}

	if resp.Total != 0 && resp.Total != TotOffset {
		TotOffset = resp.Total // update total offset
	}

	char := resp.Data[0]
	return &provider.Character{
		Name:    char.Name,
		Summary: char.Summary,
		// Image:    pickFirstImage(char.Images),
		Gender:   char.Gender,
		Birthday: parseBirthday(char),
		// Nsfw:     char.NSFW,
		// Tags:     parseTags(char.Infobox),
	}, nil
}

func (b *BmProvider) searchCharacters(ctx context.Context, offset, limit int, keyword string) (*bangumiSearchResponse, error) {
	url := fmt.Sprintf("https://api.bgm.tv/v0/search/characters?limit=%d&offset=%d", limit, offset)
	payload := map[string]any{
		"keyword": keyword,
		"filter":  map[string]any{"nsfw": false},
	}

	bodyByte, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	body := io.NopCloser(bytes.NewReader(bodyByte))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("build request failed: %w", err)
	}
	// set header
	req.Header.Set("User-Agent", "frozeel-bot/0.1")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch characters: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bangumi api returned status %d", resp.StatusCode)
	}

	var result bangumiSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("invalid response from bangumi api: %w", err)
	}

	return &result, nil
}

func pickFirstImage(img images) string {
	if img.Medium != "" {
		return img.Medium
	}
	if img.Large != "" {
		return img.Large
	}
	if img.Grid != "" {
		return img.Grid
	}
	return img.Small
}

func parseBirthday(c bangumiCharacter) string {
	if c.BirthMon > 0 && c.BirthDay > 0 {
		return fmt.Sprintf("%d月%d日", c.BirthMon, c.BirthDay)
	}
	if c.BirthMon > 0 {
		return fmt.Sprintf("%d月", c.BirthMon)
	}
	if c.BirthDay > 0 {
		return fmt.Sprintf("%d日", c.BirthDay)
	}
	for _, info := range c.Infobox {
		if info.Key == "生日" {
			if v, ok := info.Value.(string); ok {
				return v
			}
		}
	}
	return ""
}

func parseTags(infobox []boxKV) []string {
	tags := make([]string, 0, len(infobox))
	for _, info := range infobox {
		if info.Value == nil {
			continue
		}
		switch v := info.Value.(type) {
		case string:
			if v != "" {
				tags = append(tags, v)
			}
		case []any:
			for _, raw := range v {
				switch item := raw.(type) {
				case string:
					if item != "" {
						tags = append(tags, item)
					}
				case map[string]any:
					if s, ok := item["v"].(string); ok && s != "" {
						tags = append(tags, s)
					}
				}
			}
		case map[string]any:
			if s, ok := v["v"].(string); ok && s != "" {
				tags = append(tags, s)
			}
		}
	}
	return tags
}
