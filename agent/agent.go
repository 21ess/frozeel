package agent

import (
	"context"
	"os"

	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/adk"
)

func NewSimpleAgent(ctx context.Context) {
	model, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		APIKey:  os.Getenv("LLM_API_KEY"),
		Model:   os.Getenv("LLM_MODEL_ID"),
		BaseURL: os.Getenv("LLM_BASE_URL"),
	})

	if err != nil {
		return
	}

	adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "frozeel",
		Instruction: "",
		Model:       model,
	})
}
