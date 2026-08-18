package admin

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色
type Role struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Title     string         `gorm:"size:50" json:"title"`
	GuardType string         `gorm:"size:20;default:admin;index" json:"guardType"` // admin / user
	Sort      int            `gorm:"default:0" json:"sort"`
	Status    int8           `gorm:"default:1" json:"status"`
	Remark    string         `gorm:"size:255" json:"remark"`
	Menus     []Menu         `gorm:"many2many:role_menus;" json:"menus,omitempty"`
	Apis      []Api          `gorm:"many2many:role_apis;" json:"apis,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Role) TableName() string { return "roles" }

// AdminRole 管理员-角色 关联
type AdminRole struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	AdminID uint   `gorm:"uniqueIndex:idx_admin_role" json:"adminId"`
	RoleID  uint   `gorm:"uniqueIndex:idx_admin_role" json:"roleId"`
}

func (AdminRole) TableName() string { return "admin_roles" }

// UserRole 学员-角色 关联
type UserRole struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"uniqueIndex:idx_user_role" json:"userId"`
	RoleID uint `gorm:"uniqueIndex:idx_user_role" json:"roleId"`
}

func (UserRole) TableName() string { return "user_roles" }

// RoleMenu 角色-菜单 关联
type RoleMenu struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	RoleID uint `gorm:"uniqueIndex:idx_role_menu" json:"roleId"`
	MenuID uint `gorm:"uniqueIndex:idx_role_menu" json:"menuId"`
}

func (RoleMenu) TableName() string { return "role_menus" }

// RoleApi 角色-接口 关联
type RoleApi struct {
	ID    uint `gorm:"primaryKey" json:"id"`
	RoleID uint `gorm:"uniqueIndex:idx_role_api" json:"roleId"`
	ApiID uint `gorm:"uniqueIndex:idx_role_api" json:"apiId"`
}

func (RoleApi) TableName() string { return "role_apis" }
