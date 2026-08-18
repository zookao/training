package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"training/global"
	adminModel "training/model/admin"
	userModel "training/model/user"
)

// ExamQuestionVO 学员端试题（不含答案）
type ExamQuestionVO struct {
	ID      uint   `json:"id"`
	Type    int8   `json:"type"`
	Title   string `json:"title"`
	Options string `json:"options"`
	Score   int    `json:"score"` // 该题在试卷中的分值
	Sort    int    `json:"sort"`
}

// ExamTestpaperVO 学员端试卷（含试题，不含答案）
type ExamTestpaperVO struct {
	ID           uint                 `json:"id"`
	CourseID     uint                 `json:"courseId"`
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Type         int8                 `json:"type"`
	TotalScore   int                  `json:"totalScore"`
	PassScore    int                  `json:"passScore"`
	Duration     int                  `json:"duration"`    // 考试时长（分钟）
	RemainSec    int                  `json:"remainSec"`   // 剩余秒数
	Questions    []ExamQuestionVO     `json:"questions"`
	DraftAnswers []SubmitAnswerItem   `json:"draftAnswers"` // 草稿答案（断点续考用，新建时为空）
}

// SubmitAnswerItem 提交的单题答案
type SubmitAnswerItem struct {
	QuestionID uint     `json:"questionId"`
	UserAnswer []string `json:"userAnswer"`
}

// SubmitReq 考试提交请求
type SubmitReq struct {
	Answers []SubmitAnswerItem `json:"answers"`
}

// AnswerDetail 答题明细
type AnswerDetail struct {
	QuestionID    uint     `json:"questionId"`
	Title         string   `json:"title"`
	Type          int8     `json:"type"`
	UserAnswer    []string `json:"userAnswer"`
	CorrectAnswer []string `json:"correctAnswer"`
	Correct       bool     `json:"correct"`
	Score         int      `json:"score"`    // 该题得分
	MaxScore      int      `json:"maxScore"` // 该题满分
}

// SubmitResult 提交结果
type SubmitResult struct {
	Score    int            `json:"score"`
	Passed   bool           `json:"passed"`
	Total    int            `json:"total"`
	PassLine int            `json:"passLine"`
	Details  []AnswerDetail `json:"details"`
}

