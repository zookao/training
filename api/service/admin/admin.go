package admin

import (
	"errors"

	"training/global"
	adminModel "training/model/admin"
	"training/model/common"
	"training/utils"

	"gorm.io/gorm"
)

// AdminPage 管理员分页
func AdminPage(req common.PageRequest) (*common.PageList, error) {
	req.Normalize()
	var list []adminModel.Admin
	var total int64
	q := global.DB.Model(&adminModel.Admin{})
	if req.Keyword != "" {
		q = q.Where("username LIKE ? OR nickname LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	q.Count(&total)
	if err := q.Preload("Roles").Order("id desc").
		Offset(req.Offset()).Limit(req.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &common.PageList{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// AdminGet 详情
func AdminGet(id uint) (*adminModel.Admin, error) {
	var a adminModel.Admin
	if err := global.DB.Preload("Roles").First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// AdminCreate 创建
func AdminCreate(req adminModel.AdminCreateReq) error {
	var exists int64
	global.DB.Model(&adminModel.Admin{}).Where("username = ?", req.Username).Count(&exists)
	if exists > 0 {
		return errors.New("用户名已存在")
	}
	pwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}
	a := adminModel.Admin{
		Username: req.Username,
		Password: pwd,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   req.Status,
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&a).Error; err != nil {
			return err
		}
		return assignAdminRoles(tx, a.ID, req.RoleIDs)
	})
}

// AdminUpdate 更新
func AdminUpdate(id uint, req adminModel.AdminUpdateReq) error {
	if id == 1 {
		return errors.New("超级管理员不可编辑")
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"nickname": req.Nickname,
			"email":    req.Email,
			"phone":    req.Phone,
			"status":   req.Status,
		}
		if err := tx.Model(&adminModel.Admin{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		return assignAdminRoles(tx, id, req.RoleIDs)
	})
}

// AdminResetPassword 重置密码
func AdminResetPassword(id uint, password string) error {
	if id == 1 {
		// return errors.New("超级管理员不可重置密码")
	}
	pwd, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	return global.DB.Model(&adminModel.Admin{}).Where("id = ?", id).
		Update("password", pwd).Error
}

// AdminDelete 删除
func AdminDelete(id uint) error {
	if id == 1 {
		return errors.New("超级管理员不可删除")
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("admin_id = ?", id).Delete(&adminModel.AdminRole{})
		return tx.Delete(&adminModel.Admin{}, id).Error
	})
}

// AssignRoles 分配角色
func AssignRoles(id uint, roleIDs []uint) error {
	if id == 1 {
		return errors.New("超级管理员不可分配角色")
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		return assignAdminRoles(tx, id, roleIDs)
	})
}

func assignAdminRoles(tx *gorm.DB, adminID uint, roleIDs []uint) error {
	tx.Where("admin_id = ?", adminID).Delete(&adminModel.AdminRole{})
	if len(roleIDs) == 0 {
		return nil
	}
	rows := make([]adminModel.AdminRole, 0, len(roleIDs))
	for _, rid := range roleIDs {
		rows = append(rows, adminModel.AdminRole{AdminID: adminID, RoleID: rid})
	}
	return tx.Create(&rows).Error
}
