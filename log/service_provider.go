package log

import (
	"github.com/urionz/goofy"
	"github.com/urionz/service/config"
)

func NewServiceProvider(app goofy.IApplication, conf config.IConfig) error {
	return app.Provide(func() *Logger {
		return NewLogger(conf)
	})
}
