package filesystem

import (
	"github.com/urionz/goofy"
)

func NewServiceProvider(app goofy.IApplication) error {
	return app.Provide(func() *Filesystem {
		return &Filesystem{}
	})
}
