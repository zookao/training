package user

import (
	"errors"
	"math/rand"
	"time"

	"training/global"
	adminModel "training/model/admin"
	userModel "training/model/user"
)

// 学习校验相关常量：
//   - 任意时长的视频都安排校验（校验点在 CourseLearn 页面加载时即返回前端，不依赖进度上报）
//   - 校验最小间隔 45 分钟，保证 45 分钟内只校验一次
//   - 首次校验在 1 秒 ~ min(duration-1, 45 分钟) 全范围随机，学员无法预测规律
//   - 后续校验在 45~50 分钟之间随机，防止掌握固定间隔挂机
const (
	checkIntervalSec = 45 * 60 // 校验最小间隔（秒）：45 分钟内只校验一次
	checkJitterSec   = 5 * 60  // 后续校验间隔随机抖动（秒）：45~50 分钟之间
)

// firstCheckPosition 根据视频时长计算首次校验位置（秒）。
// 用 userID+videoID 做随机种子 → 同一学员同一视频的校验点位置确定（每次打开页面一致），
// 不同学员或不同视频的校验点位置不同。
//
// 校验点范围：1 秒 ~ min(duration-1, 45 分钟)，覆盖整个第一个 45 分钟窗口
// （或整个短视频），最大化随机性，学员无法预测规律挂机。
//   - 时长未知（0）：按 45 分钟窗口处理，实际不够长则自然不触发
func firstCheckPosition(userID, videoID uint, duration int) int {
	r := rand.New(rand.NewSource(int64(userID)*1000003 + int64(videoID)))
	maxPos := checkIntervalSec // 45 * 60
	if duration > 0 && duration-1 < maxPos {
		maxPos = duration - 1
	}
	minPos := 1 // 不设 0，避免视频刚播放就弹窗
	if maxPos <= minPos {
		return minPos
	}
	return minPos + r.Intn(maxPos-minPos+1)
}

// MyClasses 当前学员的班级（含完成度）
func MyClasses(userID uint) ([]userModel.ClassProgressItem, error) {
	var classes []adminModel.Class
	if err := global.DB.
		Joins("JOIN class_users ON class_users.class_id = classes.id").
		Where("class_users.user_id = ? AND classes.status = 1", userID).
		Order("classes.sort asc, classes.id desc").
		Find(&classes).Error; err != nil {
		return nil, err
	}
	result := make([]userModel.ClassProgressItem, 0, len(classes))
	for _, c := range classes {
		courseIDs := classCourseIDs(c.ID)
		percent := avgCoursePercent(userID, courseIDs)
		result = append(result, userModel.ClassProgressItem{
			ID:          c.ID,
			Name:        c.Name,
			Cover:       c.Cover,
			Description: c.Description,
			CourseCount: len(courseIDs),
			Percent:     percent,
		})
	}
	return result, nil
}

// ClassDetail 班级详情（含每课程进度）
func ClassDetail(userID, classID uint) (*userModel.ClassDetailRes, error) {
	if !userInClass(userID, classID) {
		return nil, errors.New("未加入该班级")
	}
	var c adminModel.Class
	if err := global.DB.First(&c, classID).Error; err != nil {
		return nil, errors.New("班级不存在")
	}
	courseIDs := classCourseIDs(classID)
	var courses []adminModel.Course
	if len(courseIDs) > 0 {
		global.DB.Where("id IN ?", courseIDs).Order("sort asc, id desc").Find(&courses)
	}
	items := make([]userModel.CourseProgressItem, 0, len(courses))
	percentSum := 0
	for _, co := range courses {
		videos := courseVideos(co.ID)
		completed, pct := courseProgress(userID, co.ID, len(videos))
		items = append(items, userModel.CourseProgressItem{
			ID:              co.ID,
			Title:           co.Title,
			Cover:           co.Cover,
			Description:     co.Description,
			VideoCount:      len(videos),
			CompletedVideos: completed,
			Percent:         pct,
		})
		percentSum += pct
	}
	classPercent := 0
	if len(items) > 0 {
		classPercent = percentSum / len(items)
	}
	return &userModel.ClassDetailRes{
		ID:          c.ID,
		Name:        c.Name,
		Cover:       c.Cover,
		Description: c.Description,
		Courses:     items,
		Percent:     classPercent,
	}, nil
}

