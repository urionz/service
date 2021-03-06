package config

import (
	"github.com/urionz/config"
	"github.com/urionz/ini/dotenv"
)

type IConfig interface {
	Get(key string, findByPath ...bool) interface{}
	Set(key string, val interface{}, setByPath ...bool) error
	String(key string, defVal ...string) string
	Strings(key string) (arr []string)
	Int(key string, defVal ...int) int
	Ints(key string) (arr []int)
	Int64(key string, defVal ...int64) (value int64)
	Uint(key string, defVal ...uint) (value uint)
	Bool(key string, defVal ...bool) bool
	Env(key string, defVal interface{}) interface{}
	Exists(key string, findByPath ...bool) bool
	LoadExists(...string) error
	Object(key string, findByPath ...bool) IConfig
	Data() map[string]interface{}
}

var _ IConfig = new(Configure)

type Configure struct {
	*config.Config
}

func (*Configure) Env(key string, defVal interface{}) interface{} {
	switch defVal.(type) {
	case bool:
		return dotenv.Bool(key, defVal.(bool))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return dotenv.Int(key, defVal.(int))
	case string:
		break
	}
	return dotenv.Get(key, defVal.(string))
}

func (c *Configure) Object(key string, findByPath ...bool) IConfig {
	conf := config.New(key)
	val, ok := c.GetValue(key, findByPath...)
	if !ok {
		conf.SetData(make(map[string]interface{}))
	} else {
		conf.SetData(val.(map[string]interface{}))
	}
	return &Configure{
		Config: conf,
	}
}

var serve *Configure

func LoadExists(files ...string) error {
	return serve.LoadExists(files...)
}

func Get(key string, findByPath ...bool) interface{} {
	return serve.Get(key, findByPath...)
}

func Exists(key string, findByPath ...bool) bool {
	return serve.Exists(key, findByPath...)
}

func Env(key string, defVal interface{}) interface{} {
	return serve.Env(key, defVal)
}

func Object(key string) IConfig {
	return serve.Object(key)
}

func String(key string, defVal ...string) string {
	return serve.String(key, defVal...)
}

func Strings(key string) []string {
	return serve.Strings(key)
}

func Int(key string, defVal ...int) int {
	return serve.Int(key, defVal...)
}

func Ints(key string) []int {
	return serve.Ints(key)
}

func Int64(key string, defVal ...int64) int64 {
	return serve.Int64(key, defVal...)
}

func Uint(key string, defVal ...uint) uint {
	return serve.Uint(key, defVal...)
}

func Bool(key string, defVal ...bool) bool {
	return serve.Bool(key, defVal...)
}
