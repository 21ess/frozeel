package game

import (
	"context"
	"fmt"
	"strings"

	"github.com/21ess/frozeel/provider"
)

// State represents the current state of the game state machine.
type State int

const (
	StateIdle        State = iota
	StateConfiguring       // reserved for future configuration phase
	StatePlaying
	StateEnded
)

// GameConfig holds game configuration (reserved for future use).
type GameConfig struct {
	// placeholder for future configurable options
	// TODO: add guess times limit?
}

// Game is a channel-driven state machine. The Bot layer sends Events to In and
// reads Responses from Out. All mutable state is owned by the Run goroutine,
// so no mutex is needed.
type Game struct {
	provider   provider.AnimeProvider
	state      State
	config     GameConfig
	answer     *provider.Character
	guessCount int

	In  chan Event
	Out chan Response
}

// NewGame creates a new Game with buffered channels.
func NewGame(p provider.AnimeProvider) *Game {
	return &Game{
		provider: p,
		state:    StateIdle,
		In:       make(chan Event, 16),
		Out:      make(chan Response, 16),
	}
}

// Run is the state machine main loop. It should be launched as a goroutine.
// It returns (and closes Out) when the game reaches StateEnded or the context
// is cancelled.
func (g *Game) Run(ctx context.Context) {
	defer close(g.Out)

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-g.In:
			if !ok {
				return
			}
			g.handle(ctx, ev)
			if g.state == StateEnded {
				return
			}
		}
	}
}

func (g *Game) handle(ctx context.Context, ev Event) {
	switch g.state {
	case StateIdle:
		g.handleIdle(ctx, ev)
	case StateConfiguring:
		g.handleConfiguring(ctx, ev)
	case StatePlaying:
		g.handlePlaying(ctx, ev)
	case StateEnded:
		// ignore events after game ended
	}
}

func (g *Game) handleIdle(ctx context.Context, ev Event) {
	switch ev.Type {
	case EventStart:
		g.startPlaying(ctx)
	default:
		g.Out <- Response{Type: RespError, Text: "游戏还没开始，请先 /start"}
	}
}

func (g *Game) handleConfiguring(ctx context.Context, ev Event) {
	switch ev.Type {
	case EventConfigure:
		// reserved: update g.config from ev.Options
		g.Out <- Response{Type: RespText, Text: "配置已更新"}
	case EventStart:
		g.startPlaying(ctx)
	case EventEnd:
		g.state = StateEnded
		g.Out <- Response{Type: RespGameEnded, Text: "游戏已取消"}
	default:
		g.Out <- Response{Type: RespError, Text: "配置阶段，请先完成配置或 /start 开始游戏"}
	}
}

func (g *Game) handlePlaying(ctx context.Context, ev Event) {
	switch ev.Type {
	case EventGuess:
		g.processGuess(ev)
	case EventHint:
		g.sendHint()
	case EventGiveUp:
		g.state = StateEnded
		g.Out <- Response{
			Type:   RespGameEnded,
			Text:   fmt.Sprintf("答案是：%s", g.answer.Name),
			Answer: g.answer,
		}
	case EventEnd:
		g.state = StateEnded
		g.Out <- Response{
			Type:   RespGameEnded,
			Text:   "游戏已结束",
			Answer: g.answer,
		}
	default:
		g.Out <- Response{Type: RespError, Text: "无法识别的操作"}
	}
}

func (g *Game) startPlaying(ctx context.Context) {
	c, err := g.provider.GetRandomCharacter(ctx)
	if err != nil {
		g.Out <- Response{Type: RespError, Text: fmt.Sprintf("获取角色失败: %v", err)}
		return
	}
	g.answer = c
	g.state = StatePlaying
	g.Out <- Response{
		Type: RespGameStarted,
		Text: "游戏开始！猜猜这是哪个角色吧",
	}
}

func (g *Game) processGuess(ev Event) {
	guess := strings.TrimSpace(ev.Payload)
	if guess == "" {
		g.Out <- Response{Type: RespError, Text: "猜测内容不能为空"}
		return
	}

	detail := &GuessDetail{
		PlayerID:   ev.SenderID,
		PlayerName: ev.SenderName,
	}

	// TODO 修改为 agent 判断
	if strings.EqualFold(guess, g.answer.Name) {
		detail.Outcome = GuessCorrect
		detail.Feedback = fmt.Sprintf("恭喜 %s 猜对了！答案就是 %s", ev.SenderName, g.answer.Name)
		g.state = StateEnded
		g.Out <- Response{
			Type:   RespGuessResult,
			Guess:  detail,
			Answer: g.answer,
		}
	} else {
		detail.Outcome = GuessWrong
		detail.Feedback = fmt.Sprintf("%s 猜错了，再想想？", ev.SenderName)
		g.Out <- Response{
			Type:  RespGuessResult,
			Guess: detail,
		}
	}
}

// TODO: enhance hit with more info according to guessCount, e.g. tags, summary, etc.
func (g *Game) sendHint() {
	if g.answer == nil {
		g.Out <- Response{Type: RespError, Text: "当前没有进行中的游戏"}
		return
	}

	var hints []string
	if g.answer.Gender != "" {
		hints = append(hints, fmt.Sprintf("性别: %s", g.answer.Gender))
	}
	if g.answer.Birthday != "" {
		hints = append(hints, fmt.Sprintf("生日: %s", g.answer.Birthday))
	}

	text := "暂时没有更多提示了"
	if len(hints) > 0 {
		text = "提示：" + strings.Join(hints, "，")
	}

	g.Out <- Response{Type: RespHint, Text: text}
}