// CourseLearn 课程学习详情（视频按序 + 学员每视频进度）
func CourseLearn(userID, courseID uint) (*userModel.CourseLearnRes, error) {
	// 校验：学员必须通过班级选了该课程才能学习
	if !userHasCourseAccess(userID, courseID) {
		return nil, errors.New("未加入包含该课程的班级")
	}
	var c adminModel.Course
	if err := global.DB.First(&c, courseID).Error; err != nil {
		return nil, errors.New("课程不存在")
	}
	videos := courseVideos(courseID)
	records := userCourseRecords(userID, courseID) // map[videoID]VideoRecord
	items := make([]userModel.VideoLearnItem, 0, len(videos))
	for _, v := range videos {
		var rec *userModel.VideoRecord
		if r, ok := records[v.ID]; ok {
			rec = &r
		}
		// 时长以服务端 videos 表存储的值为准（上传时解析）
		duration := v.Duration
		position := 0
		maxPos := 0
		percent := 0
		completed := false
		nextCheck := 0
		checkPending := false
		if rec != nil {
			position = rec.Position
			maxPos = rec.MaxPosition
			completed = rec.Completed
			percent = calcPercent(maxPos, duration)
			nextCheck = rec.NextCheckPosition
			checkPending = rec.CheckPending
		} else {
			// 新视频无记录：计算确定性校验点，使前端打开页面即知道校验位置
			nextCheck = firstCheckPosition(userID, v.ID, duration)
		}
		items = append(items, userModel.VideoLearnItem{
			ID:               v.ID,
			URL:              v.URL,
			Thumbnail:        v.Thumbnail,
			Courseware:        v.Courseware,
			CoursewarePages:   v.CoursewarePages,
			CoursewarePDF:     v.CoursewarePDF,
			Title:            v.Title,
			Description:      v.Description,
			Sort:             v.Sort,
			Duration:         duration,
			Position:         position,
			MaxPosition:      maxPos,
			Percent:          percent,
			Completed:        completed,
			NextCheckPosition: nextCheck,
			CheckPending:      checkPending,
		})
	}
	return &userModel.CourseLearnRes{
		ID:          c.ID,
		Title:       c.Title,
		Cover:       c.Cover,
		Description: c.Description,
		Videos:      items,
	}, nil
}

// 防作弊相关常量
const (
	reportInterval    = 10 // 前端上报周期（秒）
	bufferSeconds     = 3  // 时钟/网络抖动容忍（秒）
	completeThreshold = 2  // 距结尾多少秒即视为学完
)

