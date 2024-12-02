package botmother

type Logger interface {
	Debug(msg ...interface{})

	Debugf(template string, args ...interface{})

	Info(msg ...interface{})

	Infof(template string, args ...interface{})

	Warn(msg ...interface{})

	Warnf(template string, args ...interface{})

	Error(msg ...interface{})

	Errorf(template string, args ...interface{})

	Fatal(msg ...interface{})

	Fatalf(template string, args ...interface{})
}
