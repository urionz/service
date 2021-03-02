package cache_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urionz/goofy"
	"github.com/urionz/service/cache"
	"github.com/urionz/service/config"
	"github.com/urionz/service/filesystem"
	"github.com/urionz/service/redis"
)

func TestNewServiceProvider(t *testing.T) {
	require.NotPanics(t, func() {
		goofy.Default.AddServices(config.NewServiceProvider, filesystem.NewServiceProvider, redis.NewServiceProvider, cache.NewServiceProvider, func(conf config.IConfig, c cache.Factory) {
			store := c.Store()
			require.NoError(t, store.Set("testk", "testv", 0))
			require.Equal(t, "testv", store.Get("testk"))
			require.NoError(t, store.Forget("testk"))
		}).Run()
	})
}
