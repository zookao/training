package admin

import (
	"errors"

	"training/global"
	adminModel "training/model/admin"
	"training/model/common"

	"gorm.io/gorm"
)

// ApiPage 接口分页
func ApiPage(req common.PageRequest) (*common.PageList, error) {
	req.Normalize()
	var list []adminModel.Api
	var total int64
	q := global.DB.Model(&adminModel.Api{})
	if req.Keyword != "" {
		q = q.Where("path LIKE ? OR description LIKE ? OR `group` LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	q.Count(&total)
	if err := q.Order("id desc").
		Offset(req.Offset()).Limit(req.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &common.PageList{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// ApiAll 全部接口（分配用）
func ApiAll() ([]adminModel.Api, error) {
	var list []adminModel.Api
	if err := global.DB.Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ApiCreate 创建
func ApiCreate(req adminModel.ApiReq) error {
	a := adminModel.Api{
		Group:       req.Group,
		Method:      req.Method,
		Path:        req.Path,
		Description: req.Description,
	}
	return global.DB.Create(&a).Error
}

// ApiUpdate 更新
func ApiUpdate(id uint, req adminModel.ApiReq) error {
	updates := map[string]interface{}{
		"group":       req.Group,
		"method":      req.Method,
		"path":        req.Path,
		"description": req.Description,
	}
	return global.DB.Model(&adminModel.Api{}).Where("id = ?", id).Updates(updates).Error
}

// ApiDelete 删除
func ApiDelete(id uint) error {
	var exists int64
	global.DB.Model(&adminModel.Api{}).Where("id = ?", id).Count(&exists)
	if exists == 0 {
		return errors.New("接口不存在")
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		tx.Table("role_apis").Where("api_id = ?", id).Delete(nil)
		return tx.Delete(&adminModel.Api{}, id).Error
	})
}
