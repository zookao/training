package admin

import (
	"time"

	"gorm.io/gorm"
)

// Menu 菜单/按钮（树形）
type Menu struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ParentID  uint           `gorm:"default:0;index" json:"parentId"`
	Name      string         `gorm:"size:50" json:"name"`
	Type      string         `gorm:"size:10;comment:M目录 C菜单 F按钮" json:"type"` // M / C / F
	Path      string         `gorm:"size:200" json:"path"`
	Component string         `gorm:"size:200" json:"component"`
	Redirect  string         `gorm:"size:200" json:"redirect"`
	Icon      string         `gorm:"size:50" json:"icon"`
	Hidden    bool           `gorm:"default:false" json:"hidden"`
	KeepAlive bool           `gorm:"default:false" json:"keepAlive"`
	Sort      int            `gorm:"default:0" json:"sort"`
	GuardType string         `gorm:"size:20;default:admin;index" json:"guardType"`
	Perms     string         `gorm:"size:100;comment:按钮权限码" json:"perms"` // 如 admin:add
	Children  []Menu         `gorm:"-" json:"children,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Menu) TableName() string { return "menus" }
