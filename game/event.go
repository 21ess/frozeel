package game

// EventType represents the type of event sent to the game state machine.
type EventType int

const (
	EventStart     EventType = iota
	EventConfigure           // reserved for future configuration phase
	EventGuess
	EventHint
	EventGiveUp
	EventEnd
)

// Event is an input to the game state machine, sent via the In channel.
type Event struct {
	Type       EventType
	SenderID   string
	SenderName string
	Payload    string            // guess text, etc.
	Options    map[string]string // configuration key-value pairs (reserved)
}
