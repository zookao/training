package initialize

import (
	"log"

	"training/model/admin"
	"training/model/user"
	"training/utils"

	"gorm.io/gorm"
)

// Migrate 自动建表
func Migrate(db *gorm.DB) {
	if err := db.AutoMigrate(
		&admin.Admin{},
		&user.User{},
		&admin.Role{},
		&admin.Menu{},
		&admin.Api{},
		&admin.AdminRole{},
		&admin.UserRole{},
		&admin.RoleMenu{},
		&admin.RoleApi{},
		&admin.Course{},
		&admin.Video{},
		&admin.Class{},
		&admin.ClassCourse{},
		&admin.ClassUser{},
		&admin.Department{},
		&admin.Question{},
		&admin.Testpaper{},
		&admin.TestpaperQuestion{},
		&user.VideoRecord{},
		&user.CheckLog{},
		&user.TestpaperRecord{},
		&admin.Setting{},
	); err != nil {
		log.Fatalf("[Migrate] 迁移失败: %v", err)
	}
	log.Println("[Migrate] 表结构同步完成")
}

// Seed 种子数据
func Seed(db *gorm.DB) {
	seedSuperAdmin(db)
	seedMenus(db)
	seedApis(db)
	seedCourseMenus(db)
	seedCourseApis(db)
	seedUserMenus(db)
	seedClassMenus(db)
	seedUserApis(db)
	seedClassApis(db)
	seedUserImportButton(db)
	migrateUserMenu(db)
	seedClassReportButton(db)
	seedDepartmentMenus(db)
	seedDepartmentApis(db)
	seedExamApis(db)
	migrateDepartmentMenu(db)
	seedSettingMenu(db)
	seedSettingApi(db)
}

func seedSuperAdmin(db *gorm.DB) {
	var count int64
	db.Model(&admin.Admin{}).Count(&count)
	if count > 0 {
		return
	}
	pwd, _ := utils.HashPassword("admin123")
	a := admin.Admin{
		Username: "admin",
		Password: pwd,
		Nickname: "超级管理员",
		Status:   1,
	}
	if err := db.Create(&a).Error; err != nil {
		log.Println("[Seed] 超管创建失败:", err)
		return
	}
	// 超管角色
	r := admin.Role{
		Name:      "super_admin",
		Title:     "超级管理员",
		GuardType: "admin",
		Sort:      0,
		Status:    1,
	}
	if err := db.Create(&r).Error; err == nil {
		db.Create(&admin.AdminRole{AdminID: a.ID, RoleID: r.ID})
	}
	log.Println("[Seed] 超管已创建: admin / admin123")
}

