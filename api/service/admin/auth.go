package admin

import (
	"errors"
	"log"
	"time"

	"training/global"
	adminModel "training/model/admin"
	"training/utils"

	"gorm.io/gorm"
)

// Login 管理员登录
func Login(req adminModel.LoginReq, ip string) (string, error) {
	var a adminModel.Admin
	if err := global.DB.Where("username = ?", req.Username).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("账号不存在")
		}
		return "", err
	}
	if a.Status != 1 {
		return "", errors.New("账号已被禁用")
	}
	if !utils.CheckPassword(a.Password, req.Password) {
		return "", errors.New("密码错误")
	}
	token, err := utils.GenerateToken(a.ID, a.Username, "admin",
		global.Config.JWT.SigningKey, global.Config.JWT.Issuer, global.Config.JWT.ExpiresTime)
	if err != nil {
		return "", err
	}
	now := time.Now()
	if err := global.DB.Model(&a).Updates(map[string]interface{}{
		"last_login_at": &now,
		"last_login_ip": ip,
	}).Error; err != nil {
		// 登录追踪为非关键操作，记录日志但不阻断登录流程
		log.Printf("[WARN] 更新管理员登录信息失败 admin_id=%d: %v", a.ID, err)
	}
	return token, nil
}

// UserInfo 当前管理员信息 + 权限码
func UserInfo(userID uint) (*adminModel.UserInfoRes, error) {
	var a adminModel.Admin
	if err := global.DB.First(&a, userID).Error; err != nil {
		return nil, err
	}
	res := &adminModel.UserInfoRes{
		ID:       a.ID,
		Username: a.Username,
		Nickname: a.Nickname,
		Avatar:   a.Avatar,
		Roles:    []string{},
		Perms:    []string{},
	}
	// 超管拥有全部权限码（用 * 表示）
	if a.ID == 1 {
		res.Roles = []string{"super_admin"}
		res.Perms = []string{"*"}
		return res, nil
	}
	// 角色名
	var roles []adminModel.Role
	if err := global.DB.Model(&a).Association("Roles").Find(&roles); err != nil {
		return nil, err
	}
	for _, r := range roles {
		res.Roles = append(res.Roles, r.Name)
	}
	// 权限码：通过 role→menus(F按钮) 的 perms
	var perms []string
	if err := global.DB.Table("menus").
		Joins("JOIN role_menus ON role_menus.menu_id = menus.id").
		Joins("JOIN admin_roles ON admin_roles.role_id = role_menus.role_id").
		Where("admin_roles.admin_id = ? AND menus.type = 'F' AND menus.perms <> ''", a.ID).
		Pluck("DISTINCT menus.perms", &perms).Error; err != nil {
		return nil, err
	}
	res.Perms = perms
	return res, nil
}

// CurrentMenus 当前管理员可见菜单树
func CurrentMenus(userID uint) ([]adminModel.Menu, error) {
	var menus []adminModel.Menu
	if userID == 1 {
		// 超管：全部 admin 端菜单
		if err := global.DB.Where("guard_type = ? AND type IN ?", "admin", []string{"M", "C"}).
			Order("sort asc, id asc").Find(&menus).Error; err != nil {
			return nil, err
		}
	} else {
		if err := global.DB.Where("guard_type = ? AND type IN ?", "admin", []string{"M", "C"}).
			Joins("JOIN role_menus ON role_menus.menu_id = menus.id").
			Joins("JOIN admin_roles ON admin_roles.role_id = role_menus.role_id").
			Where("admin_roles.admin_id = ?", userID).
			Order("sort asc, id asc").Find(&menus).Error; err != nil {
			return nil, err
		}
	}
	return buildMenuTree(menus), nil
}
