package admin

import (
	"errors"

	"training/global"
	adminModel "training/model/admin"
)

// MenuList 全部菜单树（管理用）
func MenuList(guard string) ([]adminModel.Menu, error) {
	var menus []adminModel.Menu
	q := global.DB.Order("sort asc, id asc")
	if guard != "" {
		q = q.Where("guard_type = ?", guard)
	}
	if err := q.Find(&menus).Error; err != nil {
		return nil, err
	}
	return buildMenuTree(menus), nil
}

// MenuGet 详情
func MenuGet(id uint) (*adminModel.Menu, error) {
	var m adminModel.Menu
	if err := global.DB.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// MenuCreate 创建
func MenuCreate(req adminModel.MenuReq) error {
	m := adminModel.Menu{
		ParentID:  req.ParentID,
		Name:      req.Name,
		Type:      req.Type,
		Path:      req.Path,
		Component: req.Component,
		Redirect:  req.Redirect,
		Icon:      req.Icon,
		Hidden:    req.Hidden,
		KeepAlive: req.KeepAlive,
		Sort:      req.Sort,
		GuardType: defaultGuard(req.GuardType),
		Perms:     req.Perms,
	}
	return global.DB.Create(&m).Error
}

// MenuUpdate 更新
func MenuUpdate(id uint, req adminModel.MenuReq) error {
	if req.ParentID == id {
		return errors.New("上级菜单不能选择自己")
	}
	updates := map[string]interface{}{
		"parent_id":  req.ParentID,
		"name":       req.Name,
		"type":       req.Type,
		"path":       req.Path,
		"component":  req.Component,
		"redirect":   req.Redirect,
		"icon":       req.Icon,
		"hidden":     req.Hidden,
		"keep_alive": req.KeepAlive,
		"sort":       req.Sort,
		"guard_type": defaultGuard(req.GuardType),
		"perms":      req.Perms,
	}
	return global.DB.Model(&adminModel.Menu{}).Where("id = ?", id).Updates(updates).Error
}

// MenuDelete 删除（含子菜单）
func MenuDelete(id uint) error {
	var count int64
	global.DB.Model(&adminModel.Menu{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("请先删除子菜单")
	}
	return global.DB.Delete(&adminModel.Menu{}, id).Error
}

// buildMenuTree 构建菜单树
func buildMenuTree(menus []adminModel.Menu) []adminModel.Menu {
	byID := make(map[uint]*adminModel.Menu, len(menus))
	roots := make([]adminModel.Menu, 0)
	// 先建立索引
	ptrs := make([]*adminModel.Menu, len(menus))
	for i := range menus {
		ptrs[i] = &menus[i]
		byID[menus[i].ID] = ptrs[i]
	}
	for i := range menus {
		m := ptrs[i]
		if m.ParentID == 0 {
			roots = append(roots, *m)
		} else if parent, ok := byID[m.ParentID]; ok {
			parent.Children = append(parent.Children, *m)
		} else {
			roots = append(roots, *m)
		}
	}
	return roots
}

func defaultGuard(g string) string {
	if g == "" {
		return "admin"
	}
	return g
}