// GetExam 获取试卷（学员端，不含正确答案）
// 进入考试时调用：有未完成记录则恢复（剩余时间），否则新建记录
func GetExam(userID, testpaperID uint) (*ExamTestpaperVO, error) {
	var tp adminModel.Testpaper
	if err := global.DB.First(&tp, testpaperID).Error; err != nil {
		return nil, errors.New("试卷不存在")
	}
	if tp.Status != 1 {
		return nil, errors.New("试卷未启用")
	}

	// 校验课程访问权限
	if !userHasCourseAccess(userID, tp.CourseID) {
		return nil, errors.New("无权访问该课程")
	}

	// 课程完成后才能考试 → 校验课程是否已完成
	if tp.Type == 2 {
		videos := courseVideos(tp.CourseID)
		completed, _ := courseProgress(userID, tp.CourseID, len(videos))
		if completed < len(videos) {
			return nil, errors.New("课程未完成，暂不能考试")
		}
	}

	// 使用 MySQL GET_LOCK 串行化同一用户同一试卷的并发进入请求，
	// 防止 check-then-create 竞态导致重复创建未提交考试记录
	lockKey := fmt.Sprintf("exam_enter:%d:%d", userID, testpaperID)
	global.DB.Exec("SELECT GET_LOCK(?, 10)", lockKey)
	defer global.DB.Exec("SELECT RELEASE_LOCK(?)", lockKey)

	// 检查是否有未完成的考试记录（submitted_at IS NULL）
	var record userModel.TestpaperRecord
	now := time.Now()
	totalSec := tp.Duration * 60
	hasUnfinished := false

	if err := global.DB.Where("testpaper_id = ? AND user_id = ? AND submitted_at IS NULL", testpaperID, userID).
		First(&record).Error; err == nil {
		// 有未完成记录
		hasUnfinished = true
		elapsed := int(now.Sub(record.StartedAt).Seconds())
		remainSec := totalSec - elapsed
		if remainSec <= 0 {
			// 超时自动交卷：用草稿答案判分（已答内容给分，未答算错）
			gradeAndSave(&tp, &record, parseDraftMap(record.DraftAnswers), now, elapsed)
			return nil, errors.New("考试已超时，已自动交卷")
		}
	} else {
		// 没有未完成记录，新建一条
		record = userModel.TestpaperRecord{
			TestpaperID: testpaperID,
			UserID:      userID,
			CourseID:    tp.CourseID,
			StartedAt:   now,
		}
		if err := global.DB.Create(&record).Error; err != nil {
			return nil, errors.New("创建考试记录失败")
		}
	}

	// 获取试卷试题
	var tqs []adminModel.TestpaperQuestion
	if err := global.DB.Where("testpaper_id = ?", testpaperID).Order("sort asc, id asc").Find(&tqs).Error; err != nil {
		return nil, err
	}

	questions := make([]ExamQuestionVO, 0, len(tqs))
	for _, tq := range tqs {
		var q adminModel.Question
		if err := global.DB.First(&q, tq.QuestionID).Error; err != nil {
			continue
		}
		questions = append(questions, ExamQuestionVO{
			ID:      q.ID,
			Type:    q.Type,
			Title:   q.Title,
			Options: q.Options,
			Score:   tq.Score,
			Sort:    tq.Sort,
		})
	}

	// 计算剩余时间
	remainSec := totalSec
	if hasUnfinished {
		elapsed := int(now.Sub(record.StartedAt).Seconds())
		remainSec = totalSec - elapsed
	}

	// 恢复草稿答案（断点续考）：未完成记录中保存的草稿回带前端
	draftAnswers := make([]SubmitAnswerItem, 0)
	if hasUnfinished && record.DraftAnswers != "" {
		_ = json.Unmarshal([]byte(record.DraftAnswers), &draftAnswers)
	}

	return &ExamTestpaperVO{
		ID:           tp.ID,
		CourseID:     tp.CourseID,
		Name:         tp.Name,
		Description:  tp.Description,
		Type:         tp.Type,
		TotalScore:   tp.TotalScore,
		PassScore:    tp.PassScore,
		Duration:     tp.Duration,
		RemainSec:    remainSec,
		Questions:    questions,
		DraftAnswers: draftAnswers,
	}, nil
}

// SaveExamDraft 保存考试草稿（中途答案，用于断点续考）
// 仅更新未完成记录（submitted_at IS NULL）的 draft_answers 字段
func SaveExamDraft(userID, testpaperID uint, req SubmitReq) error {
	var record userModel.TestpaperRecord
	if err := global.DB.Where("testpaper_id = ? AND user_id = ? AND submitted_at IS NULL", testpaperID, userID).
		First(&record).Error; err != nil {
		return errors.New("无未完成考试，无需保存草稿")
	}
	draftJSON, err := json.Marshal(req.Answers)
	if err != nil {
		return errors.New("草稿序列化失败")
	}
	return global.DB.Model(&record).Update("draft_answers", string(draftJSON)).Error
}