func seedMenus(db *gorm.DB) {
	var count int64
	db.Model(&admin.Menu{}).Count(&count)
	if count > 0 {
		return
	}
	menus := []admin.Menu{
		{Name: "系统管理", Type: "M", Path: "/system", Icon: "Setting", Sort: 90, GuardType: "admin"},
	}
	if err := db.Create(&menus[0]).Error; err != nil {
		log.Println("[Seed] 菜单创建失败:", err)
		return
	}
	sysID := menus[0].ID
	children := []admin.Menu{
		{ParentID: sysID, Name: "管理员", Type: "C", Path: "admin", Component: "system/admin/index", Icon: "User", Sort: 1, GuardType: "admin", Perms: "admin:list"},
		{ParentID: sysID, Name: "角色管理", Type: "C", Path: "role", Component: "system/role/index", Icon: "UserFilled", Sort: 2, GuardType: "admin", Perms: "role:list"},
		{ParentID: sysID, Name: "菜单管理", Type: "C", Path: "menu", Component: "system/menu/index", Icon: "Menu", Sort: 3, GuardType: "admin", Perms: "menu:list"},
		{ParentID: sysID, Name: "接口管理", Type: "C", Path: "api", Component: "system/api/index", Icon: "Connection", Sort: 4, GuardType: "admin", Perms: "api:list"},
	}
	db.Create(&children)
	// 按钮权限
	buttons := []admin.Menu{
		{ParentID: children[0].ID, Name: "新增", Type: "F", Perms: "admin:add", GuardType: "admin"},
		{ParentID: children[0].ID, Name: "编辑", Type: "F", Perms: "admin:edit", GuardType: "admin"},
		{ParentID: children[0].ID, Name: "删除", Type: "F", Perms: "admin:delete", GuardType: "admin"},
		{ParentID: children[1].ID, Name: "新增", Type: "F", Perms: "role:add", GuardType: "admin"},
		{ParentID: children[1].ID, Name: "编辑", Type: "F", Perms: "role:edit", GuardType: "admin"},
		{ParentID: children[1].ID, Name: "删除", Type: "F", Perms: "role:delete", GuardType: "admin"},
		{ParentID: children[2].ID, Name: "新增", Type: "F", Perms: "menu:add", GuardType: "admin"},
		{ParentID: children[2].ID, Name: "编辑", Type: "F", Perms: "menu:edit", GuardType: "admin"},
		{ParentID: children[2].ID, Name: "删除", Type: "F", Perms: "menu:delete", GuardType: "admin"},
		{ParentID: children[3].ID, Name: "新增", Type: "F", Perms: "api:add", GuardType: "admin"},
		{ParentID: children[3].ID, Name: "编辑", Type: "F", Perms: "api:edit", GuardType: "admin"},
		{ParentID: children[3].ID, Name: "删除", Type: "F", Perms: "api:delete", GuardType: "admin"},
	}
	db.Create(&buttons)
	log.Println("[Seed] 默认菜单已创建")
}

func seedApis(db *gorm.DB) {
	var count int64
	db.Model(&admin.Api{}).Count(&count)
	if count > 0 {
		return
	}
	apis := []admin.Api{
		{Group: "管理员", Method: "GET", Path: "/api/admin/admin", Description: "管理员列表"},
		{Group: "管理员", Method: "POST", Path: "/api/admin/admin", Description: "创建管理员"},
		{Group: "管理员", Method: "GET", Path: "/api/admin/admin/:id", Description: "管理员详情"},
		{Group: "管理员", Method: "PUT", Path: "/api/admin/admin/:id", Description: "更新管理员"},
		{Group: "管理员", Method: "PUT", Path: "/api/admin/admin/:id/roles", Description: "分配管理员角色"},
		{Group: "管理员", Method: "PUT", Path: "/api/admin/admin/:id/password", Description: "重置管理员密码"},
		{Group: "管理员", Method: "DELETE", Path: "/api/admin/admin/:id", Description: "删除管理员"},
		{Group: "角色", Method: "GET", Path: "/api/admin/role", Description: "角色列表"},
		{Group: "角色", Method: "POST", Path: "/api/admin/role", Description: "创建角色"},
		{Group: "角色", Method: "GET", Path: "/api/admin/role/:id", Description: "角色详情"},
		{Group: "角色", Method: "PUT", Path: "/api/admin/role/:id", Description: "更新角色"},
		{Group: "角色", Method: "DELETE", Path: "/api/admin/role/:id", Description: "删除角色"},
		{Group: "角色", Method: "PUT", Path: "/api/admin/role/:id/menus", Description: "分配角色菜单"},
		{Group: "角色", Method: "PUT", Path: "/api/admin/role/:id/apis", Description: "分配角色接口"},
		{Group: "菜单", Method: "GET", Path: "/api/admin/menu", Description: "菜单树"},
		{Group: "菜单", Method: "POST", Path: "/api/admin/menu", Description: "创建菜单"},
		{Group: "菜单", Method: "GET", Path: "/api/admin/menu/:id", Description: "菜单详情"},
		{Group: "菜单", Method: "PUT", Path: "/api/admin/menu/:id", Description: "更新菜单"},
		{Group: "菜单", Method: "DELETE", Path: "/api/admin/menu/:id", Description: "删除菜单"},
		{Group: "接口", Method: "GET", Path: "/api/admin/api", Description: "接口列表"},
		{Group: "接口", Method: "POST", Path: "/api/admin/api", Description: "创建接口"},
		{Group: "接口", Method: "PUT", Path: "/api/admin/api/:id", Description: "更新接口"},
		{Group: "接口", Method: "DELETE", Path: "/api/admin/api/:id", Description: "删除接口"},
	}
	db.Create(&apis)
	log.Println("[Seed] 默认接口已创建")
}

