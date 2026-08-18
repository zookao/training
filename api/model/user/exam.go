package user

import (
	"time"

	"gorm.io/gorm"
)

// TestpaperRecord 考试记录
type TestpaperRecord struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TestpaperID uint           `gorm:"not null;index" json:"testpaperId"`
	UserID      uint           `gorm:"not null;index" json:"userId"`
	CourseID    uint           `gorm:"not null;index" json:"courseId"`
	Score       int            `gorm:"default:0" json:"score"`  // 得分
	Passed      bool           `gorm:"default:false" json:"passed"` // 是否及格
	Answers     string         `gorm:"type:text" json:"answers"` // JSON: [{"questionId":1,"userAnswer":["A"],"correct":true,"score":10}]
	DraftAnswers string        `gorm:"type:text" json:"draftAnswers"` // 草稿答案 JSON: [{"questionId":1,"userAnswer":["A"]}]（交卷前自动保存，用于断点续考）
	StartedAt   time.Time      `json:"startedAt"`
	SubmittedAt *time.Time     `json:"submittedAt"`
	Duration    int            `gorm:"default:0" json:"duration"` // 实际用时（秒）
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (TestpaperRecord) TableName() string { return "testpaper_records" }
