// Package tools
// IM 软件 bot 控制工具类
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/21ess/frozeel/adapter"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type IMSendTool struct {
	imAdapter adapter.IMAdapter
}

func NewTelegramSendTool(imAdapter adapter.IMAdapter) *IMSendTool {
	return &IMSendTool{
		imAdapter: imAdapter,
	}
}

func (i *IMSendTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "send_im_message",
		Desc: "通过IM软件接口发送消息到指定聊天",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"chat_id": {
				Type:     schema.Integer,
				Desc:     "聊天 ID（群组或私聊）",
				Required: true,
			},
			"text": {
				Type:     schema.String,
				Desc:     "要发送的文本消息内容",
				Required: true,
			},
		}),
	}, nil
}

func (i *IMSendTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		ChatID int64  `json:"chat_id"`
		Text   string `json:"text"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if args.ChatID == 0 {
		return "", fmt.Errorf("chat_id is required")
	}

	if args.Text == "" {
		return "", fmt.Errorf("text is required")
	}

	err := i.imAdapter.SendText(ctx, args.ChatID, args.Text)
	if err != nil {
		return "", fmt.Errorf("failed to send telegram message: %w", err)
	}

	result := map[string]any{
		"success": true,
		"message": fmt.Sprintf("Message sent to chat %d successfully", args.ChatID),
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonResult), nil
}