// seedCourseMenus 课程菜单（幂等：按名称判断，不存在才插入）
func seedCourseMenus(db *gorm.DB) {
	var count int64
	db.Model(&admin.Menu{}).Where("name = ? AND guard_type = ?", "课程管理", "admin").Count(&count)
	if count > 0 {
		return
	}
	top := admin.Menu{Name: "课程管理", Type: "M", Path: "/course", Icon: "VideoCamera", Sort: 10, GuardType: "admin"}
	if err := db.Create(&top).Error; err != nil {
		log.Println("[Seed] 课程菜单创建失败:", err)
		return
	}
	child := admin.Menu{
		ParentID:  top.ID,
		Name:      "课程列表",
		Type:      "C",
		Path:      "list",
		Component: "course/index",
		Icon:      "Film",
		Sort:      1,
		GuardType: "admin",
		Perms:     "course:list",
	}
	if err := db.Create(&child).Error; err != nil {
		return
	}
	buttons := []admin.Menu{
		{ParentID: child.ID, Name: "新增", Type: "F", Perms: "course:add", GuardType: "admin"},
		{ParentID: child.ID, Name: "编辑", Type: "F", Perms: "course:edit", GuardType: "admin"},
		{ParentID: child.ID, Name: "删除", Type: "F", Perms: "course:delete", GuardType: "admin"},
	}
	db.Create(&buttons)
	log.Println("[Seed] 课程菜单已创建")
}

// seedCourseApis 课程接口权限点（幂等：按路径判断）
func seedCourseApis(db *gorm.DB) {
	var count int64
	db.Model(&admin.Api{}).Where("path = ?", "/api/admin/course").Count(&count)
	if count > 0 {
		// 上传图片/分片上传改为免接口权限，清理旧的 perm 记录
		db.Where("path = ?", "/api/admin/upload/image").Delete(&admin.Api{})
		db.Where("path LIKE ?", "/api/admin/upload/chunk%").Delete(&admin.Api{})
		// 兼容旧库：补充后续新增的接口权限点
		seedApiIfMissing(db, admin.Api{Group: "课程", Method: "POST", Path: "/api/admin/course/video/:vid/reparse", Description: "重新识别课件页数"})
		return
	}
	apis := []admin.Api{
		{Group: "课程", Method: "GET", Path: "/api/admin/course", Description: "课程列表"},
		{Group: "课程", Method: "POST", Path: "/api/admin/course", Description: "创建课程"},
		{Group: "课程", Method: "GET", Path: "/api/admin/course/:id", Description: "课程详情"},
		{Group: "课程", Method: "PUT", Path: "/api/admin/course/:id", Description: "更新课程"},
		{Group: "课程", Method: "DELETE", Path: "/api/admin/course/:id", Description: "删除课程"},
		{Group: "课程", Method: "DELETE", Path: "/api/admin/course/:id/video/:vid", Description: "删除视频"},
		{Group: "课程", Method: "POST", Path: "/api/admin/course/video/:vid/reparse", Description: "重新识别课件页数"},
	}
	db.Create(&apis)
	log.Println("[Seed] 课程接口已创建")
}

