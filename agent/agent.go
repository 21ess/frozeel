package agent

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"

	localbk "github.com/cloudwego/eino-ext/adk/backend/local"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/schema"
)

type FrozeelAgent struct {
	agentRunner *adk.Runner
	history     []*schema.Message
	ch          chan *schema.Message
}

func NewFrozeelAgent() {}

func NewAgentRunner(ctx context.Context, prompt string) *adk.Runner {
	model, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		APIKey:  os.Getenv("LLM_API_KEY"),
		Model:   os.Getenv("LLM_MODEL_ID"),
		BaseURL: os.Getenv("LLM_BASE_URL"),
	})

	if err != nil {
		return nil
	}

	backend, err := localbk.NewBackend(ctx, &localbk.Config{})
	if err != nil {
		slog.Log(ctx, slog.LevelError, "failed to create local backend", "error", err)
		return nil
	}

	agent, err := deep.New(ctx, &deep.Config{
		Name:           "frozeel",
		Description:    "Frozeel agent with filesystem access via LocalBackend.",
		ChatModel:      model,
		Instruction:    prompt,
		Backend:        backend,
		StreamingShell: backend,
		MaxIteration:   50,
	})

	if err != nil {
		slog.Log(ctx, slog.LevelError, "failed to create deep agent", "error", err)
		return nil
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
	})

	return runner
}

// Run 启动 Agent
func (f *FrozeelAgent) Run(ctx context.Context) error {
	var sb strings.Builder
loop:
	for {
		select {
		case <-ctx.Done():
			slog.Log(ctx, slog.LevelInfo, "agent runner stopped")
			break loop
		case msg, ok := <-f.ch:
			if !ok { // 管道关闭
				slog.Log(ctx, slog.LevelInfo, "agent runner stopped")
				break loop
			}

			f.history = append(f.history, msg)
			iter := f.agentRunner.Run(ctx, f.history)

			for {
				next, has := iter.Next()
				if !has {
					break
				}
				if next.Err != nil {
					return next.Err
				}
				if next.Output == nil && next.Output.MessageOutput == nil {
					continue
				}
				mv := next.Output.MessageOutput
				if mv.Role != schema.Assistant { // 还可能是工具调用
					continue
				}
				if mv.IsStreaming {
					mv.MessageStream.SetAutomaticClose()
					for {
						frame, err := mv.MessageStream.Recv()
						if errors.Is(err, io.EOF) {
							break
						}
						if err != nil {
							return err
						}
						if frame != nil && strings.Trim(frame.Content, " ") != "" {
							sb.WriteString(frame.Content)
						}
					}
					continue
				}

				sb.WriteString(mv.Message.Content)
				slog.Log(ctx, slog.LevelInfo, "agent output: "+sb.String())
			}
		}
	}
	return nil
}

// SendUserMsg Agent 运行过程中发送用户级信息
func (f *FrozeelAgent) SendUserMsg(msg string) {
	f.ch <- &schema.Message{Role: schema.User, Content: msg}
}
