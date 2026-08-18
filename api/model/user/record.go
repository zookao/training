package user

import (
	"time"
)

// VideoRecord 视频学习进度记录（每学员每视频一条）
type VideoRecord struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"uniqueIndex:idx_user_video;not null" json:"userId"`
	VideoID     uint       `gorm:"uniqueIndex:idx_user_video;not null" json:"videoId"`
	CourseID    uint       `gorm:"index;not null" json:"courseId"`
	ClassID     uint       `gorm:"index" json:"classId"` // 学习来源班级（可为 0）
	Position    int        `gorm:"default:0" json:"position"`       // 当前播放秒
	MaxPosition int        `gorm:"default:0" json:"maxPosition"`    // 最远到达秒（用于进度计算）
	Duration    int        `gorm:"default:0" json:"duration"`       // 视频总秒（前端上报）
	Completed   bool       `gorm:"default:false" json:"completed"`  // 是否学完
	Percent     int        `gorm:"default:0" json:"percent"`        // 进度百分比 0-100
	NextCheckPosition int  `gorm:"default:0" json:"nextCheckPosition"` // 下次校验触发的视频位置（秒），0=不校验
	CheckPending      bool `gorm:"default:false" json:"checkPending"` // 是否有待完成的滑动校验
	LastAt      time.Time  `json:"lastAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// TableName 表名
func (VideoRecord) TableName() string { return "video_records" }