// seedApiIfMissing 按路径+方法幂等插入单个接口权限点
func seedApiIfMissing(db *gorm.DB, a admin.Api) {
	var count int64
	db.Model(&admin.Api{}).Where("method = ? AND path = ?", a.Method, a.Path).Count(&count)
	if count == 0 {
		db.Create(&a)
	}
}

// seedDepartmentMenus 院系管理菜单（幂等：顶级菜单，位于用户管理之下）
func seedDepartmentMenus(db *gorm.DB) {
	var count int64
	db.Model(&admin.Menu{}).Where("name = ? AND guard_type = ?", "院系管理", "admin").Count(&count)
	if count > 0 {
		return
	}
	m := admin.Menu{
		Name:      "院系管理",
		Type:      "C",
		Path:      "/department",
		Component: "department/index",
		Icon:      "OfficeBuilding",
		Sort:      9,
		GuardType: "admin",
		Perms:     "department:list",
	}
	if err := db.Create(&m).Error; err != nil {
		return
	}
	buttons := []admin.Menu{
		{ParentID: m.ID, Name: "新增", Type: "F", Perms: "department:add", GuardType: "admin"},
		{ParentID: m.ID, Name: "编辑", Type: "F", Perms: "department:edit", GuardType: "admin"},
		{ParentID: m.ID, Name: "删除", Type: "F", Perms: "department:delete", GuardType: "admin"},
	}
	db.Create(&buttons)
	log.Println("[Seed] 院系菜单已创建")
}

// migrateDepartmentMenu 迁移：将旧版挂在「系统管理」下的院系管理菜单提升为顶级菜单（兼容已部署库）
func migrateDepartmentMenu(db *gorm.DB) {
	var m admin.Menu
	if err := db.Where("name = ? AND guard_type = ?", "院系管理", "admin").First(&m).Error; err != nil {
		return
	}
	if m.ParentID == 0 && m.Path == "/department" {
		return // 已是顶级，无需迁移
	}
	db.Model(&m).Updates(map[string]interface{}{
		"parent_id": 0,
		"path":      "/department",
		"component": "department/index",
		"type":      "C",
		"sort":      9,
	})
	log.Println("[Migrate] 院系管理菜单已提升为顶级")
}

// seedDepartmentApis 院系接口权限点（幂等）
func seedDepartmentApis(db *gorm.DB) {
	apis := []admin.Api{
		{Group: "院系", Method: "GET", Path: "/api/admin/department", Description: "院系列表"},
		{Group: "院系", Method: "GET", Path: "/api/admin/department/all", Description: "全部院系"},
		{Group: "院系", Method: "POST", Path: "/api/admin/department", Description: "创建院系"},
		{Group: "院系", Method: "PUT", Path: "/api/admin/department/:id", Description: "更新院系"},
		{Group: "院系", Method: "DELETE", Path: "/api/admin/department/:id", Description: "删除院系"},
		{Group: "院系", Method: "GET", Path: "/api/admin/department/:id/students", Description: "院系学员列表"},
		{Group: "院系", Method: "POST", Path: "/api/admin/department/:id/students/import", Description: "导入学员到院系"},
		{Group: "院系", Method: "DELETE", Path: "/api/admin/department/:id/students/:userId", Description: "移除院系学员"},
	}
	for _, a := range apis {
		seedApiIfMissing(db, a)
	}
	log.Println("[Seed] 院系接口已创建")
}

