package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/21ess/frozeel/adapter"
	"github.com/21ess/frozeel/adapter/tele"
	"github.com/21ess/frozeel/game"
	"github.com/21ess/frozeel/provider/bangumi"
	"github.com/joho/godotenv"
)

var allGames = sync.Map{}

func main() {
	// config.LoadDotEnv("../.env")
	godotenv.Load()
	bot, err := tele.NewTelegramAdapter(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalf("failed to create adapter: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 注册命令处理
	bot.OnCommand("hello", func(ctx context.Context, msg adapter.IncomingMessage) {
		if err := bot.SendText(ctx, msg.ChatID, "Hello!"); err != nil {
			log.Printf("send error: %v", err)
		}
	})

	bot.OnCommand("start", func(ctx context.Context, msg adapter.IncomingMessage) {
		if _, ok := allGames.Load(msg.ChatID); ok {
			if err := bot.SendText(ctx, msg.ChatID, "游戏已经开始了哦～"); err != nil {
				log.Printf("send error: %v", err)
				return
			}
		}

		game := game.NewGame(&bangumi.Provider{Token: os.Getenv("BANGUMI_TOKEN")})
		allGames.Store(msg.ChatID, game)
		if err = game.HandleStart(ctx); err != nil {
			bot.SendText(ctx, msg.ChatID, fmt.Sprintf("Failed to start a game: %s", err))
			return
		}
		log.Printf("New game: %v", msg.ChatID)

		bot.SendText(ctx, msg.ChatID, "Here comes a challenge!!!")
		bot.SendText(ctx, msg.ChatID, "我已经想好了，来猜吧")

		// [Test]
		bot.SendText(ctx, msg.ChatID, fmt.Sprintf("[Test] 🫣不许偷看： %v", game.Answer))
	})

	bot.OnCommand("end", func(ctx context.Context, msg adapter.IncomingMessage) {
		_, ok := allGames.LoadAndDelete(msg.ChatID)
		if ok {
			bot.SendText(ctx, msg.ChatID, "冻鳗高手们拜拜啦👋～")
		} else {
			bot.SendText(ctx, msg.ChatID, "你是猪吗，都还没开始就结束?")
		}
		log.Printf("Ended a game: %d", msg.ChatID)
	})

	// TODO: send msg only visible to player who give up
	// seems not support directly
	bot.OnCommand("giveup", func(ctx context.Context, msg adapter.IncomingMessage) {
		bot.SendText(ctx, msg.ChatID, "要放弃了吗，真是杂鱼😛")
		bot.SendText(ctx, msg.ChatID, "暂不支持放弃后查看答案，略略略")
	})

	// TODO: handle guess

	// 注册通用消息处理
	bot.OnMessage(func(ctx context.Context, msg adapter.IncomingMessage) {
		fmt.Printf("[%d] %s: %s\n", msg.ChatID, msg.SenderName, msg.Text)
	})

	// 启动 bot（阻塞），放到 goroutine 中以便监听信号
	go func() {
		if err := bot.Start(ctx); err != nil {
			log.Fatalf("bot start error: %v", err)
		}
	}()

	log.Println("bot started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	cancel()
	if err := bot.Stop(); err != nil {
		log.Printf("stop error: %v", err)
	}
}
