// Package tele implements IMAdapter for Telegram using telebot.
package tele

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/21ess/frozeel/adapter"
	"gopkg.in/telebot.v4"
)

// TelegramAdapter implements adapter.IMAdapter for Telegram.
type TelegramAdapter struct {
	bot             *telebot.Bot
	messageHandler  adapter.MessageHandler
	commandHandlers map[string]adapter.MessageHandler
}

// NewTelegramAdapter creates a new Telegram adapter with the given bot token.
func NewTelegramAdapter(token string) (adapter.IMAdapter, error) {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 1 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	ta := &TelegramAdapter{
		bot:             bot,
		commandHandlers: make(map[string]adapter.MessageHandler),
	}

	return ta, nil
}

func (t *TelegramAdapter) Start(ctx context.Context) error {
	// Register a catch-all text handler that dispatches to messageHandler.
	t.bot.Handle(telebot.OnText, func(c telebot.Context) error {
		if t.messageHandler == nil {
			return errors.New("message handler not set")
		}
		msg := toIncomingMessage(c.Message())
		t.messageHandler(ctx, msg)
		return nil
	})

	// Start blocks until bot.Stop() is called.
	t.bot.Start()
	return nil
}

func (t *TelegramAdapter) Stop() error {
	t.bot.Stop()
	return nil
}

func (t *TelegramAdapter) OnMessage(handler adapter.MessageHandler) {
	t.messageHandler = handler
}

func (t *TelegramAdapter) OnCommand(command string, handler adapter.MessageHandler) {
	t.commandHandlers[command] = handler

	t.bot.Handle("/"+command, func(c telebot.Context) error {
		msg := toIncomingMessage(c.Message())
		msg.IsCommand = true
		msg.Command = command
		if args := strings.Fields(msg.Text); len(args) > 1 {
			msg.CommandArgs = args[1:]
		}
		handler(context.Background(), msg)
		return nil
	})
}

func (t *TelegramAdapter) SendText(ctx context.Context, id int64, text string) error {
	chat := &telebot.Chat{ID: id}
	_, err := t.bot.Send(chat, text)
	return err
}

func (t *TelegramAdapter) ReplyText(ctx context.Context, chatID string, replyToMsgID string, text string) error {
	id, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chatID %q: %w", chatID, err)
	}
	msgID, err := strconv.Atoi(replyToMsgID)
	if err != nil {
		return fmt.Errorf("invalid replyToMsgID %q: %w", replyToMsgID, err)
	}
	chat := &telebot.Chat{ID: id}
	_, err = t.bot.Send(chat, text, &telebot.SendOptions{
		ReplyTo: &telebot.Message{ID: msgID},
	})
	return err
}

// toIncomingMessage converts a telebot.Message to adapter.IncomingMessage.
func toIncomingMessage(m *telebot.Message) adapter.IncomingMessage {
	chatType := adapter.ChatPrivate
	if m.Chat.Type == telebot.ChatGroup || m.Chat.Type == telebot.ChatSuperGroup {
		chatType = adapter.ChatGroup
	}

	return adapter.IncomingMessage{
		MessageID:  strconv.Itoa(m.ID),
		Text:       m.Text,
		SenderID:   strconv.FormatInt(m.Sender.ID, 10),
		SenderName: strings.TrimSpace(m.Sender.FirstName + " " + m.Sender.LastName),
		ChatID:     m.Chat.ID,
		ChatType:   chatType,
	}
}
