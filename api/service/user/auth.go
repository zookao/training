package user

import (
	"errors"
	"log"
	"time"

	"training/global"
	userModel "training/model/user"
	"training/utils"

	"gorm.io/gorm"
)

// Register 学员注册
func Register(req userModel.RegisterReq) error {
	phone := normalizePhone(req.Phone)
	if phone == "" {
		return errors.New("请填写手机号")
	}
	if !utils.IsValidPhone(phone) {
		return errors.New("手机号格式不正确")
	}
	var exists int64
	if err := global.DB.Model(&userModel.User{}).Where("username = ?", req.Username).Count(&exists).Error; err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("用户名已存在")
	}
	var phoneExists int64
	if err := global.DB.Model(&userModel.User{}).Where("phone = ?", phone).Count(&phoneExists).Error; err != nil {
		return err
	}
	if phoneExists > 0 {
		return errors.New("手机号已存在")
	}
	// 学号唯一校验（空学号不校验，允许多个空值）
	if req.StudentNo != "" {
		var studentNoExists int64
		if err := global.DB.Model(&userModel.User{}).Where("student_no = ?", req.StudentNo).Count(&studentNoExists).Error; err != nil {
			return err
		}
		if studentNoExists > 0 {
			return errors.New("学号已存在")
		}
	}
	pwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}
	u := userModel.User{
		Username:     req.Username,
		Password:     pwd,
		Nickname:     nickname,
		StudentNo:    req.StudentNo,
		DepartmentID: req.DepartmentID,
		Phone:        phone,
		Status:       1,
	}
	return global.DB.Create(&u).Error
}

// Login 学员登录（支持用户名或手机号）
func Login(req userModel.LoginReq, ip string) (string, error) {
	var u userModel.User
	if err := global.DB.Where("username = ? OR phone = ?", req.Username, req.Username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("账号不存在")
		}
		return "", err
	}
	if u.Status != 1 {
		return "", errors.New("账号已被禁用")
	}
	if !utils.CheckPassword(u.Password, req.Password) {
		return "", errors.New("密码错误")
	}
	token, err := utils.GenerateToken(u.ID, u.Username, "user",
		global.Config.JWT.SigningKey, global.Config.JWT.Issuer, global.Config.JWT.ExpiresTime)
	if err != nil {
		return "", err
	}
	now := time.Now()
	if err := global.DB.Model(&u).Updates(map[string]interface{}{
		"last_login_at": &now,
		"last_login_ip": ip,
	}).Error; err != nil {
		// 登录追踪为非关键操作，记录日志但不阻断登录流程
		log.Printf("[WARN] 更新学员登录信息失败 user_id=%d: %v", u.ID, err)
	}
	return token, nil
}

// normalizePhone 规范化手机号（去空格/横线），空则返回空串
func normalizePhone(phone string) string {
	out := make([]byte, 0, len(phone))
	for i := 0; i < len(phone); i++ {
		c := phone[i]
		if c == ' ' || c == '-' || c == '\t' {
			continue
		}
		out = append(out, c)
	}
	return string(out)
}

// UserInfo 当前学员信息
func UserInfo(userID uint) (*userModel.UserInfoRes, error) {
	var u userModel.User
	if err := global.DB.First(&u, userID).Error; err != nil {
		return nil, err
	}
	return &userModel.UserInfoRes{
		ID:           u.ID,
		Username:     u.Username,
		Nickname:     u.Nickname,
		StudentNo:    u.StudentNo,
		DepartmentID: u.DepartmentID,
		Avatar:       u.Avatar,
		Email:        u.Email,
		Phone:        u.Phone,
	}, nil
}

// UpdateProfile 更新资料
func UpdateProfile(userID uint, req userModel.UpdateProfileReq) error {
	phone := normalizePhone(req.Phone)
	if phone == "" {
		return errors.New("手机号不能为空")
	}
	if !utils.IsValidPhone(phone) {
		return errors.New("手机号格式不正确")
	}
	var phoneExists int64
	if err := global.DB.Model(&userModel.User{}).Where("phone = ? AND id <> ?", phone, userID).Count(&phoneExists).Error; err != nil {
		return err
	}
	if phoneExists > 0 {
		return errors.New("手机号已存在")
	}
	// 学号唯一校验（排除自身；空学号不校验）
	if req.StudentNo != "" {
		var studentNoExists int64
		if err := global.DB.Model(&userModel.User{}).Where("student_no = ? AND id <> ?", req.StudentNo, userID).Count(&studentNoExists).Error; err != nil {
			return err
		}
		if studentNoExists > 0 {
			return errors.New("学号已存在")
		}
	}
	// 注意：department_id 为管理员维护字段（通过 创建/更新学员、院系导入 分配），
	// 学员自助修改资料时不允许改动，避免误清空或越权变更院系归属。
	updates := map[string]interface{}{
		"nickname":   req.Nickname,
		"student_no": req.StudentNo,
		"avatar":     req.Avatar,
		"email":      req.Email,
		"phone":      phone,
	}
	return global.DB.Model(&userModel.User{}).Where("id = ?", userID).Updates(updates).Error
}

// ChangePassword 修改密码
func ChangePassword(userID uint, req userModel.ChangePwdReq) error {
	var u userModel.User
	if err := global.DB.First(&u, userID).Error; err != nil {
		return err
	}
	if !utils.CheckPassword(u.Password, req.OldPassword) {
		return errors.New("原密码错误")
	}
	pwd, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	return global.DB.Model(&u).Update("password", pwd).Error
}
