// Package tele implements IMAdapter for Telegram using telebot.
package tele

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/21ess/frozeel/adapter"
	"github.com/21ess/frozeel/domain/game"
	"github.com/21ess/frozeel/store"
	"github.com/21ess/frozeel/store/mongo"
	"gopkg.in/telebot.v4"
)

// TelegramAdapter implements adapter.IMAdapter for Telegram.
type TelegramAdapter struct {
	bot             *telebot.Bot
	messageHandler  adapter.MessageHandler
	commandHandlers map[string]adapter.MessageHandler
	buildSessionMap sync.Map
	db              store.GameDB
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
	ctx := context.Background()
	db, err := mongo.NewMongoGameDB(context.Background(), os.Getenv("MONGO_URL"))
	if err != nil {
		return nil, err
	}
	db.EnsureIndexes(ctx)

	ta := &TelegramAdapter{
		bot:             bot,
		commandHandlers: make(map[string]adapter.MessageHandler),
		db:              db,
	}

	// /build 和 IM 软件强耦合，直接在初始化注册
	ta.commandHandlers["build"] = func(ctx context.Context, msg adapter.IncomingMessage) {} // 仅做占位
	bot.Handle("/build", func(c telebot.Context) error {
		uid := c.Sender().ID
		chatId := c.Chat().ID

		sessionI, ok := ta.buildSessionMap.Load(chatId)
		var form *game.BuildForm
		var session *game.BuildSession
		if !ok {
			sessionI = &game.BuildSession{
				BuildFormMap: map[int64]*game.BuildForm{
					uid: {
						Collection:  game.Collection{},
						StartYear:   0,
						StartSeason: 0,
						EndYear:     0,
						EndSeason:   0,
						Desc:        "",
						WaitFor:     game.WaitForDesc,
					},
				},
				FormMsgIDMap: make(map[int64]int),
			}
			ta.buildSessionMap.Store(chatId, sessionI)
		}
		session = sessionI.(*game.BuildSession)
		if session.BuildFormMap[uid] == nil {
			session.BuildFormMap[uid] = &game.BuildForm{
				Collection:  game.Collection{},
				StartYear:   0,
				StartSeason: 0,
				EndYear:     0,
				EndSeason:   0,
				Desc:        "",
				WaitFor:     game.WaitForDesc,
			}
		}
		form = session.BuildFormMap[uid]

		msg, err := bot.Send(c.Recipient(), "📋 构建参数", ta.buildMenu(form)) // 从存储获取
		if err != nil {
			return err
		}

		session.FormMsgIDMap[uid] = msg.ID

		return nil
	})

	bot.Handle(&telebot.InlineButton{Unique: "build_dir"}, func(c telebot.Context) error {
		uid := c.Sender().ID
		indices, err := ta.db.ListCollections(context.Background(), 10)
		if err != nil {
			return c.Edit("查询目录失败，请重试")
		}
		userIndices, err := ta.db.ListCollectionsByUser(ctx, uid, game.Telegram)
		if err != nil {
			return c.Edit("查询目录失败，请重试")
		}

		indices = append(indices, userIndices...)
		if len(indices) == 0 {
			return c.Edit("暂无目录，请先创建目录")
		}
		return c.Edit("请选择目录：", ta.dirMenu(indices))
	})

	// 年份翻页
	bot.Handle(&telebot.InlineButton{Unique: "build_year_page"}, func(c telebot.Context) error {
		parts := strings.SplitN(c.Data(), ",", 2)
		if len(parts) != 2 {
			return c.Respond(&telebot.CallbackResponse{Text: "参数错误"})
		}
		callbackUnique := parts[0]
		page, err := strconv.Atoi(parts[1])
		if err != nil || page < 0 {
			return c.Respond(&telebot.CallbackResponse{Text: "参数错误"})
		}
		return c.Edit("选择年份：", ta.yearMenu(callbackUnique, page))
	})

	// 时间设置：显示起始年份选择
	bot.Handle(&telebot.InlineButton{Unique: "build_time"}, func(c telebot.Context) error {
		return c.Edit("选择起始年份：", ta.yearMenu("build_time_sy", 0))
	})

	// 选择起始年份 → 显示起始季度
	bot.Handle(&telebot.InlineButton{Unique: "build_time_sy"}, func(c telebot.Context) error {
		uid := c.Sender().ID
		form := ta.getForm(c.Chat().ID, uid)
		if form == nil {
			return c.Edit("请先 /build 构建信息")
		}
		year, err := strconv.ParseInt(c.Data(), 10, 32)
		if err != nil {
			return c.Edit("解析年份失败，请重试")
		}
		form.StartYear = int32(year)
		return c.Edit(fmt.Sprintf("起始年份: %d，选择起始季度：", year), ta.seasonMenu("build_time_ss"))
	})

	// 选择起始季度 → 显示结束年份
	bot.Handle(&telebot.InlineButton{Unique: "build_time_ss"}, func(c telebot.Context) error {
		uid := c.Sender().ID
		form := ta.getForm(c.Chat().ID, uid)
		if form == nil {
			return c.Edit("请先 /build 构建信息")
		}
		season, err := strconv.ParseInt(c.Data(), 10, 32)
		if err != nil {
			return c.Edit("解析季度失败，请重试")
		}
		form.StartSeason = int32(season)
		return c.Edit(fmt.Sprintf("起始: %dQ%d，选择结束年份：", form.StartYear, season), ta.yearMenu("build_time_ey", 0))
	})

	// 选择结束年份 → 显示结束季度
	bot.Handle(&telebot.InlineButton{Unique: "build_time_ey"}, func(c telebot.Context) error {
		uid := c.Sender().ID
		form := ta.getForm(c.Chat().ID, uid)
		if form == nil {
			return c.Edit("请先 /build 构建信息")
		}
		year, err := strconv.ParseInt(c.Data(), 10, 32)
		if err != nil {
			return c.Edit("解析年份失败，请重试")
		}
		form.EndYear = int32(year)
		return c.Edit(fmt.Sprintf("起始: %dQ%d，结束年份: %d，选择结束季度：", form.StartYear, form.StartSeason, year), ta.seasonMenu("build_time_es"))
	})

	// 选择结束季度 → 返回主面板
	bot.Handle(&telebot.InlineButton{Unique: "build_time_es"}, func(c telebot.Context) error {
		uid := c.Sender().ID
		form := ta.getForm(c.Chat().ID, uid)
		if form == nil {
			return c.Edit("请先 /build 构建信息")
		}
		season, err := strconv.ParseInt(c.Data(), 10, 32)
		if err != nil {
			return c.Edit("解析季度失败，请重试")
		}
		form.EndSeason = int32(season)
		return c.Edit("📋 构建参数", ta.buildMenu(form))
	})

	// 描述设置：进入等待文本输入状态
	bot.Handle(&telebot.InlineButton{Unique: "build_desc"}, func(c telebot.Context) error {
		uid := c.Sender().ID
		form := ta.getForm(c.Chat().ID, uid)
		if form == nil {
			return c.Edit("请先 /build 构建信息")
		}
		form.WaitFor = game.WaitForDesc
		return c.Edit("请发送描述文字，发送后点击 /build 查看更新")
	})

	// 提交构建
	bot.Handle(&telebot.InlineButton{Unique: "build_submit"}, func(c telebot.Context) error {
		uid := c.Sender().ID
		form := ta.getForm(c.Chat().ID, uid)
		if form == nil {
			return c.Edit("请先 /build 构建信息")
		}
		// TODO: 使用 form 数据创建游戏或持久化
		ta.clearForm(c.Chat().ID, uid)
		return c.Edit(fmt.Sprintf("✅ 构建完成！\n目录: %s\n时间: %dQ%d → %dQ%d\n描述: %s",
			form.Collection.Title, form.StartYear, form.StartSeason, form.EndYear, form.EndSeason, form.Desc))
	})

	// 取消构建
	bot.Handle(&telebot.InlineButton{Unique: "build_cancel"}, func(c telebot.Context) error {
		ta.clearForm(c.Chat().ID, c.Sender().ID)
		return c.Edit("❌ 已取消构建")
	})

	// 返回主面板
	bot.Handle(&telebot.InlineButton{Unique: "build_back"}, func(c telebot.Context) error {
		uid := c.Sender().ID
		form := ta.getForm(c.Chat().ID, uid)
		if form == nil {
			return c.Edit("请先 /build 构建信息")
		}
		return c.Edit("📋 构建参数", ta.buildMenu(form))
	})

	bot.Handle(&telebot.InlineButton{Unique: "build_dir_select"}, func(c telebot.Context) error {
		uid := c.Sender().ID
		sessionI, ok := ta.buildSessionMap.Load(c.Chat().ID)
		if !ok {
			return c.Edit("请先 /build 构建信息")
		}

		session := sessionI.(*game.BuildSession)
		col, err := strconv.ParseInt(c.Data(), 10, 64)
		if err != nil {
			return c.Edit("解析数据，请重试")
		}
		ctx := context.Background()
		collection, err := ta.db.GetCollection(ctx, col)
		defer ta.db.IncrCollectionPopularity(ctx, col) // 添加人气
		if err != nil {
			return c.Edit("查询数据，请重试")
		}

		session.BuildFormMap[uid].Collection = *collection

		return c.Edit("📋 构建参数", ta.buildMenu(session.BuildFormMap[uid]))
	})

	return ta, nil
}

