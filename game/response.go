package game

import "github.com/21ess/frozeel/provider"

// ResponseType represents the type of response emitted by the game.
type ResponseType int

const (
	RespText        ResponseType = iota
	RespImage                    // reserved
	RespGameStarted
	RespGuessResult
	RespHint
	RespGameEnded
	RespError
)

// GuessOutcome indicates how close a guess was.
type GuessOutcome int

const (
	GuessWrong   GuessOutcome = iota
	GuessClose                // reserved for fuzzy matching
	GuessCorrect
)

// Response is an output from the game state machine, received via the Out channel.
type Response struct {
	Type     ResponseType
	Text     string
	ImageURL string
	Guess    *GuessDetail
	Answer   *provider.Character
}

// GuessDetail contains information about a guess attempt.
type GuessDetail struct {
	PlayerID   string
	PlayerName string
	Outcome    GuessOutcome
	Feedback   string
}
