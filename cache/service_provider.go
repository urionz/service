package cache

import (
	"path"
	"runtime"

	"github.com/goava/di"
	"github.com/urionz/goofy"
	"github.com/urionz/service/config"
)

func NewServiceProvider(app goofy.IApplication, conf config.IConfig) error {
	if err := app.Provide(func() (*Manager, error) {
		_, f, _, _ := runtime.Caller(0)
		if err := conf.LoadExists(path.Join(path.Dir(f), "cache.toml")); err != nil {
			return nil, err
		}
		return NewManager(app, conf), nil
	}, di.As(new(Factory))); err != nil {
		return err
	}
	return nil
}
