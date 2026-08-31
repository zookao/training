package admin

import (
	"time"

	"training/global"
	adminModel "training/model/admin"
	userModel "training/model/user"
)

// RecentLearnItem 最近学习记录
type RecentLearnItem struct {
	UserID      uint      `json:"userId"`
	Username    string    `json:"username"`
	Nickname    string    `json:"nickname"`
	CourseID    uint      `json:"courseId"`
	CourseTitle string    `json:"courseTitle"`
	Percent     int       `json:"percent"`
	Completed   bool      `json:"completed"`
	LastAt      time.Time `json:"lastAt"`
}

// DashboardRes 首页统计
type DashboardRes struct {
	ClassCount     int64             `json:"classCount"`     // 班级总数
	ActiveClass    int64             `json:"activeClass"`    // 启用班级
	CourseCount    int64             `json:"courseCount"`    // 课程总数
	ActiveCourse   int64             `json:"activeCourse"`   // 启用课程
	StudentCount   int64             `json:"studentCount"`   // 学员总数
	ActiveStudent  int64             `json:"activeStudent"`  // 启用学员
	RecordCount    int64             `json:"recordCount"`    // 学习记录数
	CompletedCount int64             `json:"completedCount"` // 已完成视频数
	AvgProgress    int               `json:"avgProgress"`    // 平均学习进度
	LearnerCount   int64             `json:"learnerCount"`   // 学习人数（有过学习行为的学员数）
	CompletionRate int               `json:"completionRate"` // 完成率（已完成/总记录，%）
	TodayActive    int64             `json:"todayActive"`    // 今日活跃学员数
	RecentRecords  []RecentLearnItem `json:"recentRecords"`  // 最近学习
}

// Dashboard 首页统计
//
// 统计口径：使用 Model(...) 而非 Table(...)，以自动排除软删除记录；
// 学习记录额外排除其关联学员/课程已被软删除的孤儿记录。
func Dashboard() (*DashboardRes, error) {
	res := &DashboardRes{RecentRecords: []RecentLearnItem{}}

	// 班级/课程/学员总数：Model(...) 自动带 deleted_at IS NULL
	global.DB.Model(&adminModel.Class{}).Count(&res.ClassCount)
	global.DB.Model(&adminModel.Class{}).Where("status = 1").Count(&res.ActiveClass)
	global.DB.Model(&adminModel.Course{}).Count(&res.CourseCount)
	global.DB.Model(&adminModel.Course{}).Where("status = 1").Count(&res.ActiveCourse)
	global.DB.Model(&userModel.User{}).Count(&res.StudentCount)
	global.DB.Model(&userModel.User{}).Where("status = 1").Count(&res.ActiveStudent)

	// 学习记录：video_records 无软删除，但需排除关联学员/课程已软删除的孤儿记录
	global.DB.Table("video_records AS vr").
		Joins("INNER JOIN users u ON u.id = vr.user_id AND u.deleted_at IS NULL").
		Joins("INNER JOIN courses c ON c.id = vr.course_id AND c.deleted_at IS NULL").
		Count(&res.RecordCount)

	global.DB.Table("video_records AS vr").
		Joins("INNER JOIN users u ON u.id = vr.user_id AND u.deleted_at IS NULL").
		Joins("INNER JOIN courses c ON c.id = vr.course_id AND c.deleted_at IS NULL").
		Where("vr.completed = ?", true).Count(&res.CompletedCount)

	// 平均进度
	today := time.Now().Format("2006-01-02")
	var avgPercent *float64
	global.DB.Table("video_records AS vr").
		Joins("INNER JOIN users u ON u.id = vr.user_id AND u.deleted_at IS NULL").
		Joins("INNER JOIN courses c ON c.id = vr.course_id AND c.deleted_at IS NULL").
		Select("COALESCE(AVG(vr.percent), 0)").Row().Scan(&avgPercent)
	if avgPercent != nil {
		res.AvgProgress = int(*avgPercent)
	}

	// 学习人数（有学习记录的去重学员数）
	global.DB.Table("video_records AS vr").
		Joins("INNER JOIN users u ON u.id = vr.user_id AND u.deleted_at IS NULL").
		Joins("INNER JOIN courses c ON c.id = vr.course_id AND c.deleted_at IS NULL").
		Select("COUNT(DISTINCT vr.user_id)").Row().Scan(&res.LearnerCount)

	// 完成率（已完成视频数 / 学习记录数）
	if res.RecordCount > 0 {
		res.CompletionRate = int(float64(res.CompletedCount) * 100 / float64(res.RecordCount))
	}

	// 今日活跃学员数（今天有学习行为的去重学员数，按 user_id 去重，避免同一学员看多个视频被重复计数）
	global.DB.Table("video_records AS vr").
		Joins("INNER JOIN users u ON u.id = vr.user_id AND u.deleted_at IS NULL").
		Joins("INNER JOIN courses c ON c.id = vr.course_id AND c.deleted_at IS NULL").
		Where("vr.last_at >= ?", today).
		Select("COUNT(DISTINCT vr.user_id)").Row().Scan(&res.TodayActive)

	// 最近 10 条学习记录（关联学员与课程标题，排除已软删除的学员/课程）
	type joinRow struct {
		UserID      uint      `gorm:"column:userId"`
		Username    string    `gorm:"column:username"`
		Nickname    string    `gorm:"column:nickname"`
		CourseID    uint      `gorm:"column:courseId"`
		CourseTitle string    `gorm:"column:courseTitle"`
		Percent     int       `gorm:"column:percent"`
		Completed   bool      `gorm:"column:completed"`
		LastAt      time.Time `gorm:"column:lastAt"`
	}
	var rows []joinRow
	err := global.DB.Table("video_records AS vr").
		Select("vr.user_id AS userId, u.username AS username, u.nickname AS nickname, "+
			"vr.course_id AS courseId, c.title AS courseTitle, "+
			"vr.percent AS percent, vr.completed AS completed, vr.last_at AS lastAt").
		Joins("INNER JOIN users u ON u.id = vr.user_id AND u.deleted_at IS NULL").
		Joins("INNER JOIN courses c ON c.id = vr.course_id AND c.deleted_at IS NULL").
		Order("vr.last_at DESC").
		Limit(4).
		Scan(&rows).Error
	if err == nil {
		for _, r := range rows {
			res.RecentRecords = append(res.RecentRecords, RecentLearnItem{
				UserID:      r.UserID,
				Username:    r.Username,
				Nickname:    r.Nickname,
				CourseID:    r.CourseID,
				CourseTitle: r.CourseTitle,
				Percent:     r.Percent,
				Completed:   r.Completed,
				LastAt:      r.LastAt,
			})
		}
	}
	return res, nil
}
