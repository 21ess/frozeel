package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/21ess/frozeel/adapter/tele"
	"github.com/21ess/frozeel/config"
)

func main() {
	config.LoadDotEnv(".env") // 加载 .env 文件
	tele.GetBot()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
}
