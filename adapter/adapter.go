package adapter

import "context"

// MessageHandler 消息处理回调函数
type MessageHandler func(ctx context.Context, msg IncomingMessage)

// IMAdapter 各种 IM 软件接入的适配器接口
type IMAdapter interface {
	// Start 启动 adapter，开始接收消息（阻塞）
	Start(ctx context.Context) error
	Stop() error

	// OnMessage 注册消息处理回调
	OnMessage(handler MessageHandler)

	// OnCommand 注册命令处理回调（如 /start, /quiz）
	OnCommand(command string, handler MessageHandler)

	// SendText 发送文本消息到指定会话
	SendText(ctx context.Context, chatID string, text string) error

	// ReplyText 回复某条消息
	ReplyText(ctx context.Context, chatID string, replyToMsgID string, text string) error
}
