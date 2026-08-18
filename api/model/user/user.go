package user

import (
	"time"

	"gorm.io/gorm"
)

// User 学员
type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Username    string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password    string         `gorm:"size:100;not null" json:"-"`
	StudentNo   string         `gorm:"size:50" json:"studentNo"`
	DepartmentID uint          `gorm:"index" json:"departmentId"`
	Nickname    string         `gorm:"size:50" json:"nickname"`
	Avatar      string         `gorm:"size:255" json:"avatar"`
	Email       string         `gorm:"size:100" json:"email"`
	Phone       string         `gorm:"size:20" json:"phone"`
	Status      int8           `gorm:"default:1;comment:1启用0禁用" json:"status"`
	LastLoginIP string         `gorm:"size:50" json:"lastLoginIp"`
	LastLoginAt *time.Time     `json:"lastLoginAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (User) TableName() string { return "users" }