func (t *TelegramAdapter) Start(ctx context.Context) error {
	// Register a catch-all text handler that dispatches to messageHandler.
	t.bot.Handle(telebot.OnText, func(c telebot.Context) error {
		uid := c.Sender().ID
		chatId := c.Chat().ID
		sessionI, ok := t.buildSessionMap.Load(chatId)
		if ok {
			if session := sessionI.(*game.BuildSession); session.BuildFormMap[uid] != nil {
				form := session.BuildFormMap[uid]
				switch form.WaitFor {
				case game.WaitForDesc:
					form.Desc = c.Text()
					return nil
				default:
				}
			}
		}

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
	if t.commandHandlers[command] != nil {
		return
	}
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

// 主面板，不需要 indices
func (t *TelegramAdapter) buildMenu(form *game.BuildForm) *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}

	dirLabel := "未设置"
	if form.Collection.ID != 0 {
		dirLabel = form.Collection.Title
	}

	timeLabel := "未设置"
	if form.StartYear != 0 {
		timeLabel = fmt.Sprintf("%dQ%d → %dQ%d", form.StartYear, form.StartSeason, form.EndYear, form.EndSeason)
	}

	descLabel := "未设置"
	if form.Desc != "" {
		descLabel = form.Desc
	}

	rm.InlineKeyboard = [][]telebot.InlineButton{
		{telebot.InlineButton{Unique: "build_dir", Text: "📂 目录: " + dirLabel}},
		{telebot.InlineButton{Unique: "build_time", Text: "📅 时间: " + timeLabel}},
		{telebot.InlineButton{Unique: "build_desc", Text: "📝 描述: " + descLabel}},
		{
			*rm.Data("🚀 提交", "build_submit", "").Inline(),
			*rm.Data("❌ 取消", "build_cancel", "").Inline(),
		},
	}

	return rm
}

// 目录子面板，点击 build_dir 时动态查询后调用
func (t *TelegramAdapter) dirMenu(indices []*game.Collection) *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}

	var rows [][]telebot.InlineButton
	for _, col := range indices { // Telegram 官方 API 限制了 Data 大小，所以只存 ID
		btn := rm.Data(col.Title, "build_dir_select", strconv.FormatInt(col.ID, 10)).Inline()
		rows = append(rows, []telebot.InlineButton{*btn})
	}
	// 返回主面板
	rows = append(rows, []telebot.InlineButton{
		*rm.Data("⬅️ 返回", "build_back", "").Inline(),
	})

	rm.InlineKeyboard = rows
	return rm
}

