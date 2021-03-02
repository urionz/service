package db

import (
	"github.com/goava/di"
	"github.com/urionz/goofy"
	"github.com/urionz/service/config"
	"github.com/urionz/service/db/migrate"
	"github.com/urionz/service/db/model"
	"github.com/urionz/service/db/seed"
)

func NewServiceProvider(app goofy.IApplication, conf config.IConfig) error {
	app.Provide(func() *Manager {
		return NewManager(conf)
	}, di.As(new(Factory)))
	app.AddCommanders(
		new(migrate.MakeCommand), new(migrate.Command),
		new(migrate.RollbackCommand), new(migrate.StatusCommand),
		new(migrate.FreshCommand), new(migrate.ResetCommand),
		new(migrate.RefreshCommand), new(model.Command), new(seed.Command),
	)
	return nil
}
