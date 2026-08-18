package admin

import (
	"errors"
	"fmt"

	"training/global"
	adminModel "training/model/admin"
	"training/model/common"
	userModel "training/model/user"
	"training/utils"

	"gorm.io/gorm"
)

// ClassPage 班级分页
func ClassPage(req common.PageRequest) (*common.PageList, error) {
	req.Normalize()
	var list []adminModel.Class
	var total int64
	q := global.DB.Model(&adminModel.Class{})
	if req.Keyword != "" {
		q = q.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	q.Count(&total)
	if err := q.Preload("Courses").Preload("Users").
		Order("sort asc, id desc").
		Offset(req.Offset()).Limit(req.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &common.PageList{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// ClassGet 班级详情
func ClassGet(id uint) (*adminModel.Class, error) {
	var c adminModel.Class
	if err := global.DB.Preload("Courses").Preload("Users").First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// ClassCreate 创建班级
func ClassCreate(req adminModel.ClassReq) error {
	c := adminModel.Class{
		Name:        req.Name,
		Cover:       req.Cover,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      req.Status,
	}
	return global.DB.Create(&c).Error
}

// ClassUpdate 更新班级
func ClassUpdate(id uint, req adminModel.ClassReq) error {
	updates := map[string]interface{}{
		"name":        req.Name,
		"cover":       req.Cover,
		"description": req.Description,
		"sort":        req.Sort,
		"status":      req.Status,
	}
	return global.DB.Model(&adminModel.Class{}).Where("id = ?", id).Updates(updates).Error
}

// ClassDelete 删除班级
func ClassDelete(id uint) error {
	var c adminModel.Class
	if err := global.DB.Select("cover").First(&c, id).Error; err != nil {
		return err
	}
	// 班级有学习数据时不允许删除
	var count int64
	global.DB.Model(&userModel.VideoRecord{}).Where("class_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("该班级已有 %d 条视频学习记录，不允许删除", count)
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("class_id = ?", id).Delete(&adminModel.ClassCourse{})
		tx.Where("class_id = ?", id).Delete(&adminModel.ClassUser{})
		if err := tx.Delete(&adminModel.Class{}, id).Error; err != nil {
			return err
		}
		utils.DeleteUploadByURL(c.Cover)
		return nil
	})
}

// ClassCourseIDs 班级已绑课程ID
func ClassCourseIDs(id uint) ([]uint, error) {
	var ids []uint
	global.DB.Table("class_courses").Where("class_id = ?", id).Pluck("course_id", &ids)
	return ids, nil
}

// ClassUserIDs 班级已绑学员ID
func ClassUserIDs(id uint) ([]uint, error) {
	var ids []uint
	global.DB.Table("class_users").Where("class_id = ?", id).Pluck("user_id", &ids)
	return ids, nil
}

// AssignClassCourses 分配课程（仅允许分配启用状态的课程）
func AssignClassCourses(id uint, courseIDs []uint) error {
	var exists int64
	global.DB.Model(&adminModel.Class{}).Where("id = ?", id).Count(&exists)
	if exists == 0 {
		return errors.New("班级不存在")
	}
	// 过滤：只保留启用状态的课程
	var validIDs []uint
	if len(courseIDs) > 0 {
		global.DB.Model(&adminModel.Course{}).
			Where("id IN ? AND status = 1", courseIDs).
			Pluck("id", &validIDs)
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("class_id = ?", id).Delete(&adminModel.ClassCourse{})
		if len(validIDs) == 0 {
			return nil
		}
		rows := make([]adminModel.ClassCourse, 0, len(validIDs))
		for _, cid := range validIDs {
			rows = append(rows, adminModel.ClassCourse{ClassID: id, CourseID: cid})
		}
		return tx.Create(&rows).Error
	})
}

// AssignClassUsers 分配学员（仅允许分配启用状态的学员）
func AssignClassUsers(id uint, userIDs []uint) error {
	var exists int64
	global.DB.Model(&adminModel.Class{}).Where("id = ?", id).Count(&exists)
	if exists == 0 {
		return errors.New("班级不存在")
	}
	// 过滤：只保留启用状态的学员
	var validIDs []uint
	if len(userIDs) > 0 {
		global.DB.Model(&userModel.User{}).
			Where("id IN ? AND status = 1", userIDs).
			Pluck("id", &validIDs)
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("class_id = ?", id).Delete(&adminModel.ClassUser{})
		if len(validIDs) == 0 {
			return nil
		}
		rows := make([]adminModel.ClassUser, 0, len(validIDs))
		for _, uid := range validIDs {
			rows = append(rows, adminModel.ClassUser{ClassID: id, UserID: uid})
		}
		return tx.Create(&rows).Error
	})
}