// SubmitExam 提交考试
func SubmitExam(userID, testpaperID uint, req SubmitReq) (*SubmitResult, error) {
	var tp adminModel.Testpaper
	if err := global.DB.First(&tp, testpaperID).Error; err != nil {
		return nil, errors.New("试卷不存在")
	}

	// 查找未完成的考试记录
	var record userModel.TestpaperRecord
	if err := global.DB.Where("testpaper_id = ? AND user_id = ? AND submitted_at IS NULL", testpaperID, userID).
		First(&record).Error; err != nil {
		return nil, errors.New("考试已交卷或未开始")
	}

	now := time.Now()
	elapsed := int(now.Sub(record.StartedAt).Seconds())
	// 用时不超过考试最大时长（迟交/异常情况封顶）
	if maxSec := tp.Duration * 60; elapsed > maxSec {
		elapsed = maxSec
	}

	// 构建用户答案 map
	userAnsMap := make(map[uint][]string, len(req.Answers))
	for _, a := range req.Answers {
		userAnsMap[a.QuestionID] = a.UserAnswer
	}

	// 判分
	totalScore, passed, details := gradeExam(&tp, userAnsMap)

	// 更新考试记录（条件：仍未完成，避免与定时任务/并发重复判分）
	answersJSON, _ := json.Marshal(details)
	global.DB.Model(&userModel.TestpaperRecord{}).
		Where("id = ? AND submitted_at IS NULL", record.ID).
		Updates(map[string]interface{}{
			"score":        totalScore,
			"passed":       passed,
			"answers":      string(answersJSON),
			"submitted_at": now,
			"duration":     elapsed,
		})

	return &SubmitResult{
		Score:    totalScore,
		Passed:   passed,
		Total:    tp.TotalScore,
		PassLine: tp.PassScore,
		Details:  details,
	}, nil
}

// gradeExam 判分核心：根据用户答案计算得分、是否及格、答题明细（只读，不写库）
func gradeExam(tp *adminModel.Testpaper, userAnsMap map[uint][]string) (int, bool, []AnswerDetail) {
	var tqs []adminModel.TestpaperQuestion
	global.DB.Where("testpaper_id = ?", tp.ID).Order("sort asc, id asc").Find(&tqs)

	type qInfo struct {
		question adminModel.Question
		score    int
	}
	qMap := make(map[uint]qInfo, len(tqs))
	for _, tq := range tqs {
		var q adminModel.Question
		if err := global.DB.First(&q, tq.QuestionID).Error; err != nil {
			continue
		}
		qMap[tq.QuestionID] = qInfo{question: q, score: tq.Score}
	}

	details := make([]AnswerDetail, 0, len(tqs))
	totalScore := 0
	for _, tq := range tqs {
		info, ok := qMap[tq.QuestionID]
		if !ok {
			continue
		}
		var correctAns []string
		_ = json.Unmarshal([]byte(info.question.Answer), &correctAns)
		userAns := userAnsMap[tq.QuestionID]
		if userAns == nil {
			userAns = []string{}
		}
		correct := compareAnswers(userAns, correctAns)
		score := 0
		if correct {
			score = info.score
			totalScore += score
		}
		details = append(details, AnswerDetail{
			QuestionID:    tq.QuestionID,
			Title:         info.question.Title,
			Type:          info.question.Type,
			UserAnswer:    userAns,
			CorrectAnswer: correctAns,
			Correct:       correct,
			Score:         score,
			MaxScore:      info.score,
		})
	}
	passed := totalScore >= tp.PassScore
	return totalScore, passed, details
}

// parseDraftMap 解析草稿答案 JSON 为 map[questionID]→userAnswer
func parseDraftMap(draftJSON string) map[uint][]string {
	m := make(map[uint][]string)
	if draftJSON == "" {
		return m
	}
	var items []SubmitAnswerItem
	if err := json.Unmarshal([]byte(draftJSON), &items); err != nil {
		return m
	}
	for _, it := range items {
		m[it.QuestionID] = it.UserAnswer
	}
	return m
}

// gradeAndSave 判分并保存（超时自动交卷用）
func gradeAndSave(tp *adminModel.Testpaper, record *userModel.TestpaperRecord, userAnsMap map[uint][]string, submittedAt time.Time, elapsed int) {
	// 用时不超过考试最大时长（超时自动交卷时 elapsed 可能已超过考试时长）
	if maxSec := tp.Duration * 60; elapsed > maxSec {
		elapsed = maxSec
	}
	totalScore, passed, details := gradeExam(tp, userAnsMap)
	answersJSON, _ := json.Marshal(details)
	// 条件更新：仅当记录仍为未完成时才写入，避免与定时任务/并发提交重复判分
	global.DB.Model(&userModel.TestpaperRecord{}).
		Where("id = ? AND submitted_at IS NULL", record.ID).
		Updates(map[string]interface{}{
			"score":        totalScore,
			"passed":       passed,
			"answers":      string(answersJSON),
			"submitted_at": submittedAt,
			"duration":     elapsed,
		})
}

