package cache

import (
	"time"
)

type ICache interface {
	Get(key string, defVal ...interface{}) interface{}
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Clear() error
	GetMultiple(keys []string, defVal interface{}) map[string]interface{}
	SetMultiple(values map[string]interface{}, ttl ...time.Duration) error
	DelMultiple(keys []string) error
	Has(key string) bool
}

type BaseCache struct {
}

func (*BaseCache) Get(_ string, _ ...interface{}) interface{} {
	return nil
}
func (*BaseCache) Set(_ string, _ interface{}, _ time.Duration) error {
	return nil
}
func (*BaseCache) Delete(_ string) error {
	return nil
}
func (*BaseCache) Clear() error {
	return nil
}
func (*BaseCache) GetMultiple(_ []string, _ interface{}) map[string]interface{} {
	return map[string]interface{}{}
}
func (*BaseCache) SetMultiple(_ map[string]interface{}, _ ...time.Duration) error {
	return nil
}
func (*BaseCache) DelMultiple(_ []string) error {
	return nil
}
func (*BaseCache) Has(_ string) bool {
	return false
}

type Store interface {
	Get(key string) interface{}
	Many(keys []string) []interface{}
	Put(key string, value interface{}, seconds time.Duration) error
	PutMany(kv map[string]interface{}, seconds int) error
	Increment(key string, value ...int) error
	Decrement(key string, value ...int) error
	Forever(key string, value interface{}) error
	Forget(key string) error
	ItemKey(key string) string
	Flush() error
	GetPrefix() string
}

type BaseStore struct {
}

func (*BaseStore) Get(_ string) interface{} {
	return nil
}
func (*BaseStore) Many(_ []string) []interface{} {
	return nil
}
func (*BaseStore) Put(_ string, _ interface{}, _ time.Duration) error {
	return nil
}
func (*BaseStore) PutMany(_ map[string]interface{}, _ int) error {
	return nil
}
func (*BaseStore) Increment(_ string, _ ...int) error {
	return nil
}
func (*BaseStore) Decrement(_ string, _ ...int) error {
	return nil
}
func (*BaseStore) Forever(_ string, _ interface{}) error {
	return nil
}
func (*BaseStore) Forget(_ string) error {
	return nil
}
func (*BaseStore) Flush() error {
	return nil
}
func (*BaseStore) GetPrefix() string {
	return ""
}
func (*BaseStore) ItemKey(key string) string {
	return key
}

type Closure = func() interface{}

type IRepository interface {
	ICache
	Tags(names ...string) (ITaggableStore, error)
	Pull(key string, defVal ...interface{}) interface{}
	Put(key string, value interface{}, ttl time.Duration) error
	Add(key string, value interface{}, ttl ...time.Duration) error
	Increment(key string, value ...int) error
	Decrement(key string, value ...int) error
	Forever(key string, value interface{}) error
	Remember(key string, ttl time.Duration, closure Closure) interface{}
	Sear(key string, closure Closure) interface{}
	RememberForever(key string, closure Closure) interface{}
	Forget(key string) error
	GetStore() ICache
}

type BaseRepository struct {
	BaseCache
}

func (*BaseRepository) Pull(_ string, _ ...interface{}) interface{} {
	return nil
}
func (*BaseRepository) Put(_ string, _ interface{}, _ time.Duration) error {
	return nil
}
func (*BaseRepository) Add(_ string, _ interface{}, _ ...time.Duration) error {
	return nil
}
func (*BaseRepository) Increment(_ string, _ ...int) error {
	return nil
}
func (*BaseRepository) Decrement(_ string, _ ...int) error {
	return nil
}
func (*BaseRepository) Forever(_ string, _ interface{}) error {
	return nil
}
func (*BaseRepository) Remember(_ string, _ time.Duration, _ Closure) interface{} {
	return nil
}
func (*BaseRepository) Sear(_ string, _ Closure) interface{} {
	return nil
}
func (*BaseRepository) RememberForever(_ string, _ Closure) interface{} {
	return nil
}
func (*BaseRepository) Forget(_ string) error {
	return nil
}
func (*BaseRepository) GetStore() ICache {
	return nil
}

type Factory interface {
	Store(name ...string) IRepository
}

type BaseFactory struct {
}

func (*BaseFactory) Store(_ ...string) (IRepository, error) {
	return nil, nil
}
