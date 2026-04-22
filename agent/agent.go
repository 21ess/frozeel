package agent

import (
	"context"

	"github.com/cloudwego/eino/adk"
)

func NewSimpleAgent() {
	adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name: "frozeel",
		Instruction: "",
		Model:
	})
}
