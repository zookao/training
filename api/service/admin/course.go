package admin

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"training/global"
	adminModel "training/model/admin"
	"training/model/common"
	"training/utils"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// CoursePage 课程分页
func CoursePage(req common.PageRequest) (*common.PageList, error) {
	req.Normalize()
	var list []adminModel.Course
	var total int64
	q := global.DB.Model(&adminModel.Course{})
	if req.Keyword != "" {
		q = q.Where("title LIKE ?", "%"+req.Keyword+"%")
	}
	q.Count(&total)
	if err := q.Preload("Videos").
		Order("sort asc, id desc").
		Offset(req.Offset()).Limit(req.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &common.PageList{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// CourseAll 全部课程（班级绑定用）
func CourseAll() ([]adminModel.Course, error) {
	var list []adminModel.Course
	if err := global.DB.Order("sort asc, id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// CourseGet 课程详情（含视频，按 sort 排序）
func CourseGet(id uint) (*adminModel.Course, error) {
	var c adminModel.Course
	if err := global.DB.Preload("Videos", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort asc, id asc")
	}).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// CourseCreate 创建课程（含视频）
func CourseCreate(req adminModel.CourseReq) error {
	c := adminModel.Course{
		Title:       req.Title,
		Cover:       req.Cover,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      req.Status,
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&c).Error; err != nil {
			return err
		}
		return syncVideos(tx, c.ID, req.Videos)
	})
}

// CourseUpdate 更新课程（含视频同步：新增/更新/删除）
func CourseUpdate(id uint, req adminModel.CourseReq) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"title":       req.Title,
			"cover":       req.Cover,
			"description": req.Description,
			"sort":        req.Sort,
			"status":      req.Status,
		}
		if err := tx.Model(&adminModel.Course{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		return syncVideos(tx, id, req.Videos)
	})
}

// CourseDelete 删除课程（含视频记录与磁盘文件）
func CourseDelete(id uint) error {
	var videos []adminModel.Video
	if err := global.DB.Where("course_id = ?", id).Find(&videos).Error; err != nil {
		return err
	}
	var c adminModel.Course
	if err := global.DB.Select("cover").First(&c, id).Error; err != nil {
		return err
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("course_id = ?", id).Delete(&adminModel.Video{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&adminModel.Course{}, id).Error; err != nil {
			return err
		}
		// 事务提交成功后再清理磁盘文件
		for _, v := range videos {
			utils.DeleteUploadByURL(v.URL)
			utils.DeleteUploadByURL(v.Thumbnail)
			utils.DeleteUploadByURL(v.Courseware)
			utils.DeleteUploadByURL(v.CoursewarePDF)
		}
		utils.DeleteUploadByURL(c.Cover)
		return nil
	})
}

// VideoDelete 删除单个视频（记录 + 磁盘文件）
func VideoDelete(courseID, videoID uint) error {
	var v adminModel.Video
	if err := global.DB.Where("id = ? AND course_id = ?", videoID, courseID).First(&v).Error; err != nil {
		return errors.New("视频不存在")
	}
	if err := global.DB.Delete(&v).Error; err != nil {
		return err
	}
	utils.DeleteUploadByURL(v.URL)
	utils.DeleteUploadByURL(v.Thumbnail)
	utils.DeleteUploadByURL(v.Courseware)
	utils.DeleteUploadByURL(v.CoursewarePDF)
	return nil
}

// ReparseCoursewarePages 重新解析视频课件的页数并落库
// 用于：旧课件上传时未自动识别页数，或解析失败后手动重试
// 优先从 PDF 统计（支持所有课件格式），回退到 PPTX ZIP 解析（兼容旧数据无 PDF）
func ReparseCoursewarePages(videoID uint) (int, error) {
	var v adminModel.Video
	if err := global.DB.First(&v, videoID).Error; err != nil {
		return 0, errors.New("视频不存在")
	}
	if v.Courseware == "" {
		return 0, errors.New("该视频未上传课件")
	}

	// 1. 优先从 PDF 统计页数（新上传的课件都有 PDF）
	if v.CoursewarePDF != "" {
		pdfPath, ok := utils.URLToLocalPath(v.CoursewarePDF)
		if ok {
			if count, err := utils.CountPDFPages(pdfPath); err == nil && count > 0 {
				if err := global.DB.Model(&adminModel.Video{}).Where("id = ?", videoID).
					Update("courseware_page_count", count).Error; err != nil {
					return 0, err
				}
				return count, nil
			}
		}
	}

	// 2. 回退：从 PPTX 原始文件统计（兼容旧数据，仅对 PPTX 格式有效）
	localPath, ok := utils.URLToLocalPath(v.Courseware)
	if !ok {
		return 0, errors.New("课件路径无效")
	}
	count, err := utils.CountPptxSlides(localPath)
	if err != nil {
		return 0, errors.New("课件解析失败: " + err.Error())
	}
	if err := global.DB.Model(&adminModel.Video{}).Where("id = ?", videoID).
		Update("courseware_page_count", count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// syncVideos 同步课程下的视频：req 中带 ID 的更新、不带 ID 的新增；现有但不在 req 中的删除（含磁盘文件）
func syncVideos(tx *gorm.DB, courseID uint, reqs []adminModel.VideoReq) error {
	// 现有视频
	var existing []adminModel.Video
	if err := tx.Where("course_id = ?", courseID).Find(&existing).Error; err != nil {
		return err
	}
	keepIDs := make(map[uint]bool, len(reqs))
	for _, r := range reqs {
		if r.ID > 0 {
			keepIDs[r.ID] = true
		}
	}
	// 删除：现有但不在 keep 集合中
	for _, e := range existing {
		if !keepIDs[e.ID] {
			if err := tx.Delete(&adminModel.Video{}, e.ID).Error; err != nil {
				return err
			}
			utils.DeleteUploadByURL(e.URL)
			utils.DeleteUploadByURL(e.Thumbnail)
			utils.DeleteUploadByURL(e.Courseware)
			utils.DeleteUploadByURL(e.CoursewarePDF)
		}
	}
	// 新增/更新
	for _, r := range reqs {
		if r.ID > 0 {
			updates := map[string]interface{}{
				"url":                   r.URL,
				"thumbnail":             r.Thumbnail,
				"courseware":            r.Courseware,
				"courseware_page_count": r.CoursewarePageCount,
				"courseware_pages":      r.CoursewarePages,
				"courseware_pdf":        r.CoursewarePDF,
				"title":                 r.Title,
				"description":           r.Description,
				"sort":                  r.Sort,
				"duration":              r.Duration,
			}
			if err := tx.Model(&adminModel.Video{}).Where("id = ? AND course_id = ?", r.ID, courseID).
				Updates(updates).Error; err != nil {
				return err
			}
		} else {
			v := adminModel.Video{
				CourseID:            courseID,
				URL:                 r.URL,
				Thumbnail:           r.Thumbnail,
				Courseware:          r.Courseware,
				CoursewarePageCount: r.CoursewarePageCount,
				CoursewarePages:     r.CoursewarePages,
				CoursewarePDF:       r.CoursewarePDF,
				Title:               r.Title,
				Description:         r.Description,
				Sort:                r.Sort,
				Duration:            r.Duration,
			}
			if err := tx.Create(&v).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// ParseDurations 解析Excel打点，计算每页时长（秒）
// Excel: 页码 | 视频打点(hh:mm:ss)，每页时长 = 下一页打点 - 当前页打点，末页 = totalDuration - 末页打点
func ParseDurations(file []byte, totalDuration int) ([]int, error) {
	f, err := excelize.OpenReader(bytes.NewReader(file))
	if err != nil {
		return nil, errors.New("无法读取 Excel 文件")
	}
	defer f.Close()
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, errors.New("无法读取工作表")
	}
	if len(rows) <= 1 {
		return nil, errors.New("Excel 无数据行")
	}
	type pt struct{ page, sec int }
	var points []pt
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 2 {
			continue
		}
		pageStr := strings.TrimSpace(row[0])
		timeStr := strings.TrimSpace(row[1])
		if pageStr == "" || timeStr == "" {
			continue
		}
		page, _ := strconv.Atoi(pageStr)
		sec := parseTimeToSeconds(timeStr)
		if sec < 0 {
			return nil, fmt.Errorf("第%d行打点格式错误: %s（应为 hh:mm:ss 或 mm:ss）", i+1, timeStr)
		}
		points = append(points, pt{page, sec})
	}
	if len(points) == 0 {
		return nil, errors.New("未解析到有效打点")
	}
	sort.Slice(points, func(i, j int) bool { return points[i].page < points[j].page })
	durations := make([]int, len(points))
	for i := 0; i < len(points); i++ {
		// 打点[i] = 页(i+1)的结束时间，页1从0开始，末页到视频结束
		var start, end int
		if i == 0 {
			start = 0
		} else {
			start = points[i-1].sec
		}
		if i == len(points)-1 && totalDuration > points[i].sec {
			end = totalDuration // 末页：视频比最后打点长时，到视频结束
		} else {
			end = points[i].sec
		}
		durations[i] = end - start
		if durations[i] < 0 {
			durations[i] = 0
		}
	}
	return durations, nil
}

// parseTimeToSeconds "00:03:23" → 203, "03:23" → 203, "00:00:05.86" → 6, Excel时间序列号 → 秒
// 秒部分支持小数毫秒（如 05.86），四舍五入到整数秒，兼容外部播放器/工具复制的时间戳
func parseTimeToSeconds(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return -1
	}
	// 全角冒号转半角
	s = strings.ReplaceAll(s, "：", ":")
	// 1. hh:mm:ss 或 mm:ss（秒部分可为 05.86 形式）
	parts := strings.Split(s, ":")
	if len(parts) == 3 {
		h, _ := strconv.Atoi(parts[0])
		m, _ := strconv.Atoi(parts[1])
		secF, _ := strconv.ParseFloat(parts[2], 64)
		return h*3600 + m*60 + int(math.Round(secF))
	}
	if len(parts) == 2 {
		m, _ := strconv.Atoi(parts[0])
		secF, _ := strconv.ParseFloat(parts[1], 64)
		return m*60 + int(math.Round(secF))
	}
	// 2. Excel 时间序列号（浮点数，1天=86400秒）
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		sec := int(math.Round(f * 86400))
		if sec >= 0 && sec < 86400*366 {
			return sec
		}
	}
	return -1
}
