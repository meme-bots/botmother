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
		GetLocaleFromStorage(ctx telebot.Context) (string, error)
	}

	BotMother struct {
		Bot         *telebot.Bot
		Logger      Logger
		layouts     sync.Map
		callbackMap sync.Map
		messageMap  sync.Map
		router      Router
		localeMap   *expirable.LRU[int64, string]
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
		Logger:    cfg.Logger,
		localeMap: expirable.NewLRU[int64, string](10000, nil, time.Hour),
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

func (bm *BotMother) GetLocale(ctx telebot.Context) string {
	tid := ctx.Sender().ID
	locale, ok := bm.localeMap.Get(tid)
	if ok {
		return locale
	} else {
		ret, err := bm.router.GetLocaleFromStorage(ctx)
		if err != nil {
			return DEFAULT_LOCALE
		}
		bm.localeMap.Add(tid, ret)
		return ret
	}
}

func (bm *BotMother) SetLocale(ctx telebot.Context, locale string) {
	bm.localeMap.Add(ctx.Sender().ID, locale)
}

func (bm *BotMother) GetText(ctx telebot.Context, text string) string {
	return bm.GetLayout(bm.GetLocale(ctx)).GetText(text)
}
