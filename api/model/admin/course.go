package admin

import (
	"time"

	"gorm.io/gorm"
)

// Course 课程
type Course struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"size:100;not null" json:"title"`
	Cover       string         `gorm:"size:255" json:"cover"` // 封面图 URL
	Description string         `gorm:"size:1000" json:"description"`
	Sort        int            `gorm:"default:0" json:"sort"`
	Status      int8           `gorm:"default:1" json:"status"` // 1启用 0禁用
	Videos      []Video        `gorm:"foreignKey:CourseID" json:"videos,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Course) TableName() string { return "courses" }

// Video 视频
type Video struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	CourseID             uint           `gorm:"index;not null" json:"courseId"`
	URL                  string         `gorm:"size:255;not null" json:"url"`
	Thumbnail            string         `gorm:"size:255" json:"thumbnail"`
	Courseware            string         `gorm:"size:255" json:"courseware"`        // 课件URL（PPTX）
	CoursewarePageCount   int            `gorm:"default:0" json:"coursewarePageCount"` // 课件幻灯片页数
	CoursewarePages       string         `gorm:"type:text" json:"coursewarePages"`  // 课件每页时长 JSON: [10,60,300]（秒）
	CoursewarePDF         string         `gorm:"size:255" json:"coursewarePdf"`     // 课件 PDF URL（由 PPTX 转换生成，用于在线预览）
	Title                string         `gorm:"size:100" json:"title"`
	Description          string         `gorm:"size:1000" json:"description"`
	Sort                 int            `gorm:"default:0" json:"sort"`
	Duration             int            `gorm:"default:0" json:"duration"` // 视频时长（秒，服务端解析）
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Video) TableName() string { return "videos" }
