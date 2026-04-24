package agent

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

type FrozeelAgent struct {
	agentRunner *adk.Runner
}

func NewAgentRunner(ctx context.Context) *adk.Runner {
	model, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		APIKey:  os.Getenv("LLM_API_KEY"),
		Model:   os.Getenv("LLM_MODEL_ID"),
		BaseURL: os.Getenv("LLM_BASE_URL"),
	})

	if err != nil {
		return nil
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "frozeel",
		Instruction: "",
		Model:       model,
	})

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: agent,
	})

	return runner
}

// Run 启动 Agent
func (f *FrozeelAgent) Run(ctx context.Context, history []*schema.Message) {
	for {
		// 1. 读取用户输入
		//line := readUserInput()
		//if line == "" {
		//	break
		//}
		line := "hello"

		// 2. 追加用户消息到 history
		history = append(history, schema.UserMessage(line))

		// 3. 调用 Runner 执行 Agent
		events := f.agentRunner.Run(ctx, history)

		// 4. 收集模型的回复
		content, _ := collectAssistantFromEvents(events)

		// 5. 追加 assistant 消息到 history
		history = append(history, schema.AssistantMessage(content, nil))
	}
}

func collectAssistantFromEvents(events *adk.AsyncIterator[*adk.AgentEvent]) (string, error) {
	var sb strings.Builder

	for {
		event, ok := events.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return "", event.Err
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}

		mv := event.Output.MessageOutput
		if mv.Role != schema.Assistant {
			continue
		}

		if mv.IsStreaming {

		}

		if mv.Message != nil {
			sb.WriteString(mv.Message.Content)
			_, _ = fmt.Fprintln(os.Stdout, mv.Message.Content)
		} else {
			_, _ = fmt.Fprintln(os.Stdout)
		}

	}

	return sb.String(), nil
}
