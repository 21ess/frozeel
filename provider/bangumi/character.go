package bangumi

import (
	"context"
	"encoding/json"
	"fmt"
	rand2 "math/rand/v2"

	"github.com/21ess/frozeel/provider"
)

// GetRandomCharacter get character from random subject
func (b *BmProvider) GetRandomCharacter(ctx context.Context, opts ...provider.SubjectOption) (*provider.Character, error) {
	sub, err := b.GetRandomSubject(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if sub == nil || sub.ID == 0 {
		return b.getRandomCharacterFallback(ctx)
		// return nil, fmt.Errorf("failed to resolve random subject for character selection")
	}

	chList, err := b.searchSubjectCharacters(ctx, sub.ID)
	if err != nil || chList == nil || len(chList.Data) == 0 {
		return b.getRandomCharacterFallback(ctx)
	}
	ch := chList.Data[rand2.IntN(len(chList.Data))]
	// get character details
	chDetail, err := b.searchCharacterById(ctx, ch.ID)
	if err != nil {
		return b.getRandomCharacterFallback(ctx)
	}
	chDetail.Actor = ch.Actor // get actor/seiyu from subject character relation
	return chDetail, nil
}

func (b *BmProvider) GetCharacterByName(ctx context.Context, name string) (*provider.Character, error) {
	resp, err := b.searchCharacters(ctx, 0, 1, name)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("failed to search character: %s", name)
	}

	if resp.Total != 0 && resp.Total != totOffset {
		totOffset = resp.Total // update total offset
	}

	char := resp.Data[0]
	return &provider.Character{
		Name:     char.Name,
		Summary:  char.Summary,
		Image:    pickFirstImage(char.Images),
		Gender:   char.Gender,
		Birthday: parseBirthday(char),
		Tags:     parseTags(char.Infobox),
		// Nsfw:     char.NSFW,
	}, nil
}

func (b *BmProvider) getRandomCharacterFallback(ctx context.Context) (*provider.Character, error) {
	offset := rand2.IntN(totOffset)
	resp, err := b.searchCharacters(ctx, offset, 1, "")
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("failed to fetch random character")
	}

	if resp.Total != 0 && resp.Total != totOffset {
		totOffset = resp.Total // update total offset
	}

	char := resp.Data[0]
	return &provider.Character{
		Name:     char.Name,
		Summary:  char.Summary,
		Image:    pickFirstImage(char.Images),
		Gender:   char.Gender,
		Birthday: parseBirthday(char),
		Tags:     parseTags(char.Infobox),
	}, nil
}

func (b *BmProvider) searchCharacterById(ctx context.Context, chId int) (*provider.Character, error) {
	url := fmt.Sprintf("%s/v0/characters/%d", BaseUrl, chId)
	var result bangumiCharacter
	if err := provider.DoHTTPJSON(ctx, "GET", url, nil, b.Token, &result); err != nil {
		return nil, fmt.Errorf("search character by id: %w", err)
	}

	return &provider.Character{
		Name:     result.Name,
		Summary:  result.Summary,
		Image:    pickFirstImage(result.Images),
		Gender:   result.Gender,
		Birthday: parseBirthday(result),
		Tags:     parseTags(result.Infobox),
	}, nil
}

func (b *BmProvider) searchSubjectCharacters(ctx context.Context, subjectID int) (*bangumiSubjectCharacterResponse, error) {
	url := fmt.Sprintf("%s/v0/subjects/%d/characters", BaseUrl, subjectID)
	var result bangumiSubjectCharacterResponse
	if err := provider.DoHTTPJSON(ctx, "GET", url, nil, b.Token, &result); err != nil {
		return nil, fmt.Errorf("search subject characters: %w", err)
	}

	return &result, nil
}

// searchCharacters search character by keyword, if keyword is empty
// it will return random characters with offset and limit
func (b *BmProvider) searchCharacters(ctx context.Context, offset, limit int, keyword string) (*bangumiSearchResponse, error) {
	url := fmt.Sprintf("%s/v0/search/characters?limit=%d&offset=%d", BaseUrl, limit, offset)
	payload := map[string]any{
		"keyword": keyword,
		"filter":  map[string]any{"nsfw": false},
	}

	bodyByte, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	var result bangumiSearchResponse
	if err := provider.DoHTTPJSON(ctx, "POST", url, bodyByte, b.Token, &result); err != nil {
		return nil, err
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

func pickSubjectCharacterName(char bangumiSubjectCharacter) string {
	if char.Name != "" {
		return char.Name
	}
	if char.NameCN != "" {
		return char.NameCN
	}
	return char.NameJP
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
