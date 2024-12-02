package botmother

type (
	LocaleConfig struct {
		Code       string
		LocaleFile string
	}

	Config struct {
		Token      string
		LayoutFile string
		Locales    []LocaleConfig
		Logger     Logger
		Router     Router
	}

	ConfigOption func(*Config)
)

func DefaultConfig() *Config {
	return &Config{}
}

func WithToken(token string) ConfigOption {
	return func(c *Config) {
		c.Token = token
	}
}

func WithLogger(logger Logger) ConfigOption {
	return func(c *Config) {
		c.Logger = logger
	}
}

func WithLayoutFile(file string) ConfigOption {
	return func(c *Config) {
		c.LayoutFile = file
	}
}

func WithLocales(locales []LocaleConfig) ConfigOption {
	return func(c *Config) {
		c.Locales = locales[:]
	}
}

func WithRouter(router Router) ConfigOption {
	return func(c *Config) {
		c.Router = router
	}
}
