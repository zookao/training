package admin

import (
	"io"

	"github.com/xuri/excelize/v2"

	adminModel "training/model/admin"
	"training/model/common"
	adminService "training/service/admin"

	"github.com/gin-gonic/gin"
)

// UserList 学员列表
func UserList(c *gin.Context) {
	var req common.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	departmentID := common.ParseQueryID(c, "departmentId")
	res, err := adminService.UserPage(req, departmentID)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// UserAll 全部学员（班级绑定用）
func UserAll(c *gin.Context) {
	list, err := adminService.UserAll()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}

// UserDetail 学员详情
func UserDetail(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	u, err := adminService.UserGet(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, u)
}

// UserCreate 创建学员
func UserCreate(c *gin.Context) {
	var req adminModel.UserCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.UserCreate(req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "创建成功")
}

// UserUpdate 更新学员
func UserUpdate(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req adminModel.UserUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.UserUpdate(id, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "更新成功")
}

// UserResetPwd 重置学员密码
func UserResetPwd(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req adminModel.UserResetPwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.UserResetPassword(id, req.Password); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "重置成功")
}

// UserDelete 删除学员
func UserDelete(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.UserDelete(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "删除成功")
}

// UserImport 批量导入学员（Excel）
func UserImport(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		common.Fail(c, "请上传 Excel 文件")
		return
	}
	src, err := file.Open()
	if err != nil {
		common.Fail(c, "读取文件失败")
		return
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		common.Fail(c, "读取文件失败")
		return
	}
	res, err := adminService.UserImport(data)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// UserImportTemplate 下载学员导入 Excel 模板
func UserImportTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0)
	_ = f.SetCellValue(sheet, "A1", "账号")
	_ = f.SetCellValue(sheet, "B1", "手机号")
	_ = f.SetCellValue(sheet, "C1", "姓名")
	_ = f.SetCellValue(sheet, "D1", "学号")
	_ = f.SetCellValue(sheet, "E1", "院系")
	_ = f.SetCellValue(sheet, "A2", "zhangsan")
	_ = f.SetCellValue(sheet, "B2", "13800000000")
	_ = f.SetCellValue(sheet, "C2", "张三")
	_ = f.SetCellValue(sheet, "D2", "20260001")
	_ = f.SetCellValue(sheet, "E2", "计算机学院")
	// 列宽
	for _, col := range []string{"A", "B", "C", "D", "E"} {
		_ = f.SetColWidth(sheet, col, col, 18)
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="user_import_template.xlsx"`)
	buf, err := f.WriteToBuffer()
	if err != nil {
		common.Fail(c, "生成模板失败")
		return
	}
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
