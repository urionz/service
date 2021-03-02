package seed

import (
	"github.com/urionz/cobra"
	"github.com/urionz/color"
	"github.com/urionz/goofy"
	"gorm.io/gorm"
)

type Command struct {
}

func (*Command) Handle(app goofy.IApplication) *cobra.Command {
	var db *gorm.DB
	command := &cobra.Command{
		Use:   "db:seed",
		Short: "生成数据",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.Resolve(&db); err != nil {
				return err
			}
			for _, seeder := range seederFiles {
				if err := seeder.Handle(db); err != nil {
					color.Errorln(err)
					return err
				}
			}
			return nil
		},
	}
	return command
}
