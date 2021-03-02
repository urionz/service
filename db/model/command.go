package model

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/jinzhu/inflection"
	"github.com/urionz/cobra"
	"github.com/urionz/color"
	"github.com/urionz/goofy"
	"github.com/urionz/goutil/strutil"
)

var modelStub = `package models

import (
	"github.com/urionz/goofy/db/model"
)

type {{ .StructName }} struct {
	model.BaseModel
}

func (*{{ .StructName }}) TableName() string {
	return "{{ .TableName }}"
}
`

type Command struct {
	name       string
	migration  bool
	repository bool
	service    bool
}

func (cmd *Command) Handle(app goofy.IApplication) *cobra.Command {
	command := &cobra.Command{
		Use:   "make:model",
		Short: "创建数据模型",
		RunE: func(c *cobra.Command, args []string) error {
			var prompt *survey.Input
			var name string
			for {
				if name != "" || len(args) >= 1 {
					break
				}
				prompt = &survey.Input{
					Message: "请输入文件名称：",
				}
				survey.AskOne(prompt, &name)
			}
			if name == "" {
				name = args[0]
			}
			tableName := strutil.ToSnake(inflection.Plural(name))

			if cmd.migration {
				if err := createMigration(app, tableName); err != nil {
					color.Errorln(err)
					return err
				}
			}

			if err := createModel(name, tableName, app.Workspace()); err != nil {
				color.Errorln(err)
				return err
			}

			if cmd.repository {
				if err := createRepository(app, name); err != nil {
					color.Errorln(err)
					return err
				}
			}

			if cmd.service {
				if err := createService(app, name); err != nil {
					color.Errorln(err)
					return err
				}
			}
			return nil
		},
	}
	return command
}

func createModel(name, tableName, root string) error {
	fileName := strings.ToLower(strutil.ToSnake(name))

	structName := strutil.ToCamel(name)

	filePath := path.Join(root, "models", fileName+".go")

	stubString, err := modelPopulateStub(structName, tableName, modelStub)

	if err != nil {
		return err
	}

	if f, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666); err == nil {
		f.WriteString(stubString)
	} else {
		return err
	}

	return nil
}

func modelPopulateStub(structName, tableName, stub string) (string, error) {
	var templateBuffer bytes.Buffer
	tpl, err := template.New("model").Parse(stub)
	if err != nil {
		return templateBuffer.String(), err
	}

	if err := tpl.ExecuteTemplate(&templateBuffer, "model", map[string]interface{}{
		"StructName": structName,
		"TableName":  tableName,
	}); err != nil {
		return templateBuffer.String(), err
	}

	return templateBuffer.String(), nil
}

func createMigration(app goofy.IApplication, tableName string) error {
	_, _, err := app.Call("make:migration", fmt.Sprintf("create_%s_table", tableName))
	return err
}

func createRepository(app goofy.IApplication, name string) error {
	_, _, err := app.Call("make:repository", name)
	return err
}

func createService(app goofy.IApplication, name string) error {
	_, _, err := app.Call("make:service", name)
	return err
}
