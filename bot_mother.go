package botmother

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type (
	Router interface {
		DefaultMessageHandler(ctx telebot.Context) error
	}

	BotMother struct {
		Bot           *telebot.Bot
		Logger        Logger
		layouts       sync.Map
		callbackMap   sync.Map
		messageMap    sync.Map
		router        Router
		localeMap     *expirable.LRU[int64, string]
		GetLocaleImpl func(telegramID int64) (string, error)
	}
)

const (
	DEFAULT_LOCALE = "en"
)

func NewBotMother(opts ...ConfigOption) (*BotMother, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if len(cfg.Token) == 0 {
		panic("token MUST be set")
	}

	bm := &BotMother{
		Logger: cfg.Logger,
	}

	pref := telebot.Settings{
		Token:       cfg.Token,
		Poller:      &telebot.LongPoller{Timeout: 10 * time.Second},
		ParseMode:   telebot.ModeHTML,
		Synchronous: false,
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, err
	}

	bot.Use(middleware.AutoRespond())
	bm.Bot = bot

	for _, locale := range cfg.Locales {
		layout, err := NewLayout(locale.Code, cfg.LayoutFile, locale.LocaleFile)
		if err != nil {
			return nil, err
		}
		bm.layouts.Store(locale.Code, layout)
	}

	if cfg.Router != nil {
		bm.InitRouter(cfg.Router)
	}

	return bm, nil
}

func (bm *BotMother) Start() {
	bot := bm.Bot

	if bm.router == nil {
		panic("router MUST be set before start")
	}

	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		return bm.handleMessage(c)
	})

	bot.Handle(telebot.OnCallback, func(c telebot.Context) error {
		return bm.handleCallback(c)
	})

	go bm.Bot.Start()
}

func (bm *BotMother) Stop() {
	bm.Bot.Stop()
}

func (bm *BotMother) GetLayout(locale string) *Layout {
	value, ok := bm.layouts.Load(locale)
	if !ok {
		return nil
	}
	return value.(*Layout)
}

func (bm *BotMother) GetDefaultLayout() *Layout {
	return bm.GetLayout(DEFAULT_LOCALE)
}

func (bm *BotMother) GetLocale(telegramID int64) string {
	locale, ok := bm.localeMap.Get(telegramID)
	if ok {
		return locale
	} else {
		ret, err := bm.GetLocaleImpl(telegramID)
		if err != nil {
			return DEFAULT_LOCALE
		}
		bm.localeMap.Add(telegramID, ret)
		return ret
	}
}

func (bm *BotMother) SetLocale(telegramID int64, locale string) {
	bm.localeMap.Add(telegramID, locale)
}

func (bm *BotMother) GetText(tid int64, text string) string {
	return bm.GetLayout(bm.GetLocale(tid)).GetText(text)
}
