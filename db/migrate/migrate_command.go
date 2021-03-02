package migrate

import (
	"bytes"
	"errors"
	"html/template"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/golang-module/carbon"
	"github.com/urionz/cobra"
	"github.com/urionz/cobra/interact"
	"github.com/urionz/cobra/progress"
	"github.com/urionz/cobra/show"
	"github.com/urionz/collection"
	"github.com/urionz/color"
	"github.com/urionz/goofy"
	"github.com/urionz/goutil/fsutil"
	"github.com/urionz/goutil/strutil"
	"gorm.io/gorm"
)

var createStub = `package migrations

import (
	"github.com/goofy/service/db/migrate"
	"github.com/goofy/service/db/model"
	"gorm.io/gorm"
)

func init() {
	migrate.Register(&{{ .StructName }}{})
}

type {{ .StructName }} struct {
	model.BaseModel
}

func (table *{{ .StructName }}) TableName() string {
	return "{{ .TableName }}"
}

func (table *{{ .StructName }}) MigrateTimestamp() int {
	return {{ .Timestamp }}
}

func (table *{{ .StructName }}) Up(db *gorm.DB) error {
	if !db.Migrator().HasTable(table) {
		return db.Migrator().CreateTable(table)
	}
	return nil
}

func (table *{{ .StructName }}) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(table)
}
`

var blankStub = `package migrations

import (
	"github.com/urionz/goofy/db/migrate"
	"github.com/urionz/goofy/db/model"
	"gorm.io/gorm"
)

func init() {
	migrate.Register(&{{.StructName}}{})
}

type {{ .StructName }} struct {
	model.BaseModel
}

func (table *{{ .StructName }}) MigrateTimestamp() int {
	return {{ .Timestamp }}
}

func (table *{{ .StructName }}) TableName() string {
	return "{{ .TableName }}"
}

func (table *{{ .StructName }}) Up(db *gorm.DB) error {
	return nil
}

func (table *{{ .StructName }}) Down(db *gorm.DB) error {
	return nil
}
`

type Factory interface {
	Connection(...string) *gorm.DB
}

type Command struct {
	step int
}

func (cmd *Command) Handle(app goofy.IApplication) *cobra.Command {
	command := &cobra.Command{
		Use:   "migrate",
		Short: "运行迁移",
		RunE: func(c *cobra.Command, args []string) error {
			var manager Factory
			if err := app.Resolve(&manager); err != nil {
				return err
			}
			if err := runMigrate(cmd.step, manager); err != nil {
				color.Errorln(err)
			}
			return nil
		},
	}

	command.PersistentFlags().IntVarP(&cmd.step, "step", "s", 0, "指定迁移阶段")

	return command
}

type RollbackCommand struct {
	step int
}

func (cmd *RollbackCommand) Handle(app goofy.IApplication) *cobra.Command {
	command := &cobra.Command{
		Use:   "migrate:rollback",
		Short: "迁移回滚",
		RunE: func(c *cobra.Command, args []string) error {
			var manager Factory
			if err := app.Resolve(&manager); err != nil {
				return err
			}
			if err := runRollback(cmd.step, manager); err != nil {
				color.Errorln(err)
			}
			return nil
		},
	}

	command.PersistentFlags().IntVarP(&cmd.step, "step", "s", 0, "指定迁移阶段")

	return command
}

type RefreshCommand struct {
	step int
}

func (cmd *RefreshCommand) Handle(app goofy.IApplication) *cobra.Command {
	command := &cobra.Command{
		Use:   "migrate:refresh",
		Short: "刷新迁移",
		RunE: func(c *cobra.Command, args []string) error {
			var manager Factory
			if err := app.Resolve(&manager); err != nil {
				return err
			}
			var err error

			if cmd.step > 0 {
				err = runRollback(cmd.step, manager)
			} else {
				err = runReset(manager)
			}

			if err != nil {
				color.Errorln(err)
				return err
			}

			if err = runMigrate(cmd.step, manager); err != nil {
				color.Errorln(err)
				return err
			}
			return nil
		},
	}

	command.PersistentFlags().IntVarP(&cmd.step, "step", "s", 0, "指定迁移阶段")

	return command
}

type FreshCommand struct {
	step int
}

func (cmd *FreshCommand) Handle(app goofy.IApplication) *cobra.Command {
	command := &cobra.Command{
		Use:   "migrate:fresh",
		Short: "migrate fresh",
		RunE: func(c *cobra.Command, args []string) error {
			var manager Factory
			if err := app.Resolve(&manager); err != nil {
				return err
			}
			var tables []string
			db := manager.Connection()
			if err := db.Raw("show tables").Scan(&tables).Error; err != nil {
				return err
			}

			for _, table := range tables {
				if err := db.Migrator().DropTable(table); err != nil {
					return err
				}
			}

			color.Infoln("Dropped all tables successfully.")

			if err := runMigrate(cmd.step, manager); err != nil {
				color.Errorln(err)
				return err
			}
			return nil
		},
	}

	command.PersistentFlags().IntVarP(&cmd.step, "step", "s", 0, "指定迁移阶段")

	return command
}

