package redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/urionz/goofy"
	"github.com/urionz/service/config"
)

type Manager struct {
	app         goofy.IApplication
	driver      string
	conf        config.IConfig
	connections sync.Map
}

var _ Factory = (*Manager)(nil)

func NewRedisManager(app goofy.IApplication, conf config.IConfig) *Manager {
	manager := &Manager{
		app:  app,
		conf: conf,
	}
	return manager
}

func (m *Manager) Connection(name ...string) (*Connection, error) {
	var err error
	var conn *Connection
	if len(name) == 0 {
		name = append(name, "default")
	}
	if conn, ok := m.connections.Load(name[0]); ok {
		return conn.(*Connection), nil
	}
	if conn, err = m.configure(m.conf.Object(fmt.Sprintf("database.redis.%s", name[0])), name[0]); err != nil {
		return nil, err
	}
	m.connections.Store(name[0], conn)

	return conn, nil
}

func (m *Manager) configure(conf config.IConfig, name string) (*Connection, error) {
	conn := NewConnection(redis.NewClient(
		&redis.Options{
			Addr:     conf.String("address", "localhost:6379"),
			Password: conf.String("password", ""),
			DB:       conf.Int("db", 0),
		},
	)).SetName(name)
	return conn, conn.client.Ping(context.Background()).Err()
}