// ReportProgress 上报学习进度（upsert video_records）
//
// 防作弊策略：
//  1. 不信任前端传入的 completed 标记和 duration，完成判定仅由服务端依据 maxPosition 与 video.duration 计算。
//  2. maxPosition 的增长受实时约束：每次上报最多超过历史最远点 (距上次上报的真实秒数 + buffer) 秒，
//     且不超过一个上报周期 + buffer。这样无法通过一次 position=duration 的请求直接标记完成。
//  3. duration 以服务端 videos 表存储的值为准（上传时由 MP4 box 解析）；旧数据无值时兜底取前端上报值。
//  4. 当前播放位置 position 不得超过防作弊后的最远点。
func ReportProgress(userID uint, req userModel.ProgressReq) (*userModel.ProgressRes, error) {
	// 校验：视频必须存在且属于请求的 courseID，且学员有该课程的访问权限
	var video adminModel.Video
	if dbErr := global.DB.Where("id = ? AND course_id = ?", req.VideoID, req.CourseID).First(&video).Error; dbErr != nil {
		return nil, errors.New("视频不存在或不属于该课程")
	}
	if !userHasCourseAccess(userID, req.CourseID) {
		return nil, errors.New("未加入包含该课程的班级")
	}

	var rec userModel.VideoRecord
	err := global.DB.Where("user_id = ? AND video_id = ?", userID, req.VideoID).First(&rec).Error

	now := time.Now()

	// 时长：优先取服务端 videos 表存储的值（上传时解析），旧数据无值时兜底取前端上报值与历史值最大值
	duration := video.Duration
	if duration == 0 {
		// 兜底：旧视频无服务端时长，取前端上报与历史记录的最大值
		duration = req.Duration
		if rec.Duration > duration {
			duration = rec.Duration
		}
	}

	// 计算本次允许超过历史最远点的最大秒数
	maxAdvance := reportInterval + bufferSeconds // 默认上限（同时用于首次上报无基准的情况）
	if err == nil && !rec.LastAt.IsZero() {
		elapsed := int(now.Sub(rec.LastAt).Seconds())
		if elapsed < 0 {
			elapsed = 0
		}
		maxAdvance = elapsed + bufferSeconds
		if maxAdvance > reportInterval+bufferSeconds {
			maxAdvance = reportInterval + bufferSeconds
		}
	}

	oldMax := 0
	nextCheckPos := 0
	checkPending := false
	if err == nil {
		oldMax = rec.MaxPosition
		nextCheckPos = rec.NextCheckPosition
		checkPending = rec.CheckPending
	}
	advance := req.Position - oldMax
	if advance < 0 {
		advance = 0
	}
	if advance > maxAdvance {
		advance = maxAdvance
	}
	newMax := oldMax + advance
	if newMax < 0 {
		newMax = 0
	}

	// 当前播放位置不得超过防作弊后的最远点
	newPosition := req.Position
	if newPosition > newMax {
		newPosition = newMax
	}

	// 学习校验：无校验点时安排首次校验（任意时长都安排，校验点在页面加载时即返回前端）
	nextCheckChanged := false
	if nextCheckPos == 0 {
		if oldMax == 0 {
			// 新记录：安排首次校验位置（确定性，与 CourseLearn 返回的一致）
			nextCheckPos = firstCheckPosition(userID, req.VideoID, duration)
		} else {
			// 旧记录无校验点：从当前进度往后 45~50 分钟随机安排（不回溯阻断已有进度）
			r := rand.New(rand.NewSource(int64(userID)*1000003 + int64(req.VideoID) + int64(oldMax)))
			nextCheckPos = oldMax + checkIntervalSec + r.Intn(checkJitterSec+1)
		}
		if nextCheckPos > 0 {
			nextCheckChanged = true
		}
	}
	// 校验点强制：到达校验点则置为待校验；待校验期间不得越过校验点（防止跳过校验继续学习）
	if nextCheckPos > 0 {
		if checkPending {
			if newMax > nextCheckPos {
				newMax = nextCheckPos
			}
			if newPosition > nextCheckPos {
				newPosition = nextCheckPos
			}
		} else if newMax >= nextCheckPos || req.Position >= nextCheckPos {
			// 用户已播放到校验点（前端防快进保证 currentTime 不超真实最远点）。
			// 跨会话重新进入时 oldMax 可能落后 nextCheckPos 超过单次上报上限（13s），
			// 此时 newMax 受防作弊 cap 限制到不了 nextCheckPos，但用户确实已到达，
			// 故以 req.Position 兜底置为待校验，避免 CheckPass 因 maxPosition<nextCheckPos 拒绝、弹窗卡死。
			checkPending = true
			newMax = nextCheckPos
			if newPosition > nextCheckPos {
				newPosition = nextCheckPos
			}
		}
	}

	// 完成判定：仅服务端计算，时长须有效；被校验阻断时不判定完成
	completed := false
	if !checkPending && duration > 0 && newMax >= duration-completeThreshold {
		completed = true
		newMax = duration
		if newPosition > newMax {
			newPosition = newMax
		}
	}
	percent := calcPercent(newMax, duration)

	if err == nil {
		updates := map[string]interface{}{
			"course_id":      req.CourseID,
			"class_id":       req.ClassID,
			"position":       newPosition,
			"max_position":   newMax,
			"duration":       duration,
			"completed":      completed,
			"percent":        percent,
			"last_at":        &now,
			"check_pending":  checkPending,
		}
		if nextCheckChanged {
			updates["next_check_position"] = nextCheckPos
		}
		if e := global.DB.Model(&rec).Updates(updates).Error; e != nil {
			return nil, e
		}
	} else {
		rec = userModel.VideoRecord{
			UserID:             userID,
			VideoID:            req.VideoID,
			CourseID:           req.CourseID,
			ClassID:            req.ClassID,
			Position:           newPosition,
			MaxPosition:        newMax,
			Duration:           duration,
			Completed:          completed,
			Percent:            percent,
			NextCheckPosition:  nextCheckPos,
			CheckPending:       checkPending,
			LastAt:             now,
		}
		if e := global.DB.Create(&rec).Error; e != nil {
			return nil, e
		}
	}
	return &userModel.ProgressRes{
		Percent:           percent,
		Completed:         completed,
		MaxPosition:       newMax,
		CheckPending:      checkPending,
		NextCheckPosition: nextCheckPos,
	}, nil
}