type StatusCommand struct {
}

func (cmd *StatusCommand) Handle(app goofy.IApplication) *cobra.Command {
	var db *gorm.DB
	command := &cobra.Command{
		Use:   "migrate:status",
		Short: "查看迁移状态",
		RunE: func(c *cobra.Command, args []string) error {
			if err := app.Resolve(&db); err != nil {
				return err
			}
			var batches map[string]int
			var ran []*Model
			var err error
			repository := NewDBMigration(db)
			ran, err = repository.GetRan()
			if err != nil {
				color.Errorln(err)
				return nil
			}
			ranNameCollection := collection.NewObjPointCollection(ran).Pluck("Migration")

			batches, err = repository.GetMigrationBatches()
			if err != nil {
				color.Errorln(err)
				return err
			}
			table := show.NewTable("migrate status")
			table.Cols = []string{"Ran?", "Migration", "Batch"}
			for _, migrateFile := range getMigrationFiles() {
				migrationName := getMigrationName(migrateFile)
				if ranNameCollection.Contains(migrationName) {
					table.Cols = []string{
						color.String("<green>Yes</>"),
						migrationName,
						strconv.Itoa(batches[migrationName]),
					}
				} else {
					table.Cols = []string{
						color.String("<red>No</>"),
						migrationName,
						"",
					}
				}
			}
			table.SetOutput(os.Stdout)
			table.Println()
			return nil
		},
	}

	return command
}

type ResetCommand struct {
}

func (cmd *ResetCommand) Handle(app goofy.IApplication) *cobra.Command {
	command := &cobra.Command{
		Use:   "migrate:reset",
		Short: "重置迁移",
		RunE: func(c *cobra.Command, args []string) error {
			var manager Factory
			if err := app.Resolve(&manager); err != nil {
				return err
			}
			if err := runReset(manager); err != nil {
				color.Errorln(err)
				return err
			}
			return nil
		},
	}

	return command
}

type MakeCommand struct {
	table  string
	create string
}

func (cmd *MakeCommand) Handle(app goofy.IApplication) *cobra.Command {
	command := &cobra.Command{
		Use:        "make:migration {name : 迁移名称}",
		ArgAliases: []string{"name"},
		Short:      "创建迁移文件",
		RunE: func(c *cobra.Command, args []string) error {
			var prompt *survey.Input
			var name string
			var isCreate bool
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

			if err := os.MkdirAll(path.Join(app.Workspace(), "databases", "migration"), os.ModePerm); err != nil {
				color.Errorln(err)
			}

			name = strutil.ToSnake(name)

			if cmd.table == "" && cmd.create != "" {
				cmd.table = cmd.create
				isCreate = true
			}

			if cmd.table == "" {
				cmd.table, isCreate = NewTableGuesser().Guess(name)
			}

			generatePath := path.Join(app.Workspace(), "databases", "migration")

			if err := writeMigration(name, cmd.table, generatePath, isCreate); err != nil {
				color.Errorln(err)
				return err
			}

			color.Infoln("执行完毕")

			return nil
		},
	}

	command.PersistentFlags().StringVarP(&cmd.table, "table", "t", "", "The table to migrate")
	command.PersistentFlags().StringVarP(&cmd.create, "create", "c", "", "The table to be created")

	return command
}

func getStub(isCreate bool) string {
	if isCreate {
		return createStub
	}

	return blankStub
}

func writeMigration(name, table, generatePath string, isCreate bool) error {
	stub := getStub(isCreate)

	filePath := path.Join(generatePath, strings.ToLower(name)+".go")

	if fsutil.FileExists(filePath) {
		color.Infoln("该迁移文件已存在，是否覆盖？")
		if !interact.AnswerIsYes(false) {
			return nil
		}
	}

	stubString, err := populateStub(stub, table)

	if err != nil {
		return err
	}

	if f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666); err == nil {
		f.WriteString(stubString)
	} else {
		return err
	}

	return nil
}

func populateStub(stub, table string) (string, error) {
	var templateBuffer bytes.Buffer
	tpl, err := template.New("migration").Parse(stub)
	if err != nil {
		return templateBuffer.String(), err
	}

	if err := tpl.ExecuteTemplate(&templateBuffer, "migration", map[string]interface{}{
		"StructName": strings.ToUpper(strutil.RandomChars(6)),
		"TableName":  table,
		"Timestamp":  carbon.Now().ToTimestamp(),
	}); err != nil {
		return templateBuffer.String(), err
	}

	return templateBuffer.String(), nil
}

