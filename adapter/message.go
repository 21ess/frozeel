// Package adapter
package adapter

// IncomingMessage 统一的消息模型
type IncomingMessage struct {
	// 消息基本信息
	MessageID string
	Text      string

	// 发送者信息
	SenderID   string
	SenderName string

	// 会话信息
	ChatID   int64
	ChatType ChatType // Private or Group

	// 是否是命令 (如 /start, /quiz)
	IsCommand   bool
	Command     string // 不含 / 前缀
	CommandArgs []string
}

type ChatType int

const (
	ChatPrivate ChatType = iota
	ChatGroup
)
