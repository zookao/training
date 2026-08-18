package admin

import (
	"time"

	"gorm.io/gorm"
)

// Api 接口权限点
type Api struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Group       string         `gorm:"size:50;index" json:"group"`
	Method      string         `gorm:"size:10" json:"method"`
	Path        string         `gorm:"size:200" json:"path"`
	Description string         `gorm:"size:255" json:"description"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Api) TableName() string { return "apis" }