// ---------- helpers ----------

// CheckPass 通过滑动校验：清除待校验状态，并在当前最远点之后 45~50 分钟随机安排下一次校验。
// 仅在已到达校验点或已有待校验时允许通过，防止提前调用跳过校验。
func CheckPass(userID, videoID uint) (int, error) {
	var rec userModel.VideoRecord
	if err := global.DB.Where("user_id = ? AND video_id = ?", userID, videoID).First(&rec).Error; err != nil {
		return 0, errors.New("未找到学习记录")
	}
	if !rec.CheckPending && (rec.NextCheckPosition == 0 || rec.MaxPosition < rec.NextCheckPosition) {
		return 0, errors.New("当前无需校验")
	}
	// 后续校验在 45~50 分钟之间随机，防止学员掌握固定间隔规律挂机
	r := rand.New(rand.NewSource(int64(userID)*1000003 + int64(videoID) + int64(rec.NextCheckPosition)))
	nextPos := rec.MaxPosition + checkIntervalSec + r.Intn(checkJitterSec+1)
	if e := global.DB.Model(&rec).Updates(map[string]interface{}{
		"check_pending":       false,
		"next_check_position": nextPos,
	}).Error; e != nil {
		return 0, e
	}
	// 记录校验通过日志（非关键路径，失败不影响校验流程）
	global.DB.Create(&userModel.CheckLog{
		UserID:            userID,
		VideoID:           videoID,
		CourseID:          rec.CourseID,
		ClassID:           rec.ClassID,
		CheckPosition:     rec.MaxPosition,
		MaxPosition:       rec.MaxPosition,
		NextCheckPosition: nextPos,
	})
	return nextPos, nil
}

func calcPercent(maxPos, duration int) int {
	if duration <= 0 {
		return 0
	}
	p := maxPos * 100 / duration
	if p > 100 {
		p = 100
	}
	if p < 0 {
		p = 0
	}
	return p
}

func classCourseIDs(classID uint) []uint {
	var ids []uint
	global.DB.Table("class_courses").Where("class_id = ?", classID).Pluck("course_id", &ids)
	return ids
}

func userInClass(userID, classID uint) bool {
	var count int64
	global.DB.Table("class_users").Where("class_id = ? AND user_id = ?", classID, userID).Count(&count)
	return count > 0
}

// userHasCourseAccess 校验学员是否通过班级选了指定课程
// 即：存在一条 class_users(user_id=userID) 关联 class_courses(course_id=courseID) 的记录
func userHasCourseAccess(userID, courseID uint) bool {
	var count int64
	global.DB.Table("class_users cu").
		Joins("INNER JOIN class_courses cc ON cc.class_id = cu.class_id").
		Where("cu.user_id = ? AND cc.course_id = ?", userID, courseID).
		Count(&count)
	return count > 0
}

func courseVideos(courseID uint) []adminModel.Video {
	var videos []adminModel.Video
	global.DB.Where("course_id = ?", courseID).Order("sort asc, id asc").Find(&videos)
	return videos
}

