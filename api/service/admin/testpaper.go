package admin

import (
	"errors"

	"training/global"
	adminModel "training/model/admin"
	"training/model/common"
	userModel "training/model/user"

	"gorm.io/gorm"
)

// TestpaperList 试卷列表（分页）
func TestpaperList(courseID uint, req common.PageRequest) (*common.PageList, error) {
	var list []adminModel.Testpaper
	var total int64
	q := global.DB.Model(&adminModel.Testpaper{}).Where("course_id = ?", courseID)
	if req.Keyword != "" {
		q = q.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := q.Order("sort asc, id desc").Offset(req.Offset()).Limit(req.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &common.PageList{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// TestpaperCreate 创建试卷
func TestpaperCreate(courseID uint, req adminModel.TestpaperReq) error {
	// 校验：同一课程下试卷名称不能重复
	var dupCount int64
	if err := global.DB.Model(&adminModel.Testpaper{}).Where("course_id = ? AND name = ?", courseID, req.Name).Count(&dupCount).Error; err != nil {
		return err
	}
	if dupCount > 0 {
		return errors.New("该课程下已存在相同名称的试卷")
	}
	tp := adminModel.Testpaper{
		CourseID:    courseID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		TotalScore:  100,
		PassScore:   req.PassScore,
		Duration:    req.Duration,
		Sort:        req.Sort,
		Status:      req.Status,
	}
	if tp.PassScore == 0 {
		tp.PassScore = 60
	}
	if tp.Duration == 0 {
		tp.Duration = 60
	}
	if tp.Type == 0 {
		tp.Type = 2 // 默认课程完成后考
	}
	if tp.Status == 0 {
		tp.Status = 1
	}

	return global.DB.Transaction(func(tx *gorm.DB) error {
		// 同一课程只允许一张启用试卷：新建启用的试卷时，自动禁用其他试卷
		if tp.Status == 1 {
			if err := tx.Model(&adminModel.Testpaper{}).Where("course_id = ? AND status = 1", courseID).
				Update("status", 0).Error; err != nil {
				return err
			}
		}
		return tx.Create(&tp).Error
	})
}

// TestpaperUpdate 更新试卷
func TestpaperUpdate(id uint, req adminModel.TestpaperReq) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		var tp adminModel.Testpaper
		if err := tx.First(&tp, id).Error; err != nil {
			return errors.New("试卷不存在")
		}
		// 校验：同一课程下试卷名称不能重复
		var dupCount int64
		if err := tx.Model(&adminModel.Testpaper{}).Where("course_id = ? AND name = ? AND id != ?", tp.CourseID, req.Name, id).Count(&dupCount).Error; err != nil {
			return err
		}
		if dupCount > 0 {
			return errors.New("该课程下已存在相同名称的试卷")
		}
		// 同一课程只允许一张启用试卷：启用时自动禁用同课程其他试卷
		if req.Status == 1 {
			if err := tx.Model(&adminModel.Testpaper{}).
				Where("course_id = ? AND status = 1 AND id != ?", tp.CourseID, id).
				Update("status", 0).Error; err != nil {
				return err
			}
		}
		return tx.Model(&adminModel.Testpaper{}).Where("id = ?", id).Updates(map[string]interface{}{
			"name":        req.Name,
			"description": req.Description,
			"type":        req.Type,
			"pass_score":  req.PassScore,
			"duration":    req.Duration,
			"sort":        req.Sort,
			"status":      req.Status,
		}).Error
	})
}

// TestpaperDelete 删除试卷（同时删除关联的试题和考试记录）
func TestpaperDelete(id uint) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("testpaper_id = ?", id).Delete(&adminModel.TestpaperQuestion{}).Error; err != nil {
			return err
		}
		if err := tx.Where("testpaper_id = ?", id).Delete(&userModel.TestpaperRecord{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&adminModel.Testpaper{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

// TestpaperGetQuestions 获取试卷的试题列表（含分值和试题内容）
func TestpaperGetQuestions(testpaperID uint) ([]map[string]interface{}, error) {
	var items []adminModel.TestpaperQuestion
	if err := global.DB.Where("testpaper_id = ?", testpaperID).Order("sort asc, id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		var q adminModel.Question
		if err := global.DB.First(&q, item.QuestionID).Error; err != nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"id":          item.ID,
			"testpaperId": item.TestpaperID,
			"questionId":  item.QuestionID,
			"score":       item.Score,
			"sort":        item.Sort,
			"question":    q,
		})
	}
	return result, nil
}

// TestpaperSetQuestions 设置试卷试题（全量覆盖）
func TestpaperSetQuestions(testpaperID uint, items []adminModel.TestpaperQuestionItem) error {
	// 校验：所有试题必须属于同一课程
	var tp adminModel.Testpaper
	if err := global.DB.First(&tp, testpaperID).Error; err != nil {
		return errors.New("试卷不存在")
	}
	for _, item := range items {
		var q adminModel.Question
		if err := global.DB.First(&q, item.QuestionID).Error; err != nil {
			return errors.New("试题不存在")
		}
		if q.CourseID != tp.CourseID {
			return errors.New("试题不属于该课程")
		}
	}

	return global.DB.Transaction(func(tx *gorm.DB) error {
		// 先删除旧的
		if err := tx.Where("testpaper_id = ?", testpaperID).Delete(&adminModel.TestpaperQuestion{}).Error; err != nil {
			return err
		}
		// 批量插入
		for i, item := range items {
			tq := adminModel.TestpaperQuestion{
				TestpaperID: testpaperID,
				QuestionID:  item.QuestionID,
				Score:       item.Score,
				Sort:        i + 1,
			}
			if err := tx.Create(&tq).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// TestpaperGet 获取单条试卷
func TestpaperGet(id uint) (*adminModel.Testpaper, error) {
	var tp adminModel.Testpaper
	if err := global.DB.First(&tp, id).Error; err != nil {
		return nil, errors.New("试卷不存在")
	}
	return &tp, nil
}
