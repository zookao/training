package admin

import (
	"errors"

	"training/global"
	adminModel "training/model/admin"
	"training/model/common"

	"gorm.io/gorm"
)

// RolePage 角色分页
func RolePage(req common.PageRequest, guard string) (*common.PageList, error) {
	req.Normalize()
	var list []adminModel.Role
	var total int64
	q := global.DB.Model(&adminModel.Role{})
	if guard != "" {
		q = q.Where("guard_type = ?", guard)
	} else {
		q = q.Where("guard_type = ?", "admin")
	}
	if req.Keyword != "" {
		q = q.Where("name LIKE ? OR title LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	q.Count(&total)
	if err := q.Order("sort asc, id desc").
		Offset(req.Offset()).Limit(req.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &common.PageList{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// RoleAll 全部角色（下拉用）
func RoleAll(guard string) ([]adminModel.Role, error) {
	var list []adminModel.Role
	q := global.DB
	if guard == "" {
		guard = "admin"
	}
	if err := q.Where("guard_type = ? AND status = 1", guard).Order("sort asc, id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// RoleGet 详情
func RoleGet(id uint) (*adminModel.Role, error) {
	var r adminModel.Role
	if err := global.DB.First(&r, id).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

// RoleGetMenus 角色的菜单ID
func RoleGetMenuIDs(id uint) ([]uint, error) {
	var ids []uint
	global.DB.Table("role_menus").Where("role_id = ?", id).Pluck("menu_id", &ids)
	return ids, nil
}

// RoleGetApiIDs 角色的接口ID
func RoleGetApiIDs(id uint) ([]uint, error) {
	var ids []uint
	global.DB.Table("role_apis").Where("role_id = ?", id).Pluck("api_id", &ids)
	return ids, nil
}

// RoleCreate 创建
func RoleCreate(req adminModel.RoleReq) error {
	var exists int64
	global.DB.Model(&adminModel.Role{}).Where("name = ?", req.Name).Count(&exists)
	if exists > 0 {
		return errors.New("角色标识已存在")
	}
	r := adminModel.Role{
		Name:      req.Name,
		Title:     req.Title,
		GuardType: defaultGuard(req.GuardType),
		Sort:      req.Sort,
		Status:    req.Status,
		Remark:    req.Remark,
	}
	return global.DB.Create(&r).Error
}

// RoleUpdate 更新
func RoleUpdate(id uint, req adminModel.RoleReq) error {
	if id == 1 {
		return errors.New("超级管理员角色不可编辑")
	}
	updates := map[string]interface{}{
		"name":       req.Name,
		"title":      req.Title,
		"guard_type": defaultGuard(req.GuardType),
		"sort":       req.Sort,
		"status":     req.Status,
		"remark":     req.Remark,
	}
	return global.DB.Model(&adminModel.Role{}).Where("id = ?", id).Updates(updates).Error
}

// RoleDelete 删除
func RoleDelete(id uint) error {
	if id == 1 {
		return errors.New("超级管理员角色不可删除")
	}
	var cnt int64
	global.DB.Table("admin_roles").Where("role_id = ?", id).Count(&cnt)
	if cnt > 0 {
		return errors.New("该角色下存在管理员，无法删除")
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("role_id = ?", id).Delete(&adminModel.RoleMenu{})
		tx.Where("role_id = ?", id).Delete(&adminModel.RoleApi{})
		return tx.Delete(&adminModel.Role{}, id).Error
	})
}

// AssignRoleMenus 分配菜单
func AssignRoleMenus(id uint, menuIDs []uint) error {
	if id == 1 {
		return errors.New("超级管理员角色不可分配菜单")
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("role_id = ?", id).Delete(&adminModel.RoleMenu{})
		if len(menuIDs) == 0 {
			return nil
		}
		rows := make([]adminModel.RoleMenu, 0, len(menuIDs))
		for _, mid := range menuIDs {
			rows = append(rows, adminModel.RoleMenu{RoleID: id, MenuID: mid})
		}
		return tx.Create(&rows).Error
	})
}

// AssignRoleApis 分配接口
func AssignRoleApis(id uint, apiIDs []uint) error {
	if id == 1 {
		return errors.New("超级管理员角色不可分配接口")
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("role_id = ?", id).Delete(&adminModel.RoleApi{})
		if len(apiIDs) == 0 {
			return nil
		}
		rows := make([]adminModel.RoleApi, 0, len(apiIDs))
		for _, aid := range apiIDs {
			rows = append(rows, adminModel.RoleApi{RoleID: id, ApiID: aid})
		}
		return tx.Create(&rows).Error
	})
}
