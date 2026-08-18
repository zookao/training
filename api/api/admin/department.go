package admin

import (
	"io"

	"training/model/admin"
	"training/model/common"
	adminService "training/service/admin"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// DepartmentList 院系分页
func DepartmentList(c *gin.Context) {
	var req common.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := adminService.DepartmentPage(req)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// DepartmentAll 全部院系
func DepartmentAll(c *gin.Context) {
	list, err := adminService.DepartmentAll()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}

// DepartmentCreate 创建院系
func DepartmentCreate(c *gin.Context) {
	var req admin.DepartmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.DepartmentCreate(req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, nil)
}

// DepartmentUpdate 更新院系
func DepartmentUpdate(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req admin.DepartmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.DepartmentUpdate(id, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, nil)
}

// DepartmentDelete 删除院系
func DepartmentDelete(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.DepartmentDelete(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, nil)
}

// DepartmentStudentList 院系学员列表
func DepartmentStudentList(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	list, err := adminService.DepartmentStudents(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}

// DepartmentStudentRemove 解除院系与学员的绑定
func DepartmentStudentRemove(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	userID, err := common.ParseIDParam(c, "userId")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.DepartmentRemoveStudent(id, userID); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, nil)
}

// DepartmentStudentImport 导入学员到院系
func DepartmentStudentImport(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		common.Fail(c, "请上传文件")
		return
	}
	src, err := file.Open()
	if err != nil {
		common.Fail(c, "文件读取失败")
		return
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		common.Fail(c, "文件读取失败")
		return
	}
	res, err := adminService.DepartmentImportStudents(id, data)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// DepartmentStudentImportTemplate 下载院系学员导入模板
func DepartmentStudentImportTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0)
	_ = f.SetCellValue(sheet, "A1", "手机号")
	_ = f.SetCellValue(sheet, "A2", "13800000000")
	_ = f.SetColWidth(sheet, "A", "A", 22)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="department_student_import_template.xlsx"`)
	buf, err := f.WriteToBuffer()
	if err != nil {
		common.Fail(c, "生成模板失败")
		return
	}
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
