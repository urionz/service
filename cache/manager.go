package cache

import (
	"fmt"
	"sync"

	"github.com/urionz/goofy"
	"github.com/urionz/service/config"
	"github.com/urionz/service/filesystem"
	"github.com/urionz/service/redis"
)

const (
	DvrFile  = "file"
	DvrRedis = "redis"
)

type Manager struct {
	app    goofy.IApplication
	conf   config.IConfig
	stores sync.Map
}

var _ Factory = new(Manager)

func NewManager(app goofy.IApplication, conf config.IConfig) *Manager {
	manager := &Manager{
		app:  app,
		conf: conf,
	}
	return manager
}

// Get a cache store instance by name, wrapped in a repository.
func (m *Manager) Store(name ...string) IRepository {
	var store IRepository
	var err error
	if len(name) == 0 {
		name = append(name, m.getDefaultDriver())
	}
	if store, ok := m.stores.Load(name[0]); ok {
		return store.(IRepository)
	}
	if store, err = m.resolve(name[0]); err != nil {
		return nil
	}
	m.stores.Store(name[0], store)
	return store
}

// Get a cache driver instance.
func (m *Manager) Driver(driver ...string) IRepository {
	return m.Store(driver...)
}

// Resolve the given store.
func (m *Manager) resolve(name string) (repo IRepository, err error) {
	conf := m.getConfig(name)
	if conf == nil {
		return nil, fmt.Errorf("cache store %s is not defined", name)
	}
	driver := conf.String("driver")
	switch driver {
	case DvrFile:
		repo = m.createFileDriver(conf)
		break
	case DvrRedis:
		repo = m.createRedisDriver(conf)
		break
	}
	return repo, nil
}

// Create an instance of the file cache driver.
func (m *Manager) createFileDriver(conf config.IConfig) *Repository {
	var files *filesystem.Filesystem
	if err := m.app.Resolve(&files); err != nil {
		return nil
	}
	return m.repository(NewFileStore(files, conf.String("path", "./")))
}

// Create an instance of the Redis cache driver.
func (m *Manager) createRedisDriver(conf config.IConfig) *Repository {
	var rdm *redis.Manager
	var err error
	if err = m.app.Resolve(&rdm); err != nil {
		return nil
	}
	connection := conf.String("connection", "default")
	return m.repository(NewRedisStore(rdm, m.getPrefix(conf), connection))
}

// Create a new cache repository with the given implementation.
func (m *Manager) repository(store Store) *Repository {
	return NewRepository(store)
}

func (m *Manager) getConfig(name string) config.IConfig {
	return m.conf.Object(fmt.Sprintf("cache.stores.%s", name))
}

func (m *Manager) getDefaultDriver() string {
	return m.conf.String("cache.default")
}

// Get the cache prefix.
func (m *Manager) getPrefix(conf config.IConfig) string {
	return conf.String("prefix", m.conf.String("cache.prefix"))
}