func runPending(migrations []File, step int, db *gorm.DB) error {
	repository := &Model{
		DB: db,
	}

	if len(migrations) == 0 {
		err := errors.New("nothing to migrate")
		return err
	}

	batch, err := repository.GetNextBatchNumber()

	if err != nil {
		return err
	}

	p := progress.Bar(len(migrations))
	p.Start()

	for _, migrationFile := range migrations {

		if err := runUp(migrationFile, batch, db); err != nil {
			return err
		}

		if step > 0 {
			batch++
		}
		p.Advance()
	}

	p.Finish()

	return nil
}

func runUp(file File, batch int, db *gorm.DB) error {
	repository := &Model{
		DB: db,
	}
	name := getMigrationName(file)

	color.Infoln("Migrating:", name)

	if err := file.Up(db); err != nil {
		return err
	}
	if err := repository.Log(name, batch); err != nil {
		return err
	}

	color.Infoln("Migrated:", name)

	return nil
}

func runRollback(step int, manager Factory) error {
	repository := NewDBMigration(manager.Connection())
	dbMigrations, err := getMigrationsForRollback(step, repository)

	if err != nil {
		return err
	}

	return rollbackMigrations(dbMigrations, manager.Connection())
}

func runReset(manager Factory) error {
	var migrations []*Model
	var err error
	migrations, err = NewDBMigration(manager.Connection()).GetRan()
	if err != nil {
		return err
	}

	if len(migrations) == 0 || len(getMigrationFiles()) == 0 {
		err = errors.New("nothing to rollback")
		return err
	}

	return rollbackMigrations(migrations, manager.Connection())
}

func runMigrate(step int, manager Factory) error {
	var ran []*Model
	var err error

	if ran, err = NewDBMigration(manager.Connection()).GetRan(); err != nil {
		return err
	}

	migrateFiles := getMigrationFiles()

	pendingMigrateFiles := getPendingMigrations(migrateFiles, ran)

	sortFileMigrations(pendingMigrateFiles)

	if err := runPending(pendingMigrateFiles, step, manager.Connection()); err != nil {
		return err
	}
	return nil
}

func getMigrationFiles() []File {
	return migrationFiles
}

func getMigrationName(migrateFile File) string {
	migrationNames := strings.Split(reflect.TypeOf(migrateFile).String(), ".")
	return strutil.ToSnake(migrationNames[len(migrationNames)-1])
}

func sortFileMigrations(files []File) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].MigrateTimestamp() < files[j].MigrateTimestamp()
	})
}

func getPendingMigrations(files []File, ran []*Model) []File {
	var pendingMigrations []File
	ranNameCollection := collection.NewObjPointCollection(ran).Pluck("Migration")
	for _, migrateFile := range files {
		if !ranNameCollection.Contains(getMigrationName(migrateFile)) {
			pendingMigrations = append(pendingMigrations, migrateFile)
		}
	}
	return pendingMigrations
}

func getMigrationsForRollback(step int, repository *Model) ([]*Model, error) {
	var dbMigrates []*Model
	var err error
	if step > 0 {
		dbMigrates, err = repository.GetMigrations(step)
	} else {
		dbMigrates, err = repository.GetLast()
	}
	return dbMigrates, err
}

func rollbackMigrations(migrations []*Model, db *gorm.DB) error {
	files := getMigrationFiles()

	existsFileMigrates := func(dbMigrate *Model) (File, bool) {
		for _, migrateFile := range files {
			migrationNames := strings.Split(reflect.TypeOf(migrateFile).String(), ".")
			migrationName := strutil.ToSnake(migrationNames[len(migrationNames)-1])
			if dbMigrate.Migration == migrationName {
				return migrateFile, true
			}
		}
		return nil, false
	}

	for _, migration := range migrations {
		file, exists := existsFileMigrates(migration)

		if !exists {
			color.Warnln("Migration not found:", migration.Migration)
			continue
		}

		if err := runDown(file, migration, db); err != nil {
			return err
		}
	}

	return nil
}

func runDown(file File, migration *Model, db *gorm.DB) (err error) {
	repository := &Model{
		DB: db,
	}

	name := getMigrationName(file)

	color.Infoln("Rolling back:", name)

	if err := file.Down(db); err != nil {
		return err
	}

	if err := repository.Delete(migration); err != nil {
		return err
	}

	color.Infoln("Rolled back:", name)

	return nil
}