// GetExamRecords 获取学员的考试记录
func GetExamRecords(userID, courseID uint) ([]map[string]interface{}, error) {
	var records []userModel.TestpaperRecord
	if err := global.DB.Where("user_id = ? AND course_id = ?", userID, courseID).
		Order("created_at desc").Find(&records).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(records))
	for _, r := range records {
		var tp adminModel.Testpaper
		global.DB.First(&tp, r.TestpaperID)
		submittedAt := ""
		if r.SubmittedAt != nil {
			submittedAt = r.SubmittedAt.Format("2006-01-02 15:04:05")
		}
		// 用时不超过考试最大时长（兼容历史已存的超长用时）
		duration := r.Duration
		if maxSec := tp.Duration * 60; duration > maxSec {
			duration = maxSec
		}
		result = append(result, map[string]interface{}{
			"recordId":      r.ID,
			"testpaperId":   r.TestpaperID,
			"testpaperName": tp.Name,
			"score":         r.Score,
			"passed":        r.Passed,
			"duration":      duration,
			"submittedAt":   submittedAt,
		})
	}
	return result, nil
}

// GetCourseTestpapers 获取课程下所有试卷（含可用状态）
func GetCourseTestpapers(userID, courseID uint) ([]map[string]interface{}, error) {
	var testpapers []adminModel.Testpaper
	if err := global.DB.Where("course_id = ? AND status = 1", courseID).
		Order("sort asc, id desc").Find(&testpapers).Error; err != nil {
		return nil, err
	}

	// 检查课程是否已完成
	videos := courseVideos(courseID)
	completed, _ := courseProgress(userID, courseID, len(videos))
	courseCompleted := len(videos) > 0 && completed >= len(videos)

	// 查该用户该课程下所有考试记录，按试卷聚合：最高分 / 是否有已完成 / 是否有进行中
	var records []userModel.TestpaperRecord
	_ = global.DB.Where("user_id = ? AND course_id = ?", userID, courseID).Find(&records)
	type examStat struct {
		bestScore   int
		hasFinished bool
		bestPassed  bool
		inProgress  bool
	}
	statsMap := make(map[uint]*examStat, len(testpapers))
	for i := range records {
		r := &records[i]
		s, ok := statsMap[r.TestpaperID]
		if !ok {
			s = &examStat{}
			statsMap[r.TestpaperID] = s
		}
		if r.SubmittedAt != nil {
			s.hasFinished = true
			if r.Score > s.bestScore {
				s.bestScore = r.Score
				s.bestPassed = r.Passed
			}
		} else {
			s.inProgress = true
		}
	}

	result := make([]map[string]interface{}, 0, len(testpapers))
	for _, tp := range testpapers {
		available := tp.Type == 1 || courseCompleted
		item := map[string]interface{}{
			"id":          tp.ID,
			"name":        tp.Name,
			"description": tp.Description,
			"type":        tp.Type,
			"totalScore":  tp.TotalScore,
			"passScore":   tp.PassScore,
			"duration":    tp.Duration,
			"available":   available,
			"hasFinished": false,
			"bestScore":   0,
			"bestPassed":  false,
			"inProgress":  false,
		}
		if s, ok := statsMap[tp.ID]; ok {
			item["hasFinished"] = s.hasFinished
			item["bestScore"] = s.bestScore
			item["bestPassed"] = s.bestPassed
			item["inProgress"] = s.inProgress
		}
		result = append(result, item)
	}
	return result, nil
}

// compareAnswers 比较答案是否一致（忽略顺序）
func compareAnswers(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]bool, len(a))
	for _, v := range a {
		set[v] = true
	}
	for _, v := range b {
		if !set[v] {
			return false
		}
	}
	return true
}
