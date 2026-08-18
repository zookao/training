package user

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"training/global"
	adminModel "training/model/admin"
	userModel "training/model/user"
)

// examSchedulerTimeout 单次清理任务的整体超时：DB 慢或网络抖动时避免 goroutine 无限阻塞
const examSchedulerTimeout = 30 * time.Second

// StartExamScheduler 启动考试超时清理定时任务（后台协程，每 1 分钟执行一次）
// 把已超时但仍未交卷（submitted_at IS NULL）的考试记录用草稿答案判分并标记完成。
// 需在数据库初始化完成后调用。
func StartExamScheduler() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		log.Printf("[ExamScheduler] 已启动，每 1 分钟清理一次超时考试记录")
		// 启动时先执行一次，避免重启后第一批超时记录要等满 1 分钟才被清理
		FinalizeTimedOutExams()
		for range ticker.C {
			FinalizeTimedOutExams()
		}
	}()
}

// FinalizeTimedOutExams 清理超时未交卷的考试记录：用草稿答案判分并标记完成。
// 草稿答案代表学员断点续考前最后保存的作答，超时后据此给分（未答算错）；
// 无草稿则得 0 分。返回本次处理的记录数。
func FinalizeTimedOutExams() (int, error) {
	// 整体超时：DB 慢或网络抖动时避免 goroutine 无限阻塞，超时后等下一轮 tick 重试
	ctx, cancel := context.WithTimeout(context.Background(), examSchedulerTimeout)
	defer cancel()

	// 1. 查未完成记录（submitted_at IS NULL），加 LIMIT 防止异常堆积时一次性加载过多
	var records []userModel.TestpaperRecord
	if err := global.DB.WithContext(ctx).Where("submitted_at IS NULL").Limit(500).Find(&records).Error; err != nil {
		log.Printf("[ExamScheduler] 查询未完成记录失败: %v", err)
		return 0, err
	}
	if len(records) == 0 {
		return 0, nil
	}

	// 2. 批量加载相关试卷，避免 N+1 查询
	tpIDs := make([]uint, 0, len(records))
	for _, r := range records {
		tpIDs = append(tpIDs, r.TestpaperID)
	}
	var tps []adminModel.Testpaper
	if err := global.DB.WithContext(ctx).Where("id IN ?", tpIDs).Find(&tps).Error; err != nil {
		log.Printf("[ExamScheduler] 查询试卷失败: %v", err)
		return 0, err
	}
	tpMap := make(map[uint]adminModel.Testpaper, len(tps))
	for _, tp := range tps {
		tpMap[tp.ID] = tp
	}

	// 3. 逐条判断是否超时，超时则判分
	now := time.Now()
	finalized := 0
	for i := range records {
		r := &records[i]
		tp, ok := tpMap[r.TestpaperID]
		if !ok {
			// 试卷已被删除：直接把记录标记完成（0 分），避免一直挂着
			markRecordFinished(ctx, r.ID, 0, false, "[]", now, int(now.Sub(r.StartedAt).Seconds()))
			finalized++
			continue
		}
		totalSec := tp.Duration * 60
		elapsed := int(now.Sub(r.StartedAt).Seconds())
		if elapsed < totalSec {
			continue // 未超时，跳过
		}
		// 用时封顶为考试最大时长（记录可能挂了很久，elapsed 远超考试时长）
		if elapsed > totalSec {
			elapsed = totalSec
		}
		// 超时：用草稿答案判分（已答内容给分，未答算错）
		draftMap := parseDraftMap(r.DraftAnswers)
		score, passed, details := gradeExam(&tp, draftMap)
		detailsJSON, _ := json.Marshal(details)
		markRecordFinished(ctx, r.ID, score, passed, string(detailsJSON), now, elapsed)
		finalized++
	}
	if finalized > 0 {
		log.Printf("[ExamScheduler] 本次清理超时考试记录 %d 条", finalized)
	}
	return finalized, nil
}

// markRecordFinished 条件更新：仅当记录仍为未完成（submitted_at IS NULL）时才标记完成。
// 避免与 GetExam 超时分支或 SubmitExam 并发导致重复判分（条件不命中时影响行数为 0）。
func markRecordFinished(ctx context.Context, recordID uint, score int, passed bool, answersJSON string, submittedAt time.Time, elapsed int) {
	if err := global.DB.WithContext(ctx).Model(&userModel.TestpaperRecord{}).
		Where("id = ? AND submitted_at IS NULL", recordID).
		Updates(map[string]interface{}{
			"score":        score,
			"passed":       passed,
			"answers":      answersJSON,
			"submitted_at": submittedAt,
			"duration":     elapsed,
		}).Error; err != nil {
		// 记录失败日志：下一轮 tick 会重试（submitted_at 仍为 NULL）
		log.Printf("[ExamScheduler] 标记记录完成失败 record_id=%d: %v", recordID, err)
	}
}
