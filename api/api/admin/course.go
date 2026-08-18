package admin

import (
	"io"
	"strconv"

	adminModel "training/model/admin"
	"training/model/common"
	adminService "training/service/admin"
	"training/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// CourseList 课程列表
func CourseList(c *gin.Context) {
	var req common.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	res, err := adminService.CoursePage(req)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// CourseAll 全部课程（班级绑定用）
func CourseAll(c *gin.Context) {
	list, err := adminService.CourseAll()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, list)
}

// CourseDetail 课程详情
func CourseDetail(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	course, err := adminService.CourseGet(id)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, course)
}

// CourseCreate 创建课程
func CourseCreate(c *gin.Context) {
	var req adminModel.CourseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.CourseCreate(req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "创建成功")
}

// CourseUpdate 更新课程
func CourseUpdate(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	var req adminModel.CourseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.CourseUpdate(id, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "更新成功")
}

// CourseDelete 删除课程
func CourseDelete(c *gin.Context) {
	id, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.CourseDelete(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "删除成功")
}

// VideoDelete 删除单个视频
func VideoDelete(c *gin.Context) {
	courseID, err := common.ParseIDParam(c, "id")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	videoID, err := common.ParseIDParam(c, "vid")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := adminService.VideoDelete(courseID, videoID); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "删除成功")
}

// UploadVideo 上传视频（+ 可选缩略图 + 可选课件）
func UploadVideo(c *gin.Context) {
	video, err := c.FormFile("video")
	if err != nil {
		common.Fail(c, "请上传视频文件")
		return
	}
	var thumbnail, _ = c.FormFile("thumbnail")   // 可选，缺失时为 nil
	var courseware, _ = c.FormFile("courseware") // 可选，缺失时为 nil

	url, thumbURL, duration, err := utils.SaveVideo(c, video, thumbnail)
	if err != nil {
		common.Fail(c, "上传失败: "+err.Error())
		return
	}

	// 课件（可选）：支持 PPTX/PPT/ODP/FODP（会同步转 PDF 用于预览，硬性要求 LibreOffice）
	var coursewareURL string
	var coursewarePDFURL string
	var coursewarePageCount int
	if courseware != nil {
		coursewareURL, coursewarePDFURL, coursewarePageCount, err = utils.SaveCourseware(c, courseware)
		if err != nil {
			common.Fail(c, "课件上传失败: "+err.Error())
			return
		}
	}

	common.OK(c, adminModel.UploadVideoRes{
		URL:                 url,
		Thumbnail:           thumbURL,
		Courseware:           coursewareURL,
		CoursewarePageCount: coursewarePageCount,
		CoursewarePDF:        coursewarePDFURL,
		Filename:            video.Filename,
		Duration:            duration,
	})
}

// UploadImage 上传图片（课程/班级封面等）
func UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		common.Fail(c, "请上传图片文件")
		return
	}
	url, err := utils.SaveImage(c, file)
	if err != nil {
		common.Fail(c, "上传失败: "+err.Error())
		return
	}
	common.OK(c, adminModel.UploadImageRes{URL: url})
}

// ReparseCoursewarePages 重新识别视频课件的幻灯片页数
func ReparseCoursewarePages(c *gin.Context) {
	videoID, err := common.ParseIDParam(c, "vid")
	if err != nil {
		common.Fail(c, "参数错误")
		return
	}
	count, err := adminService.ReparseCoursewarePages(videoID)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, gin.H{"coursewarePageCount": count})
}

// ImportDurations 导入课件每页打点，计算每页时长
func ImportDurations(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		common.Fail(c, "请上传 Excel 文件")
		return
	}
	durationStr := c.PostForm("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration <= 0 {
		common.Fail(c, "视频时长无效")
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
	durations, err := adminService.ParseDurations(data, duration)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, gin.H{"durations": durations})
}

// DurationTemplate 下载课件打点导入模板
func DurationTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0)
	_ = f.SetCellValue(sheet, "A1", "页码")
	_ = f.SetCellValue(sheet, "B1", "视频打点")
	_ = f.SetCellValue(sheet, "A2", 1)
	_ = f.SetCellValue(sheet, "B2", "00:00:35")
	_ = f.SetCellValue(sheet, "A3", 2)
	_ = f.SetCellValue(sheet, "B3", "00:03:23")
	_ = f.SetCellValue(sheet, "A4", 3)
	_ = f.SetCellValue(sheet, "B4", "00:11:14")
	_ = f.SetColWidth(sheet, "A", "A", 12)
	_ = f.SetColWidth(sheet, "B", "B", 18)
	c.Header("Content-Disposition", `attachment; filename="duration_template.xlsx"`)
	buf, err := f.WriteToBuffer()
	if err != nil {
		common.Fail(c, "生成模板失败")
		return
	}
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
