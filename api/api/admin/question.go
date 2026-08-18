package admin

import (
	"io"

	"training/model/admin"
	"training/model/common"
	adminService "training/service/admin"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// QuestionList 试题列表
func QuestionList(c *gin.Context) {
	courseID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req common.PageRequest
	_ = c.ShouldBindQuery(&req)
	req.Normalize()
	res, err := adminService.QuestionList(courseID, req)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// QuestionAll 课程下全部启用试题
func QuestionAll(c *gin.Context) {
	courseID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	list, err := adminService.QuestionAll(courseID)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}

// QuestionCreate 创建试题
func QuestionCreate(c *gin.Context) {
	courseID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req admin.QuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.QuestionCreate(courseID, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, nil)
}

// QuestionUpdate 更新试题
func QuestionUpdate(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req admin.QuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.QuestionUpdate(id, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, nil)
}

// QuestionDelete 删除试题
func QuestionDelete(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.QuestionDelete(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, nil)
}

// QuestionImport 批量导入试题（Excel）
func QuestionImport(c *gin.Context) {
	courseID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
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
	res, err := adminService.QuestionImport(courseID, data)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// QuestionImportTemplate 下载试题导入 Excel 模板
func QuestionImportTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0)
	headers := []string{"题型", "题干", "选项A", "选项B", "选项C", "选项D", "正确答案", "解析"}
	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetCellValue(sheet, col+"1", h)
	}
	// 示例行
	_ = f.SetCellValue(sheet, "A2", "单选")
	_ = f.SetCellValue(sheet, "B2", "以下哪个是正确的？")
	_ = f.SetCellValue(sheet, "C2", "选项一")
	_ = f.SetCellValue(sheet, "D2", "选项二")
	_ = f.SetCellValue(sheet, "E2", "选项三")
	_ = f.SetCellValue(sheet, "F2", "选项四")
	_ = f.SetCellValue(sheet, "G2", "A")
	_ = f.SetCellValue(sheet, "H2", "解析内容")
	_ = f.SetCellValue(sheet, "A3", "多选")
	_ = f.SetCellValue(sheet, "B3", "以下哪些是正确的？")
	_ = f.SetCellValue(sheet, "C3", "选项一")
	_ = f.SetCellValue(sheet, "D3", "选项二")
	_ = f.SetCellValue(sheet, "E3", "选项三")
	_ = f.SetCellValue(sheet, "G3", "AC")
	_ = f.SetCellValue(sheet, "A4", "判断")
	_ = f.SetCellValue(sheet, "B4", "地球是圆的。")
	_ = f.SetCellValue(sheet, "C4", "正确")
	_ = f.SetCellValue(sheet, "D4", "错误")
	_ = f.SetCellValue(sheet, "G4", "A")
	for i := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetColWidth(sheet, col, col, 18)
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="question_import_template.xlsx"`)
	buf, err := f.WriteToBuffer()
	if err != nil {
		common.Fail(c, "生成模板失败")
		return
	}
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
