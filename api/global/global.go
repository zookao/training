package global

import (
	"log"

	"training/config"

	"gorm.io/gorm"
)

var (
	Config *config.Config
	DB     *gorm.DB
)

func init() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
}
