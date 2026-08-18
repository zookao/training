package admin

import (
	"time"

	"gorm.io/gorm"
)

// Question 试题
type Question struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CourseID  uint           `gorm:"not null;index" json:"courseId"`
	Type      int8           `gorm:"not null;comment:1单选2多选3判断" json:"type"` // 1单选 2多选 3判断
	Title     string         `gorm:"type:text;not null" json:"title"`
	Options   string         `gorm:"type:text" json:"options"`  // JSON: [{"label":"A","content":"选项A"}]
	Answer    string         `gorm:"type:text" json:"answer"`   // JSON: ["A"] 或 ["A","C"]
	Analysis  string         `gorm:"type:text" json:"analysis"` // 解析
	Sort      int            `gorm:"default:0" json:"sort"`
	Status    int8           `gorm:"default:1" json:"status"` // 1启用 0禁用
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Question) TableName() string { return "questions" }

// Testpaper 试卷
type Testpaper struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CourseID    uint           `gorm:"not null;index" json:"courseId"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Type        int8           `gorm:"default:2;comment:1随时考2课程完成后考" json:"type"` // 1随时考 2课程完成后考
	TotalScore  int            `gorm:"default:100" json:"totalScore"`                      // 总分固定100
	PassScore   int            `gorm:"default:60" json:"passScore"`                        // 及格分
	Duration    int            `gorm:"default:60;comment:考试时长(分钟)" json:"duration"`    // 考试时长（分钟）
	Sort        int            `gorm:"default:0" json:"sort"`
	Status      int8           `gorm:"default:1" json:"status"` // 1启用 0禁用
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (Testpaper) TableName() string { return "testpapers" }

// TestpaperQuestion 试卷-试题关联（含分值）
type TestpaperQuestion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TestpaperID uint      `gorm:"not null;uniqueIndex:idx_tpq" json:"testpaperId"`
	QuestionID  uint      `gorm:"not null;uniqueIndex:idx_tpq" json:"questionId"`
	Score       int       `gorm:"default:0" json:"score"` // 该题分值
	Sort        int       `gorm:"default:0" json:"sort"`
	CreatedAt   time.Time `json:"createdAt"`
}

// TableName 表名
func (TestpaperQuestion) TableName() string { return "testpaper_questions" }
