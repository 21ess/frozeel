// Package tele
package tele

import (
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/telebot.v4"
)

// 懒加载
var once sync.Once
var gTelegramBot *TelegramBot

type TelegramBot struct {
	bot *telebot.Bot
}

func GetBot() *TelegramBot {
	once.Do(func() {
		pref := telebot.Settings{
			Token:  os.Getenv("BOT_TOKEN"),
			Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		}

		b, err := telebot.NewBot(pref)
		if err != nil {
			log.Fatalf("telegram 机器人启动失败 err:%v", err)
		}

		b.Handle("/hello", func(c telebot.Context) error {
			return c.Send("Hello!")
		})

		b.Start()

		gTelegramBot = &TelegramBot{
			bot: b,
		}
	})
	return gTelegramBot
}
