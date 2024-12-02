package botmother

import (
	"sync"
	"time"

	"gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type (
	Router interface {
		DefaultMessageHandler(ctx telebot.Context) error
	}

	BotMother struct {
		bot         *telebot.Bot
		Logger      Logger
		layouts     sync.Map
		callbackMap sync.Map
		messageMap  sync.Map
		router      Router
	}
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
	bm.bot = bot

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
	bot := bm.bot

	if bm.router == nil {
		panic("router MUST be set before start")
	}

	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		return bm.handleMessage(c)
	})

	bot.Handle(telebot.OnCallback, func(c telebot.Context) error {
		return bm.handleCallback(c)
	})

	go bm.bot.Start()
}

func (bm *BotMother) Stop() {
	bm.bot.Stop()
}

func (bm *BotMother) GetLayout(locale string) *Layout {
	value, ok := bm.layouts.Load(locale)
	if !ok {
		return nil
	}
	return value.(*Layout)
}

func (bm *BotMother) GetDefaultLayout() *Layout {
	return bm.GetLayout("en")
}
