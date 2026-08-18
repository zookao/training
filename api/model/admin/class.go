package admin

import (
	"time"

	"training/model/user"

	"gorm.io/gorm"
)

// Class 班级
type Class struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Cover       string         `gorm:"size:255" json:"cover"` // 封面图 URL
	Description string         `gorm:"size:500" json:"description"`
	Sort        int            `gorm:"default:0" json:"sort"`
	Status      int8           `gorm:"default:1" json:"status"` // 1启用 0禁用
	Courses     []Course       `gorm:"many2many:class_courses;" json:"courses,omitempty"`
	Users       []user.User    `gorm:"many2many:class_users;" json:"users,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Class) TableName() string { return "classes" }

// ClassCourse 班级-课程 关联
type ClassCourse struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	ClassID  uint `gorm:"uniqueIndex:idx_class_course" json:"classId"`
	CourseID uint `gorm:"uniqueIndex:idx_class_course" json:"courseId"`
}

func (ClassCourse) TableName() string { return "class_courses" }

// ClassUser 班级-学员 关联
type ClassUser struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	ClassID uint `gorm:"uniqueIndex:idx_class_user" json:"classId"`
	UserID  uint `gorm:"uniqueIndex:idx_class_user" json:"userId"`
}

func (ClassUser) TableName() string { return "class_users" }
