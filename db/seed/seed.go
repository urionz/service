package seed

import "gorm.io/gorm"

var seederFiles []Seeder

func Register(files ...Seeder) {
	seederFiles = append(seederFiles, files...)
}

type Seeder interface {
	Handle(db *gorm.DB) error
}
