package cache

import (
	"time"

	"github.com/urionz/goutil/jsonutil"
	"github.com/urionz/service/redis"
)

type RedisStore struct {
	redis      redis.Factory
	prefix     string
	connection string
	TaggableStore
}

func NewRedisStore(redis redis.Factory, prefix, connection string) *RedisStore {
	return &RedisStore{
		redis:      redis,
		prefix:     prefix,
		connection: connection,
	}
}

func (r *RedisStore) Get(key string) interface{} {
	var dst interface{}
	value := r.Connection().Get(r.prefix + key)
	if value == "" {
		return nil
	}
	if err := jsonutil.Decode([]byte(value), &dst); err != nil {
		return value
	}
	return dst
}

func (r *RedisStore) Set(key string, value interface{}, ttl time.Duration) error {
	return r.Put(key, value, ttl)
}

func (r *RedisStore) Put(key string, value interface{}, seconds time.Duration) error {
	var err error
	var raw []byte
	if raw, err = jsonutil.Encode(value); err != nil {
		return err
	}
	return r.Connection().SetEX(key, string(raw), seconds)
}

func (r *RedisStore) Forever(key string, value interface{}) error {
	return r.Connection().Set(key, value, 0)
}

func (r *RedisStore) Tags(names ...string) (ITaggableStore, error) {
	return NewRedisTaggedCache(r, NewTagSet(r, names...)), nil
}

func (r *RedisStore) ItemKey(key string) string {
	return key
}

func (r *RedisStore) Connection() *redis.Connection {
	rds, err := r.redis.Connection(r.connection)
	if err != nil {
		panic(err)
	}
	return rds
}
