package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/21ess/frozeel/adapter"
	"github.com/21ess/frozeel/adapter/tele"
	"github.com/21ess/frozeel/config"
)

func main() {
	config.LoadDotEnv(".env")

	bot, err := tele.NewTelegramAdapter(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalf("failed to create adapter: %v", err)
	}

	// 注册命令处理
	bot.OnCommand("hello", func(ctx context.Context, msg adapter.IncomingMessage) {
		if err := bot.SendText(ctx, msg.ChatID, "Hello!"); err != nil {
			log.Printf("send error: %v", err)
		}
	})

	// 注册通用消息处理
	bot.OnMessage(func(ctx context.Context, msg adapter.IncomingMessage) {
		fmt.Printf("[%s] %s: %s\n", msg.ChatID, msg.SenderName, msg.Text)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
