package admin

import (
	"errors"
	"strings"

	"training/global"
	adminModel "training/model/admin"
	"training/model/common"
	userModel "training/model/user"
	"training/utils"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// normalizePhone 规范化手机号（去空格/横线）
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

// UserPage 学员分页（departmentID > 0 时按院系过滤）
func UserPage(req common.PageRequest, departmentID uint) (*common.PageList, error) {
	req.Normalize()
	var list []userModel.User
	var total int64
	q := global.DB.Model(&userModel.User{})
	if req.Keyword != "" {
		q = q.Where("username LIKE ? OR nickname LIKE ? OR phone LIKE ? OR student_no LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if departmentID > 0 {
		q = q.Where("department_id = ?", departmentID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := q.Order("id desc").
		Offset(req.Offset()).Limit(req.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &common.PageList{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// UserGet 学员详情
func UserGet(id uint) (*userModel.User, error) {
	var u userModel.User
	if err := global.DB.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// UserAll 全部学员（班级绑定用）
func UserAll() ([]userModel.User, error) {
	var list []userModel.User
	if err := global.DB.Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// UserCreate 创建学员
func UserCreate(req adminModel.UserCreateReq) error {
	var exists int64
	if err := global.DB.Model(&userModel.User{}).Where("username = ?", req.Username).Count(&exists).Error; err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("用户名已存在")
	}
	phone := normalizePhone(req.Phone)
	if phone == "" {
		return errors.New("手机号不能为空")
	}
	if !utils.IsValidPhone(phone) {
		return errors.New("手机号格式不正确")
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
		Email:        req.Email,
		Phone:        phone,
		Status:       req.Status,
	}
	return global.DB.Create(&u).Error
}

// UserUpdate 更新学员
func UserUpdate(id uint, req adminModel.UserUpdateReq) error {
	phone := normalizePhone(req.Phone)
	if phone == "" {
		return errors.New("手机号不能为空")
	}
	if !utils.IsValidPhone(phone) {
		return errors.New("手机号格式不正确")
	}
	var phoneExists int64
	if err := global.DB.Model(&userModel.User{}).Where("phone = ? AND id <> ?", phone, id).Count(&phoneExists).Error; err != nil {
		return err
	}
	if phoneExists > 0 {
		return errors.New("手机号已存在")
	}
	// 学号唯一校验（排除自身；空学号不校验）
	if req.StudentNo != "" {
		var studentNoExists int64
		if err := global.DB.Model(&userModel.User{}).Where("student_no = ? AND id <> ?", req.StudentNo, id).Count(&studentNoExists).Error; err != nil {
			return err
		}
		if studentNoExists > 0 {
			return errors.New("学号已存在")
		}
	}
	updates := map[string]interface{}{
		"nickname":      req.Nickname,
		"student_no":    req.StudentNo,
		"department_id": req.DepartmentID,
		"email":         req.Email,
		"phone":         phone,
		"status":        req.Status,
	}
	return global.DB.Model(&userModel.User{}).Where("id = ?", id).Updates(updates).Error
}

// UserResetPassword 重置学员密码
func UserResetPassword(id uint, password string) error {
	pwd, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	return global.DB.Model(&userModel.User{}).Where("id = ?", id).
		Update("password", pwd).Error
}

// UserDelete 删除学员
func UserDelete(id uint) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&adminModel.ClassUser{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&userModel.VideoRecord{}).Error; err != nil {
			return err
		}
		return tx.Delete(&userModel.User{}, id).Error
	})
}

// UserImportRow 导入结果单行
type UserImportRow struct {
	Row        int    `json:"row"`        // Excel 行号（从 2 开始，1 为表头）
	Username   string `json:"username"`   // 账号
	Phone      string `json:"phone"`      // 手机号
	Name       string `json:"name"`       // 姓名
	StudentNo  string `json:"studentNo"`  // 学号
	Department string `json:"department"` // 院系名称
	Success    bool   `json:"success"`    // 是否成功
	Reason     string `json:"reason"`     // 失败原因
}

// UserImportResult 导入结果汇总
type UserImportResult struct {
	Total   int             `json:"total"`   // 总行数（不含表头）
	Success int             `json:"success"` // 成功数
	Failed  int             `json:"failed"`  // 失败数
	Rows    []UserImportRow `json:"rows"`    // 每行明细
}

// UserImport 批量导入学员（Excel：账号、手机号、姓名、学号、院系）
// 规则：账号和手机号必填，初始密码 = 手机号，状态启用。院系填了须匹配已存在的启用院系。
func UserImport(file []byte) (*UserImportResult, error) {
	f, err := excelize.OpenReader(bytesReader(file))
	if err != nil {
		return nil, errors.New("无法读取 Excel 文件")
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, errors.New("无法读取工作表")
	}

	res := &UserImportResult{Rows: []UserImportRow{}}
	if len(rows) == 0 {
		return res, nil
	}

	// 预查所有已用手机号/用户名/学号，避免逐行查询
	usedPhone := make(map[string]bool)
	usedUsername := make(map[string]bool)
	usedStudentNo := make(map[string]bool)
	var existingPhones, existingUsernames, existingStudentNos []string
	if err := global.DB.Model(&userModel.User{}).Where("phone <> ''").Pluck("phone", &existingPhones).Error; err != nil {
		return nil, err
	}
	if err := global.DB.Model(&userModel.User{}).Pluck("username", &existingUsernames).Error; err != nil {
		return nil, err
	}
	if err := global.DB.Model(&userModel.User{}).Where("student_no <> ''").Pluck("student_no", &existingStudentNos).Error; err != nil {
		return nil, err
	}
	for _, p := range existingPhones {
		usedPhone[p] = true
	}
	for _, u := range existingUsernames {
		usedUsername[u] = true
	}
	for _, s := range existingStudentNos {
		usedStudentNo[s] = true
	}

	// 预查所有启用院系，建立名称→ID 映射
	deptMap := make(map[string]uint)
	var depts []adminModel.Department
	if err := global.DB.Where("status = 1").Find(&depts).Error; err != nil {
		return nil, err
	}
	for _, d := range depts {
		deptMap[d.Name] = d.ID
	}

	startIdx := 1 // 跳过表头
	if len(rows) == 1 {
		return res, nil
	}
	for i := startIdx; i < len(rows); i++ {
		row := rows[i]
		username := ""
		phone := ""
		name := ""
		studentNo := ""
		department := ""
		if len(row) > 0 {
			username = strings.TrimSpace(row[0])
		}
		if len(row) > 1 {
			phone = strings.TrimSpace(row[1])
		}
		if len(row) > 2 {
			name = strings.TrimSpace(row[2])
		}
		if len(row) > 3 {
			studentNo = strings.TrimSpace(row[3])
		}
		if len(row) > 4 {
			department = strings.TrimSpace(row[4])
		}
		phone = normalizePhone(phone)
		item := UserImportRow{Row: i + 1, Username: username, Phone: phone, Name: name, StudentNo: studentNo, Department: department}

		if username == "" {
			item.Reason = "账号为空"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		if phone == "" {
			item.Reason = "手机号为空"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		if !utils.IsValidPhone(phone) {
			item.Reason = "手机号格式不正确"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		if usedUsername[username] {
			item.Reason = "账号已存在"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		if usedPhone[phone] {
			item.Reason = "手机号已存在"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		// 学号校验：填了就必须唯一（与库内已有及本批次已导入的均不能重复）
		if studentNo != "" && usedStudentNo[studentNo] {
			item.Reason = "学号已存在"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		// 院系校验：填了就必须匹配已存在的启用院系
		var deptID uint
		if department != "" {
			id, ok := deptMap[department]
			if !ok {
				item.Reason = "院系不存在"
				res.Failed++
				res.Rows = append(res.Rows, item)
				continue
			}
			deptID = id
		}
		pwd, err := utils.HashPassword(phone)
		if err != nil {
			item.Reason = "密码加密失败"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		nickname := name
		if nickname == "" {
			nickname = username
		}
		u := userModel.User{
			Username:     username,
			Password:     pwd,
			Nickname:     nickname,
			StudentNo:    studentNo,
			DepartmentID: deptID,
			Phone:        phone,
			Status:       1,
		}
		if err := global.DB.Create(&u).Error; err != nil {
			item.Reason = "创建失败：" + err.Error()
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		usedPhone[phone] = true
		usedUsername[username] = true
		if studentNo != "" {
			usedStudentNo[studentNo] = true
		}
		item.Success = true
		res.Success++
		res.Rows = append(res.Rows, item)
	}
	res.Total = res.Success + res.Failed
	return res, nil
}

// bytesReader 将 []byte 转为 io.Reader
func bytesReader(b []byte) *strings.Reader {
	return strings.NewReader(string(b))
}
