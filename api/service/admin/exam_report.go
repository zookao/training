package admin

import (
	"strings"

	"training/global"
	adminModel "training/model/admin"
	userModel "training/model/user"
)

// ExamReportItem 考试报告单条
type ExamReportItem struct {
	RecordID    uint   `json:"recordId"`
	TestpaperID uint   `json:"testpaperId"`
	TestpaperName string `json:"testpaperName"`
	UserID      uint   `json:"userId"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	StudentNo   string `json:"studentNo"`
	Score       int    `json:"score"`
	Passed      bool   `json:"passed"`
	Duration    int    `json:"duration"` // 秒
	StartedAt   string `json:"startedAt"`
}

// ExamReport 考试报告（keyword 按试卷名称/账号/姓名/学号模糊搜索）
func ExamReport(courseID uint, keyword string) ([]ExamReportItem, error) {
	// 查所有试卷
	var testpapers []adminModel.Testpaper
	if err := global.DB.Where("course_id = ?", courseID).Find(&testpapers).Error; err != nil {
		return nil, err
	}
	tpMap := make(map[uint]adminModel.Testpaper, len(testpapers))
	tpIDs := make([]uint, 0, len(testpapers))
	for _, tp := range testpapers {
		tpMap[tp.ID] = tp
		tpIDs = append(tpIDs, tp.ID)
	}
	if len(tpIDs) == 0 {
		return []ExamReportItem{}, nil
	}

	// 查所有考试记录
	var records []userModel.TestpaperRecord
	if err := global.DB.Where("testpaper_id IN ?", tpIDs).Order("submitted_at desc").Find(&records).Error; err != nil {
		return nil, err
	}

	// 收集用户ID
	userIDs := make([]uint, 0, len(records))
	for _, r := range records {
		userIDs = append(userIDs, r.UserID)
	}
	userMap := make(map[uint]userModel.User)
	if len(userIDs) > 0 {
		var users []userModel.User
		global.DB.Where("id IN ?", userIDs).Find(&users)
		for _, u := range users {
			userMap[u.ID] = u
		}
	}

	result := make([]ExamReportItem, 0, len(records))
	for _, r := range records {
		u := userMap[r.UserID]
		tp := tpMap[r.TestpaperID]
		// 用时不超过考试最大时长（兼容历史已存的超长用时）
		duration := r.Duration
		if maxSec := tp.Duration * 60; duration > maxSec {
			duration = maxSec
		}
		result = append(result, ExamReportItem{
			RecordID:      r.ID,
			TestpaperID:   r.TestpaperID,
			TestpaperName: tp.Name,
			UserID:        r.UserID,
			Username:      u.Username,
			Nickname:      u.Nickname,
			StudentNo:     u.StudentNo,
			Score:         r.Score,
			Passed:        r.Passed,
			Duration:      duration,
			StartedAt:     r.StartedAt.Format("2006-01-02 15:04:05"),
		})
	}
	// 按试卷名称/账号/姓名/学号模糊搜索
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		kw := strings.ToLower(keyword)
		filtered := make([]ExamReportItem, 0, len(result))
		for _, item := range result {
			if strings.Contains(strings.ToLower(item.TestpaperName), kw) ||
				strings.Contains(strings.ToLower(item.Username), kw) ||
				strings.Contains(strings.ToLower(item.Nickname), kw) ||
				strings.Contains(strings.ToLower(item.StudentNo), kw) {
				filtered = append(filtered, item)
			}
		}
		result = filtered
	}
	return result, nil
}

// ExamRecordDetail 考试记录详情（含答题明细）
func ExamRecordDetail(recordID uint) (map[string]interface{}, error) {
	var record userModel.TestpaperRecord
	if err := global.DB.First(&record, recordID).Error; err != nil {
		return nil, nil
	}
	var tp adminModel.Testpaper
	global.DB.First(&tp, record.TestpaperID)
	var u userModel.User
	global.DB.First(&u, record.UserID)

	return map[string]interface{}{
		"record":      record,
		"testpaper":   tp,
		"user":        u,
	}, nil
}
