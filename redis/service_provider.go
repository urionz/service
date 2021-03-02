package redis

import (
	"github.com/urionz/goofy"
	"github.com/urionz/service/config"
)

func NewServiceProvider(app goofy.IApplication, conf config.IConfig) error {
	app.Provide(func() (*Manager, error) {
		return NewRedisManager(app, conf), nil
	})
	return nil
}
