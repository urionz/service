package web

import (
	"fmt"

	"github.com/kataras/iris/v12"
	"github.com/urionz/cobra"
	"github.com/urionz/goofy"
	"github.com/urionz/service/config"
)

func NewServiceProvider(app goofy.IApplication) error {
	webEngine := iris.New()
	app.AddCommanders(&engine{
		Application: webEngine,
	})
	return app.Provide(func() *iris.Application {
		return webEngine
	})
}

type engine struct {
	name  string
	debug bool
	port  int
	*iris.Application
}

func (cmd *engine) Handle(app goofy.IApplication) *cobra.Command {
	var conf config.IConfig
	command := &cobra.Command{
		Use:     "web",
		Aliases: []string{"http"},
		Short:   "开启http服务",
		RunE: func(c *cobra.Command, args []string) error {
			if err := app.Resolve(&conf); err != nil {
				panic(err)
			}
			cmd.fillArgs(conf)
			addr := fmt.Sprintf("0.0.0.0:%d", cmd.port)
			if cmd.debug {
				cmd.Logger().SetLevel("debug")
			}
			cmd.SetName(cmd.name)
			return cmd.Listen(
				addr, iris.WithOptimizations,
				iris.WithRemoteAddrPrivateSubnet("192.168.0.0", "192.168.255.255"),
				iris.WithRemoteAddrPrivateSubnet("10.0.0.0", "10.255.255.255"),
			)
		},
	}
	command.PersistentFlags().StringVarP(&cmd.name, "name", "n", "", "web应用名称")
	command.PersistentFlags().BoolVarP(&cmd.debug, "debug", "d", true, "是否开启web调试")
	command.PersistentFlags().IntVarP(&cmd.port, "port", "p", 0, "web服务监听端口")
	return command
}

func (cmd *engine) fillArgs(conf config.IConfig) {
	if cmd.name == "" {
		cmd.name = conf.Env("APP_NAME", conf.String("app.name", "web")).(string)
	}
	if cmd.port == 0 {
		cmd.port = conf.Env("HTTP_PORT", conf.Int("http.port", 3000)).(int)
	}
}
