package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/21ess/frozeel/adapter"
	"github.com/21ess/frozeel/adapter/tele"
	"github.com/21ess/frozeel/game"
	"github.com/joho/godotenv"
)

var games sync.Map // chatID -> *game.Game, 一个群聊只能存在一场游戏

func main() {
	godotenv.Load()
	ctx := context.Background()

	bot, err := tele.NewTelegramAdapter(os.Getenv("BOT_TOKEN"))
	if err != nil {
		slog.Log(ctx, slog.LevelError, "failed to create adapter")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	bot.OnCommand("hello", func(ctx context.Context, msg adapter.IncomingMessage) {
		if err := bot.SendText(ctx, msg.ChatID, "Hello!"); err != nil {
			slog.Log(ctx, slog.LevelError, "send error", "error", err)
		}
	})

	// TODO: add guide ui for `/start` cmd with inline keyboard callback, palyer can choose time duration, indices etc.
	bot.OnCommand("start", func(ctx context.Context, msg adapter.IncomingMessage) {
		g, ok := loadGame(msg.ChatID)
		if ok && g.GetState() != game.StateIdle {
			bot.SendText(ctx, msg.ChatID, "游戏已经开始，请猜一猜吧！")
			descCurGame(g)
			return
		}

		if g == nil {
			g = game.NewGame()
		}

		startGame(ctx, msg, bot, g)
	})

	bot.OnCommand("end", func(ctx context.Context, msg adapter.IncomingMessage) {
		g, ok := loadGame(msg.ChatID)
		if !ok {
			bot.SendText(ctx, msg.ChatID, "你是猪吗，都还没开始就结束?")
			return
		}
		g.In <- game.Event{
			Type:       game.EventEnd,
			SenderID:   msg.SenderID,
			SenderName: msg.SenderName,
		}
	})

	bot.OnCommand("giveup", func(ctx context.Context, msg adapter.IncomingMessage) {
		g, ok := loadGame(msg.ChatID)
		if !ok {
			bot.SendText(ctx, msg.ChatID, "都还没开始，放弃什么呢？")
			return
		}
		g.In <- game.Event{
			Type:       game.EventGiveUp,
			SenderID:   msg.SenderID,
			SenderName: msg.SenderName,
		}
	})

	bot.OnCommand("hint", func(ctx context.Context, msg adapter.IncomingMessage) {
		g, ok := loadGame(msg.ChatID)
		if !ok {
			bot.SendText(ctx, msg.ChatID, "还没有进行中的游戏哦")
			return
		}
		g.In <- game.Event{
			Type:       game.EventHint,
			SenderID:   msg.SenderID,
			SenderName: msg.SenderName,
		}
	})

	// TODO: 一般消息需要过滤
	bot.OnMessage(func(ctx context.Context, msg adapter.IncomingMessage) {
		slog.Log(ctx, slog.LevelInfo, "message", "text", msg.Text, "chatID", msg.ChatID, "senderID", msg.SenderID, "senderName", msg.SenderName)

		// TODO 判断是否 @bot
		g, ok := loadGame(msg.ChatID)
		if !ok {
			return
		}
		g.In <- game.Event{
			Type:       game.EventGuess,
			SenderID:   msg.SenderID,
			SenderName: msg.SenderName,
			Payload:    msg.Text,
		}
	})

	go func() {
		if err := bot.Start(ctx); err != nil {
			slog.Log(ctx, slog.LevelError, "bot start error", "error", err)
		}
	}()

	slog.Log(ctx, slog.LevelInfo, "starting...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Log(ctx, slog.LevelInfo, "shutting down...")
	cancel()
	if err := bot.Stop(); err != nil {
		slog.Log(ctx, slog.LevelError, "stop error", "error", err)
	}
}

func loadGame(chatID int64) (*game.Game, bool) {
	v, ok := games.Load(chatID)
	if !ok {
		return nil, false
	}
	return v.(*game.Game), true
}

func startGame(ctx context.Context, msg adapter.IncomingMessage, bot adapter.IMAdapter, g *game.Game) {
	//p := &bangumi.BmProvider{Token: os.Getenv("BANGUMI_TOKEN")}
	//g := game.NewGame(p)
	games.Store(msg.ChatID, g)

	gameCtx, gameCancel := context.WithCancel(ctx)

	// Launch the game state machine
	go g.Run(gameCtx)

	// Launch the output renderer; cleans up when Out is closed
	go func() {
		defer gameCancel()
		defer games.Delete(msg.ChatID)
		for resp := range g.Out {
			renderResponse(ctx, bot, msg.ChatID, resp)
		}
		slog.Log(ctx, slog.LevelInfo, "game ended", "chatID", msg.ChatID)
	}()

	// Send the start event to kick off the game
	g.In <- game.Event{
		Type:       game.EventStart,
		SenderID:   msg.SenderID,
		SenderName: msg.SenderName,
	}
}

func renderResponse(ctx context.Context, bot adapter.IMAdapter, chatID int64, resp game.Response) {
	var text string

	switch resp.Type {
	case game.RespGameStarted:
		text = resp.Text
	case game.RespGuessResult:
		if resp.Guess != nil {
			text = resp.Guess.Feedback
		}
	case game.RespHint:
		text = resp.Text
	case game.RespGameEnded:
		text = resp.Text
	case game.RespError:
		text = fmt.Sprintf("[Error] %s", resp.Text)
	case game.RespText:
		text = resp.Text
	default:
		text = resp.Text
	}

	if text == "" {
		return
	}

	if err := bot.SendText(ctx, chatID, text); err != nil {
		slog.Log(ctx, slog.LevelError, "send error", "error", err)
	}
}

func descCurGame(g *game.Game) string {
	return fmt.Sprintf("当前游戏：%v", g)
}
