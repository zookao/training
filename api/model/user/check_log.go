package user

import "time"

// CheckLog 滑动校验通过日志（每通过一次校验写一条）
type CheckLog struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	UserID            uint      `gorm:"index;not null" json:"userId"`
	VideoID           uint      `gorm:"index;not null" json:"videoId"`
	CourseID          uint      `gorm:"index" json:"courseId"`
	ClassID           uint      `gorm:"index" json:"classId"`
	CheckPosition     int       `json:"checkPosition"`     // 触发校验的视频位置（秒）
	MaxPosition       int       `json:"maxPosition"`       // 通过时的最远播放位置（秒）
	NextCheckPosition int       `json:"nextCheckPosition"` // 通过后安排的下一个校验点（秒）
	CreatedAt         time.Time `json:"createdAt"`         // 通过时间
}

// TableName 表名
func (CheckLog) TableName() string { return "check_logs" }