// seedUserMenus 用户管理菜单（幂等：顶级菜单）
func seedUserMenus(db *gorm.DB) {
	var count int64
	db.Model(&admin.Menu{}).Where("name = ? AND guard_type = ?", "用户管理", "admin").Count(&count)
	if count > 0 {
		return
	}
	m := admin.Menu{
		Name:      "用户管理",
		Type:      "C",
		Path:      "/user",
		Component: "user/index",
		Icon:      "User",
		Sort:      8,
		GuardType: "admin",
		Perms:     "user:list",
	}
	if err := db.Create(&m).Error; err != nil {
		return
	}
	buttons := []admin.Menu{
		{ParentID: m.ID, Name: "新增", Type: "F", Perms: "user:add", GuardType: "admin"},
		{ParentID: m.ID, Name: "编辑", Type: "F", Perms: "user:edit", GuardType: "admin"},
		{ParentID: m.ID, Name: "删除", Type: "F", Perms: "user:delete", GuardType: "admin"},
		{ParentID: m.ID, Name: "导入", Type: "F", Perms: "user:import", GuardType: "admin"},
	}
	db.Create(&buttons)
	log.Println("[Seed] 用户管理菜单已创建")
}

// migrateUserMenu 迁移：将旧版挂在「系统管理」下的用户管理菜单提升为顶级菜单（兼容已部署库）
func migrateUserMenu(db *gorm.DB) {
	var m admin.Menu
	if err := db.Where("name = ? AND guard_type = ?", "用户管理", "admin").First(&m).Error; err != nil {
		return
	}
	if m.ParentID == 0 && m.Path == "/user" {
		return // 已是顶级，无需迁移
	}
	db.Model(&m).Updates(map[string]interface{}{
		"parent_id": 0,
		"path":      "/user",
		"component": "user/index",
		"type":      "C",
		"sort":      8,
	})
	log.Println("[Migrate] 用户管理菜单已提升为顶级")
}

// seedClassMenus 班级管理菜单（幂等）
func seedClassMenus(db *gorm.DB) {
	var count int64
	db.Model(&admin.Menu{}).Where("name = ? AND guard_type = ?", "班级管理", "admin").Count(&count)
	if count > 0 {
		return
	}
	top := admin.Menu{Name: "班级管理", Type: "M", Path: "/class", Icon: "School", Sort: 20, GuardType: "admin"}
	if err := db.Create(&top).Error; err != nil {
		log.Println("[Seed] 班级菜单创建失败:", err)
		return
	}
	child := admin.Menu{
		ParentID:  top.ID,
		Name:      "班级列表",
		Type:      "C",
		Path:      "list",
		Component: "class/index",
		Icon:      "Memo",
		Sort:      1,
		GuardType: "admin",
		Perms:     "class:list",
	}
	if err := db.Create(&child).Error; err != nil {
		return
	}
	buttons := []admin.Menu{
		{ParentID: child.ID, Name: "新增", Type: "F", Perms: "class:add", GuardType: "admin"},
		{ParentID: child.ID, Name: "编辑", Type: "F", Perms: "class:edit", GuardType: "admin"},
		{ParentID: child.ID, Name: "删除", Type: "F", Perms: "class:delete", GuardType: "admin"},
	}
	db.Create(&buttons)
	log.Println("[Seed] 班级管理菜单已创建")
}

// seedUserApis 学员接口权限点（幂等）
func seedUserApis(db *gorm.DB) {
	var count int64
	db.Model(&admin.Api{}).Where("path = ?", "/api/admin/user").Count(&count)
	if count > 0 {
		// 兼容旧库：补充后续新增的导入接口权限点
		seedApiIfMissing(db, admin.Api{Group: "学员", Method: "POST", Path: "/api/admin/user/import", Description: "批量导入学员"})
		return
	}
	apis := []admin.Api{
		{Group: "学员", Method: "GET", Path: "/api/admin/user", Description: "学员列表"},
		{Group: "学员", Method: "GET", Path: "/api/admin/user/all", Description: "全部学员"},
		{Group: "学员", Method: "GET", Path: "/api/admin/user/:id", Description: "学员详情"},
		{Group: "学员", Method: "POST", Path: "/api/admin/user", Description: "创建学员"},
		{Group: "学员", Method: "POST", Path: "/api/admin/user/import", Description: "批量导入学员"},
		{Group: "学员", Method: "PUT", Path: "/api/admin/user/:id", Description: "更新学员"},
		{Group: "学员", Method: "PUT", Path: "/api/admin/user/:id/password", Description: "重置学员密码"},
		{Group: "学员", Method: "DELETE", Path: "/api/admin/user/:id", Description: "删除学员"},
	}
	db.Create(&apis)
	log.Println("[Seed] 学员接口已创建")
}