func userCourseRecords(userID, courseID uint) map[uint]userModel.VideoRecord {
	var records []userModel.VideoRecord
	global.DB.Where("user_id = ? AND course_id = ?", userID, courseID).Find(&records)
	m := make(map[uint]userModel.VideoRecord, len(records))
	for _, r := range records {
		m[r.VideoID] = r
	}
	return m
}

// courseProgress 返回 (已完成视频数, 课程进度%)
func courseProgress(userID, courseID uint, videoCount int) (int, int) {
	if videoCount == 0 {
		return 0, 0
	}
	records := userCourseRecords(userID, courseID)
	completed := 0
	percentSum := 0
	for vid, r := range records {
		_ = vid
		if r.Completed {
			completed++
		}
		percentSum += calcPercent(r.MaxPosition, r.Duration)
	}
	// 未学习视频计 0%
	return completed, percentSum / videoCount
}

// avgCoursePercent 多课程的平均进度
func avgCoursePercent(userID uint, courseIDs []uint) int {
	if len(courseIDs) == 0 {
		return 0
	}
	sum := 0
	for _, cid := range courseIDs {
		videos := courseVideos(cid)
		_, pct := courseProgress(userID, cid, len(videos))
		sum += pct
	}
	return sum / len(courseIDs)
}

// ClassLearningReport 班级学习报告（admin 视角：班级所有学员的完成度）
func ClassLearningReport(classID uint) (*userModel.ClassLearningReportRes, error) {
	var c adminModel.Class
	if err := global.DB.First(&c, classID).Error; err != nil {
		return nil, errors.New("班级不存在")
	}
	var us []userModel.User
	global.DB.Joins("JOIN class_users ON class_users.user_id = users.id").
		Where("class_users.class_id = ?", classID).
		Order("users.id desc").Find(&us)
	courseIDs := classCourseIDs(classID)
	students := make([]userModel.StudentReportItem, 0, len(us))
	for _, u := range us {
		students = append(students, userModel.StudentReportItem{
			UserID:    u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			StudentNo: u.StudentNo,
			Percent:   avgCoursePercent(u.ID, courseIDs),
		})
	}
	return &userModel.ClassLearningReportRes{
		ClassID:   c.ID,
		ClassName: c.Name,
		Students:  students,
	}, nil
}

// StudentLearningDetail 学员详细学习数据（admin 视角：课程+视频进度）
func StudentLearningDetail(classID, userID uint) (*userModel.StudentLearningDetailRes, error) {
	var u userModel.User
	if err := global.DB.First(&u, userID).Error; err != nil {
		return nil, errors.New("学员不存在")
	}
	courseIDs := classCourseIDs(classID)
	var courses []adminModel.Course
	if len(courseIDs) > 0 {
		global.DB.Where("id IN ?", courseIDs).Order("sort asc, id desc").Find(&courses)
	}
	items := make([]userModel.CourseReportItem, 0, len(courses))
	for _, co := range courses {
		videos := courseVideos(co.ID)
		completed, pct := courseProgress(userID, co.ID, len(videos))
		records := userCourseRecords(userID, co.ID)
		videoItems := make([]userModel.VideoReportItem, 0, len(videos))
		for _, v := range videos {
			vPercent := 0
			vCompleted := false
			vMaxPos := 0
			if r, ok := records[v.ID]; ok {
				vPercent = calcPercent(r.MaxPosition, v.Duration)
				vCompleted = r.Completed
				vMaxPos = r.MaxPosition
			}
			videoItems = append(videoItems, userModel.VideoReportItem{
				VideoID:     v.ID,
				Title:       v.Title,
				Percent:     vPercent,
				Completed:   vCompleted,
				Duration:    v.Duration,
				MaxPosition: vMaxPos,
			})
		}
		items = append(items, userModel.CourseReportItem{
			CourseID:        co.ID,
			Title:           co.Title,
			Percent:         pct,
			CompletedVideos: completed,
			VideoCount:      len(videos),
			Videos:          videoItems,
		})
	}
	return &userModel.StudentLearningDetailRes{
		UserID:   u.ID,
		Username: u.Username,
		Nickname: u.Nickname,
		Courses:  items,
	}, nil
}
