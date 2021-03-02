package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urionz/goofy"
	"github.com/urionz/service/config"
	"github.com/urionz/service/db"
)

func TestNewServiceProvider(t *testing.T) {
	require.NotPanics(t, func() {
		goofy.Default.AddServices(config.NewServiceProvider, db.NewServiceProvider, func(manager db.Factory) {
			manager.Connection().Migrator().CreateTable()
		}).Run()
	})
}
