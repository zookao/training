package admin

import (
	"errors"
	"strings"

	"training/global"
	adminModel "training/model/admin"
	"training/model/common"
	userModel "training/model/user"

	"github.com/xuri/excelize/v2"
)

// DepartmentPage 院系分页
func DepartmentPage(req common.PageRequest) (*common.PageList, error) {
	req.Normalize()
	var list []adminModel.Department
	var total int64
	q := global.DB.Model(&adminModel.Department{})
	if req.Keyword != "" {
		q = q.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := q.Order("sort asc, id desc").
		Offset(req.Offset()).Limit(req.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &common.PageList{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// DepartmentAll 全部启用院系（下拉选择用）
func DepartmentAll() ([]adminModel.Department, error) {
	var list []adminModel.Department
	if err := global.DB.Where("status = 1").Order("sort asc, id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// DepartmentCreate 创建院系
func DepartmentCreate(req adminModel.DepartmentReq) error {
	d := adminModel.Department{
		Name:        req.Name,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      req.Status,
	}
	return global.DB.Create(&d).Error
}

// DepartmentUpdate 更新院系
func DepartmentUpdate(id uint, req adminModel.DepartmentReq) error {
	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"sort":        req.Sort,
		"status":      req.Status,
	}
	return global.DB.Model(&adminModel.Department{}).Where("id = ?", id).Updates(updates).Error
}

// DepartmentDelete 删除院系
func DepartmentDelete(id uint) error {
	var count int64
	if err := global.DB.Model(&userModel.User{}).Where("department_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该院系下有学员，不允许删除")
	}
	return global.DB.Delete(&adminModel.Department{}, id).Error
}

// DepartmentStudents 院系学员列表
func DepartmentStudents(deptID uint) ([]userModel.User, error) {
	var list []userModel.User
	if err := global.DB.Where("department_id = ?", deptID).Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// DepartmentRemoveStudent 解除院系与学员的绑定
func DepartmentRemoveStudent(deptID, userID uint) error {
	var u userModel.User
	if err := global.DB.First(&u, userID).Error; err != nil {
		return errors.New("学员不存在")
	}
	if u.DepartmentID != deptID {
		return errors.New("该学员不属于此院系")
	}
	return global.DB.Model(&u).Update("department_id", 0).Error
}

// DeptImportRow 导入结果单行
type DeptImportRow struct {
	Row    int    `json:"row"`
	Phone  string `json:"phone"`
	Name   string `json:"name"`
	Result string `json:"result"`
	Reason string `json:"reason"`
}

// DeptImportResult 导入结果汇总
type DeptImportResult struct {
	Total   int             `json:"total"`
	Success int             `json:"success"`
	Failed  int             `json:"failed"`
	Rows    []DeptImportRow `json:"rows"`
}

// DepartmentImportStudents 导入学员到院系（Excel：手机号）
// 规则：学员必须已存在；已绑定其他院系的学员提示失败。
func DepartmentImportStudents(deptID uint, file []byte) (*DeptImportResult, error) {
	f, err := excelize.OpenReader(bytesReader(file))
	if err != nil {
		return nil, errors.New("无法读取 Excel 文件")
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, errors.New("无法读取工作表")
	}

	res := &DeptImportResult{Rows: []DeptImportRow{}}
	if len(rows) <= 1 {
		return res, nil
	}

	// 预建院系 ID→名称 映射（用于错误提示）
	deptNameMap := make(map[uint]string)
	var allDepts []adminModel.Department
	global.DB.Find(&allDepts)
	for _, d := range allDepts {
		deptNameMap[d.ID] = d.Name
	}

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		phone := ""
		if len(row) > 0 {
			phone = strings.TrimSpace(row[0])
		}
		phone = normalizePhone(phone)
		item := DeptImportRow{Row: i + 1, Phone: phone}

		if phone == "" {
			item.Result = "失败"
			item.Reason = "手机号为空"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}

		var u userModel.User
		if err := global.DB.Where("phone = ?", phone).First(&u).Error; err != nil {
			item.Result = "失败"
			item.Reason = "学员不存在"
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}

		item.Name = u.Nickname

		if u.DepartmentID != 0 && u.DepartmentID != deptID {
			item.Result = "失败"
			item.Reason = "已绑定其他院系：" + deptNameMap[u.DepartmentID]
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}

		if u.DepartmentID == deptID {
			item.Result = "成功"
			item.Reason = "已绑定该院系"
			res.Success++
			res.Rows = append(res.Rows, item)
			continue
		}

		// 绑定到当前院系
		if err := global.DB.Model(&u).Update("department_id", deptID).Error; err != nil {
			item.Result = "失败"
			item.Reason = "绑定失败：" + err.Error()
			res.Failed++
			res.Rows = append(res.Rows, item)
			continue
		}
		item.Result = "成功"
		res.Success++
		res.Rows = append(res.Rows, item)
	}
	res.Total = res.Success + res.Failed
	return res, nil
}
