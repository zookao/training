package admin

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"training/global"
	adminModel "training/model/admin"
	"training/model/common"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// QuestionList 试题列表（分页）
func QuestionList(courseID uint, req common.PageRequest) (*common.PageList, error) {
	var list []adminModel.Question
	var total int64
	q := global.DB.Model(&adminModel.Question{}).Where("course_id = ?", courseID)
	if req.Keyword != "" {
		q = q.Where("title LIKE ?", "%"+req.Keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := q.Order("sort asc, id desc").Offset(req.Offset()).Limit(req.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &common.PageList{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// QuestionAll 获取课程下全部启用试题（试卷组卷用）
func QuestionAll(courseID uint) ([]adminModel.Question, error) {
	var list []adminModel.Question
	if err := global.DB.Where("course_id = ? AND status = 1", courseID).Order("sort asc, id asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// QuestionCreate 创建试题
func QuestionCreate(courseID uint, req adminModel.QuestionReq) error {
	// 校验：同一课程下试题题干不能重复
	var dupCount int64
	if err := global.DB.Model(&adminModel.Question{}).Where("course_id = ? AND title = ?", courseID, req.Title).Count(&dupCount).Error; err != nil {
		return err
	}
	if dupCount > 0 {
		return errors.New("该课程下已存在相同题干的试题")
	}
	q := adminModel.Question{
		CourseID:  courseID,
		Type:      req.Type,
		Title:     req.Title,
		Options:   req.Options,
		Answer:    req.Answer,
		Analysis:  req.Analysis,
		Sort:      req.Sort,
		Status:    req.Status,
	}
	return global.DB.Create(&q).Error
}

// QuestionUpdate 更新试题
func QuestionUpdate(id uint, req adminModel.QuestionReq) error {
	// 校验：同一课程下试题题干不能重复
	var q adminModel.Question
	if err := global.DB.First(&q, id).Error; err != nil {
		return errors.New("试题不存在")
	}
	var dupCount int64
	if err := global.DB.Model(&adminModel.Question{}).Where("course_id = ? AND title = ? AND id != ?", q.CourseID, req.Title, id).Count(&dupCount).Error; err != nil {
		return err
	}
	if dupCount > 0 {
		return errors.New("该课程下已存在相同题干的试题")
	}
	return global.DB.Model(&adminModel.Question{}).Where("id = ?", id).Updates(map[string]interface{}{
		"type":      req.Type,
		"title":     req.Title,
		"options":   req.Options,
		"answer":    req.Answer,
		"analysis":  req.Analysis,
		"sort":      req.Sort,
		"status":    req.Status,
	}).Error
}

// QuestionDelete 删除试题（同时从试卷中移除）
func QuestionDelete(id uint) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("question_id = ?", id).Delete(&adminModel.TestpaperQuestion{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&adminModel.Question{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

// QuestionGet 获取单条试题
func QuestionGet(id uint) (*adminModel.Question, error) {
	var q adminModel.Question
	if err := global.DB.First(&q, id).Error; err != nil {
		return nil, errors.New("试题不存在")
	}
	return &q, nil
}

// QuestionImportRow 试题导入结果单行
type QuestionImportRow struct {
	Row     int    `json:"row"`    // Excel 行号（从 2 开始，1 为表头）
	Type    string `json:"type"`   // 题型文本
	Title   string `json:"title"`  // 题干
	Answer  string `json:"answer"` // 原始答案文本
	Success bool   `json:"success"`
	Reason  string `json:"reason"` // 失败原因
}

// QuestionImportResult 试题导入结果汇总
type QuestionImportResult struct {
	Total   int                 `json:"total"`
	Success int                 `json:"success"`
	Failed  int                 `json:"failed"`
	Rows    []QuestionImportRow `json:"rows"`
}

// questionOption 试题选项（与前端 QuestionOption 一致：{"label":"A","content":"..."}）
type questionOption struct {
	Label   string `json:"label"`
	Content string `json:"content"`
}

// QuestionImport 批量导入试题（Excel）
// 列顺序：题型、题干、选项A、选项B、选项C、选项D、正确答案、解析
// 规则：题型/题干/答案必填；单选多选至少 2 个选项且答案字母须在选项范围内；单选答案仅 1 个；
// 判断题选项默认 正确(A)/错误(B)（填写选项A/选项B可覆盖），答案填 A 或 B。状态默认启用。
func QuestionImport(courseID uint, file []byte) (*QuestionImportResult, error) {
	f, err := excelize.OpenReader(bytesReader(file))
	if err != nil {
		return nil, errors.New("无法读取 Excel 文件")
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, errors.New("无法读取工作表")
	}

	res := &QuestionImportResult{Rows: []QuestionImportRow{}}
	if len(rows) <= 1 {
		return res, nil
	}

	// 批量导入内题干去重（同一文件内重复题干跳过）
	seenTitles := make(map[string]bool)

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		typeStr := excelCell(row, 0)
		title := excelCell(row, 1)
		optA := excelCell(row, 2)
		optB := excelCell(row, 3)
		optC := excelCell(row, 4)
		optD := excelCell(row, 5)
		ansStr := excelCell(row, 6)
		analysis := excelCell(row, 7)
		item := QuestionImportRow{Row: i + 1, Type: typeStr, Title: title, Answer: ansStr}

		qType, ok := parseQuestionType(typeStr)
		if !ok {
			item.Reason = "题型错误（应为 单选/多选/判断）"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		if title == "" {
			item.Reason = "题干为空"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		// 校验：同一课程下题干不能重复（数据库已存在 或 本批次已出现过）
		if seenTitles[title] {
			item.Reason = "导入文件中存在重复题干"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		var dbDupCount int64
		global.DB.Model(&adminModel.Question{}).Where("course_id = ? AND title = ?", courseID, title).Count(&dbDupCount)
		if dbDupCount > 0 {
			item.Reason = "该课程下已存在相同题干的试题"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		seenTitles[title] = true

		var opts []questionOption
		var answer []string
		if qType == 3 {
			// 判断题：选项默认 正确(A)/错误(B)；若填写了选项A/选项B则用填写的
			opts = buildQuestionOptions(optA, optB)
			if len(opts) < 2 {
				opts = []questionOption{{Label: "A", Content: "正确"}, {Label: "B", Content: "错误"}}
			}
			a, ok := parseJudgeAnswer(ansStr)
			if !ok {
				item.Reason = "判断题答案应为 A 或 B"
				res.Failed++
				res.Rows = append(res.Rows, item)
				continue
			}
			answer = []string{a}
		} else {
			opts = buildQuestionOptions(optA, optB, optC, optD)
			if len(opts) < 2 {
				item.Reason = "选项至少需要 2 个"
				res.Failed++
				res.Rows = append(res.Rows, item)
				continue
			}
			answer = parseChoiceAnswer(ansStr)
			if len(answer) == 0 {
				item.Reason = "答案为空或格式不正确"
				res.Failed++
				res.Rows = append(res.Rows, item)
				continue
			}
			labels := make(map[string]bool, len(opts))
			for _, o := range opts {
				labels[o.Label] = true
			}
			invalid := false
			for _, a := range answer {
				if !labels[a] {
					invalid = true
					break
				}
			}
			if invalid {
				item.Reason = "答案字母超出选项范围"
				res.Failed++
				res.Rows = append(res.Rows, item)
				continue
			}
			if qType == 1 && len(answer) > 1 {
				item.Reason = "单选题答案只能有一个"
				res.Failed++
				res.Rows = append(res.Rows, item)
				continue
			}
		}

		optsJSON, _ := json.Marshal(opts)
		ansJSON, _ := json.Marshal(answer)
		q := adminModel.Question{
			CourseID: courseID,
			Type:     qType,
			Title:    title,
			Options:  string(optsJSON),
			Answer:   string(ansJSON),
			Analysis: analysis,
			Status:   1,
		}
		if err := global.DB.Create(&q).Error; err != nil {
			item.Reason = "创建失败：" + err.Error()
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		item.Success = true
		res.Success++
		res.Rows = append(res.Rows, item)
	}
	res.Total = res.Success + res.Failed
	return res, nil
}

// excelCell 安全读取 Excel 行单元格（越界返回空串）
func excelCell(row []string, idx int) string {
	if idx >= 0 && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

// parseQuestionType 解析题型文本为 1单选 2多选 3判断
func parseQuestionType(s string) (int8, bool) {
	switch s {
	case "单选", "单选题":
		return 1, true
	case "多选", "多选题":
		return 2, true
	case "判断", "判断题":
		return 3, true
	}
	return 0, false
}

// parseJudgeAnswer 解析判断题答案：A → A，B → B（仅接受 A/B，不区分大小写）
func parseJudgeAnswer(s string) (string, bool) {
	switch strings.ToUpper(s) {
	case "A":
		return "A", true
	case "B":
		return "B", true
	}
	return "", false
}

// parseChoiceAnswer 解析单选/多选答案：支持 "A"、"AC"、"A,C"、"A、C" 等，返回去重排序后的字母数组
func parseChoiceAnswer(s string) []string {
	s = strings.ToUpper(s)
	s = strings.NewReplacer(",", "", "，", "", "、", "", " ", "", ";", "", "；", "", "/", "").Replace(s)
	set := make(map[string]bool)
	out := make([]string, 0, len(s))
	for _, c := range s {
		ch := string(c)
		if len(ch) == 1 && ch[0] >= 'A' && ch[0] <= 'H' && !set[ch] {
			set[ch] = true
			out = append(out, ch)
		}
	}
	sort.Strings(out)
	return out
}

// buildQuestionOptions 从最多 4 个选项内容构建选项列表（按列位置分配 A/B/C/D 标签，跳过空值）
func buildQuestionOptions(contents ...string) []questionOption {
	labels := "ABCDEFGH"
	opts := make([]questionOption, 0, len(contents))
	for i, c := range contents {
		if c == "" {
			continue
		}
		if i >= len(labels) {
			break
		}
		opts = append(opts, questionOption{Label: string(labels[i]), Content: c})
	}
	return opts
}