// seedUserImportButton 学员管理「导入」按钮（幂等）
func seedUserImportButton(db *gorm.DB) {
	var count int64
	db.Model(&admin.Menu{}).Where("perms = ? AND guard_type = ?", "user:import", "admin").Count(&count)
	if count > 0 {
		return
	}
	var userMenu admin.Menu
	if err := db.Where("name = ? AND guard_type = ?", "用户管理", "admin").First(&userMenu).Error; err != nil {
		return
	}
	db.Create(&admin.Menu{
		ParentID:  userMenu.ID,
		Name:      "导入",
		Type:      "F",
		Perms:     "user:import",
		GuardType: "admin",
	})
}

// seedClassApis 班级接口权限点（幂等）
func seedClassApis(db *gorm.DB) {
	var count int64
	db.Model(&admin.Api{}).Where("path = ?", "/api/admin/class").Count(&count)
	if count > 0 {
		// 兼容旧库：补充后续新增的学习报告接口权限点
		seedApiIfMissing(db, admin.Api{Group: "班级", Method: "GET", Path: "/api/admin/class/:id/learning-report", Description: "班级学习报告"})
		seedApiIfMissing(db, admin.Api{Group: "班级", Method: "GET", Path: "/api/admin/class/:id/learning-report/:userId", Description: "学员学习详情"})
		return
	}
	apis := []admin.Api{
		{Group: "班级", Method: "GET", Path: "/api/admin/class", Description: "班级列表"},
		{Group: "班级", Method: "POST", Path: "/api/admin/class", Description: "创建班级"},
		{Group: "班级", Method: "GET", Path: "/api/admin/class/:id", Description: "班级详情"},
		{Group: "班级", Method: "PUT", Path: "/api/admin/class/:id", Description: "更新班级"},
		{Group: "班级", Method: "DELETE", Path: "/api/admin/class/:id", Description: "删除班级"},
		{Group: "班级", Method: "GET", Path: "/api/admin/class/:id/courseIds", Description: "班级课程ID"},
		{Group: "班级", Method: "GET", Path: "/api/admin/class/:id/userIds", Description: "班级学员ID"},
		{Group: "班级", Method: "PUT", Path: "/api/admin/class/:id/courses", Description: "分配班级课程"},
		{Group: "班级", Method: "PUT", Path: "/api/admin/class/:id/users", Description: "分配班级学员"},
		{Group: "班级", Method: "GET", Path: "/api/admin/class/:id/learning-report", Description: "班级学习报告"},
		{Group: "班级", Method: "GET", Path: "/api/admin/class/:id/learning-report/:userId", Description: "学员学习详情"},
		{Group: "课程", Method: "GET", Path: "/api/admin/course/all", Description: "全部课程"},
	}
	db.Create(&apis)
	log.Println("[Seed] 班级接口已创建")
}

// seedClassReportButton 班级「学习报告」按钮（幂等）
func seedClassReportButton(db *gorm.DB) {
	var count int64
	db.Model(&admin.Menu{}).Where("perms = ? AND guard_type = ?", "class:report", "admin").Count(&count)
	if count > 0 {
		return
	}
	var classListMenu admin.Menu
	if err := db.Where("name = ? AND guard_type = ?", "班级列表", "admin").First(&classListMenu).Error; err != nil {
		return
	}
	db.Create(&admin.Menu{
		ParentID:  classListMenu.ID,
		Name:      "学习报告",
		Type:      "F",
		Perms:     "class:report",
		GuardType: "admin",
	})
}