// getForm returns the BuildForm for the given chat and user, or nil if not found.
func (t *TelegramAdapter) getForm(chatID, uid int64) *game.BuildForm {
	sessionI, ok := t.buildSessionMap.Load(chatID)
	if !ok {
		return nil
	}
	return sessionI.(*game.BuildSession).BuildFormMap[uid]
}

// clearForm removes the user's build form from the session.
func (t *TelegramAdapter) clearForm(chatID, uid int64) {
	sessionI, ok := t.buildSessionMap.Load(chatID)
	if !ok {
		return
	}
	session := sessionI.(*game.BuildSession)
	delete(session.BuildFormMap, uid)
	delete(session.FormMsgIDMap, uid)
	if len(session.BuildFormMap) == 0 {
		t.buildSessionMap.Delete(chatID)
	}
}

// yearMenu generates an inline keyboard for year selection.
// page 0 = most recent 9 years, page 1 = 9 years before that, etc.
func (t *TelegramAdapter) yearMenu(callbackUnique string, page int) *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}
	currentYear := int32(time.Now().Year())

	endYear := currentYear - int32(page*9)
	startYear := endYear - 8

	var rows [][]telebot.InlineButton
	var row []telebot.InlineButton
	for y := startYear; y <= endYear; y++ {
		btn := rm.Data(strconv.Itoa(int(y)), callbackUnique, strconv.Itoa(int(y))).Inline()
		row = append(row, *btn)
		if len(row) == 3 {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	// 翻页按钮
	pageData := func(p int) string { return callbackUnique + "," + strconv.Itoa(p) }
	var navRow []telebot.InlineButton
	// 总是可以看更早的年份
	navRow = append(navRow, *rm.Data("« later", "build_year_page", pageData(page+1)).Inline())
	if page > 0 {
		navRow = append(navRow, *rm.Data("earlier »", "build_year_page", pageData(page-1)).Inline())
	}
	rows = append(rows, navRow)

	rows = append(rows, []telebot.InlineButton{
		*rm.Data("⬅️ 返回", "build_back", "").Inline(),
	})

	rm.InlineKeyboard = rows
	return rm
}

// seasonMenu generates an inline keyboard for season (Q1-Q4) selection.
func (t *TelegramAdapter) seasonMenu(callbackUnique string) *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}
	rm.InlineKeyboard = [][]telebot.InlineButton{
		{
			*rm.Data("Q1 (1-3月)", callbackUnique, "1").Inline(),
			*rm.Data("Q2 (4-6月)", callbackUnique, "2").Inline(),
		},
		{
			*rm.Data("Q3 (7-9月)", callbackUnique, "3").Inline(),
			*rm.Data("Q4 (10-12月)", callbackUnique, "4").Inline(),
		},
		{*rm.Data("⬅️ 返回", "build_back", "").Inline()},
	}
	return rm
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
