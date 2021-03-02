package migrate

import "gorm.io/gorm"

var migrationFiles []File

type File interface {
	MigrateTimestamp() int
	TableName() string
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
}

func Register(migrateFile ...File) {
	migrationFiles = append(migrationFiles, migrateFile...)
}
