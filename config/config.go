package config

import (
	"github.com/urionz/config"
	"github.com/urionz/ini/dotenv"
)

type IConfig interface {
	Set(key string, val interface{}, setByPath ...bool) error
	String(key string, defVal ...string) string
	Int(key string, defVal ...int) int
	Bool(key string, defVal ...bool) bool
	Env(key string, defVal interface{}) interface{}
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

func Env(key string, defVal interface{}) interface{} {
	return serve.Env(key, defVal)
}

func Object(key string) IConfig {
	return serve.Object(key)
}

func String(key string, defVal ...string) string {
	return serve.String(key, defVal...)
}

func Int(key string, defVal ...int) (value int) {
	return serve.Int(key, defVal...)
}

func Bool(key string, defVal ...bool) (value bool) {
	return serve.Bool(key, defVal...)
}
