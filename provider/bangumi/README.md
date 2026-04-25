# Bangumi Provider API Notes

This directory uses the Bangumi API at `https://api.bgm.tv`.

## Auth

- Optional header: `Authorization: Bearer <token>`

## Paths Used

### `POST /v0/search/subjects`

Used by `GetRandomSubject()` in [subject.go](/Users/cltx/Developer/frozeel/provider/bangumi/subject.go:22).

Simple params:

- Query:
  - `limit`
  - `offset`
- JSON body:
  - `keyword`
  - `filter.nsfw`
  - `filter.sort`
  - `filter.air_date`
  - `filter.indices`
  - `filter.tag`
  - `filter.type`
  - `filter.rating`
  - `filter.rank`

Notes:

- `air_date`, `rating`, and `rank` are built as range expressions like `>=2001-01-01`, `<2005-01-01`, `>=7.5`, `<100`.
- The provider first requests `limit=1&offset=0` to get `total`, then requests a random `offset`.

### `GET /v0/subjects/{subject_id}/characters`

Used by `GetRandomCharacter()` in [character.go](/Users/cltx/Developer/frozeel/provider/bangumi/character.go:14).

Simple params:

- Path:
  - `subject_id`
- Query:
  - `limit`
  - `offset`

Notes:

- Used to pick one random character from the chosen subject.

### `POST /v0/search/characters`

Used by:

- `GetCharacterByName()` in [character.go](/Users/cltx/Developer/frozeel/provider/bangumi/character.go:67)
- `getRandomCharacterFallback()` in [character.go](/Users/cltx/Developer/frozeel/provider/bangumi/character.go:48)

Simple params:

- Query:
  - `limit`
  - `offset`
- JSON body:
  - `keyword`
  - `filter.nsfw`

Notes:

- The fallback random-character path sends empty `keyword` and random `offset`.
- Name lookup sends `keyword=<name>`.
- Current implementation always sends `filter.nsfw=false` here.

## Provider Mapping

- Random subject:
  - `POST /v0/search/subjects`
- Random character from a subject:
  - `POST /v0/search/subjects`
  - `GET /v0/subjects/{subject_id}/characters`
- Random character fallback:
  - `POST /v0/search/characters`
- Character by name:
  - `POST /v0/search/characters`
