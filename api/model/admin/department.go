package admin

import (
	"time"

	"gorm.io/gorm"
)

// Department 院系
type Department struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	Sort        int            `gorm:"default:0" json:"sort"`
	Status      int8           `gorm:"default:1;comment:1启用0禁用" json:"status"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Department) TableName() string { return "departments" }