// seedExamApis 考试接口权限点（幂等）
func seedExamApis(db *gorm.DB) {
	apis := []admin.Api{
		{Group: "考试", Method: "GET", Path: "/api/admin/course/:id/question", Description: "试题列表"},
		{Group: "考试", Method: "GET", Path: "/api/admin/course/:id/question/all", Description: "全部试题"},
		{Group: "考试", Method: "POST", Path: "/api/admin/course/:id/question", Description: "创建试题"},
		{Group: "考试", Method: "POST", Path: "/api/admin/course/:id/question/import", Description: "批量导入试题"},
		{Group: "考试", Method: "PUT", Path: "/api/admin/question/:id", Description: "更新试题"},
		{Group: "考试", Method: "DELETE", Path: "/api/admin/question/:id", Description: "删除试题"},
		{Group: "考试", Method: "GET", Path: "/api/admin/course/:id/testpaper", Description: "试卷列表"},
		{Group: "考试", Method: "POST", Path: "/api/admin/course/:id/testpaper", Description: "创建试卷"},
		{Group: "考试", Method: "PUT", Path: "/api/admin/testpaper/:id", Description: "更新试卷"},
		{Group: "考试", Method: "DELETE", Path: "/api/admin/testpaper/:id", Description: "删除试卷"},
		{Group: "考试", Method: "GET", Path: "/api/admin/testpaper/:id/questions", Description: "试卷试题"},
		{Group: "考试", Method: "PUT", Path: "/api/admin/testpaper/:id/questions", Description: "设置试卷试题"},
		{Group: "考试", Method: "GET", Path: "/api/admin/course/:id/exam-report", Description: "考试报告"},
		{Group: "考试", Method: "GET", Path: "/api/admin/exam-record/:id", Description: "考试记录详情"},
	}
	for _, a := range apis {
		seedApiIfMissing(db, a)
	}

	// 课程列表增加考试相关按钮权限（幂等）
	buttons := []struct {
		perms string
		name  string
	}{
		{"course:question", "试题管理"},
		{"course:testpaper", "试卷管理"},
		{"course:exam-report", "考试报告"},
	}
	var courseListMenu admin.Menu
	if err := db.Where("perms = ? AND guard_type = ?", "course:list", "admin").First(&courseListMenu).Error; err != nil {
		return
	}
	for _, b := range buttons {
		var count int64
		db.Model(&admin.Menu{}).Where("perms = ? AND guard_type = ?", b.perms, "admin").Count(&count)
		if count > 0 {
			continue
		}
		db.Create(&admin.Menu{
			ParentID:  courseListMenu.ID,
			Name:      b.name,
			Type:      "F",
			Perms:     b.perms,
			GuardType: "admin",
		})
	}
	log.Println("[Seed] 考试接口与按钮已创建")
}

// seedSettingMenu 系统设置菜单（幂等：挂在「系统管理」下）
func seedSettingMenu(db *gorm.DB) {
	var count int64
	db.Model(&admin.Menu{}).Where("perms = ? AND guard_type = ?", "system:setting", "admin").Count(&count)
	if count > 0 {
		return
	}
	var sysMenu admin.Menu
	if err := db.Where("name = ? AND guard_type = ?", "系统管理", "admin").First(&sysMenu).Error; err != nil {
		return
	}
	db.Create(&admin.Menu{
		ParentID:  sysMenu.ID,
		Name:      "系统设置",
		Type:      "C",
		Path:      "setting",
		Component: "system/setting",
		Icon:      "Tools",
		Sort:      5,
		GuardType: "admin",
		Perms:     "system:setting",
	})
	log.Println("[Seed] 系统设置菜单已创建")
}

// seedSettingApi 系统设置接口权限点（幂等）
func seedSettingApi(db *gorm.DB) {
	seedApiIfMissing(db, admin.Api{Group: "系统", Method: "PUT", Path: "/api/admin/setting/site", Description: "修改系统设置"})
}
