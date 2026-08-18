package admin

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRes 登录响应
type LoginRes struct {
	Token    string `json:"token"`
	Expires  int    `json:"expires"` // 小时
}

// AdminCreateReq 管理员创建
type AdminCreateReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Status   int8   `json:"status"`
	RoleIDs  []uint `json:"roleIds"`
}

// AdminUpdateReq 管理员更新
type AdminUpdateReq struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Status   int8   `json:"status"`
	RoleIDs  []uint `json:"roleIds"`
}

// AdminResetPwdReq 重置密码
type AdminResetPwdReq struct {
	Password string `json:"password" binding:"required,min=6"`
}

// RoleReq 角色创建/更新
type RoleReq struct {
	Name      string `json:"name" binding:"required"`
	Title     string `json:"title"`
	GuardType string `json:"guardType"`
	Sort      int    `json:"sort"`
	Status    int8   `json:"status"`
	Remark    string `json:"remark"`
}

// AssignIDs 分配资源请求（菜单/接口）
type AssignIDs struct {
	IDs []uint `json:"ids"`
}

// MenuReq 菜单创建/更新
type MenuReq struct {
	ParentID  uint   `json:"parentId"`
	Name      string `json:"name" binding:"required"`
	Type      string `json:"type" binding:"required"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Redirect  string `json:"redirect"`
	Icon      string `json:"icon"`
	Hidden    bool   `json:"hidden"`
	KeepAlive bool   `json:"keepAlive"`
	Sort      int    `json:"sort"`
	GuardType string `json:"guardType"`
	Perms     string `json:"perms"`
}

// ApiReq 接口创建/更新
type ApiReq struct {
	Group       string `json:"group"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

// UserInfoRes 当前用户信息
type UserInfoRes struct {
	ID       uint     `json:"id"`
	Username string   `json:"username"`
	Nickname string   `json:"nickname"`
	Avatar   string   `json:"avatar"`
	Roles    []string `json:"roles"`
	Perms    []string `json:"perms"`
}

// VideoReq 视频请求（有 ID=更新，无 ID=新增）
type VideoReq struct {
	ID                   uint   `json:"id"`
	URL                  string `json:"url" binding:"required"`
	Thumbnail            string `json:"thumbnail"`
	Courseware            string `json:"courseware"`        // 课件URL（PPTX）
	CoursewarePageCount   int    `json:"coursewarePageCount"` // 课件幻灯片页数
	CoursewarePages       string `json:"coursewarePages"`   // 课件每页时长 JSON: [10,60,300]（秒）
	CoursewarePDF         string `json:"coursewarePdf"`     // 课件 PDF URL（由 PPTX 转换生成，用于在线预览）
	Title                string `json:"title"`
	Description          string `json:"description"`
	Sort                 int    `json:"sort"`
	Duration             int    `json:"duration"` // 视频时长（秒，由上传接口解析返回）
}

// CourseReq 课程创建/更新
type CourseReq struct {
	Title       string     `json:"title" binding:"required"`
	Cover       string     `json:"cover"`
	Description string     `json:"description"`
	Sort        int        `json:"sort"`
	Status      int8       `json:"status"`
	Videos      []VideoReq `json:"videos"`
}

// UploadVideoRes 上传视频响应
type UploadVideoRes struct {
	URL                string `json:"url"`
	Thumbnail          string `json:"thumbnail"`
	Courseware          string `json:"courseware"`          // 课件URL（PPTX）
	CoursewarePageCount int    `json:"coursewarePageCount"` // 课件幻灯片页数
	CoursewarePDF       string `json:"coursewarePdf"`       // 课件 PDF URL（由 PPTX 转换生成）
	Filename           string `json:"filename"`
	Duration           int    `json:"duration"` // 视频时长（秒，服务端解析）
}

// UploadImageRes 上传图片响应（封面图等）
type UploadImageRes struct {
	URL string `json:"url"`
}

// UserCreateReq 学员创建
type UserCreateReq struct {
	Username     string `json:"username" binding:"required,min=2,max=50"`
	Password     string `json:"password" binding:"required,min=6"`
	Nickname     string `json:"nickname"`
	StudentNo    string `json:"studentNo"`
	DepartmentID uint   `json:"departmentId"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Status       int8   `json:"status"`
}

// UserUpdateReq 学员更新
type UserUpdateReq struct {
	Nickname     string `json:"nickname"`
	StudentNo    string `json:"studentNo"`
	DepartmentID uint   `json:"departmentId"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Status       int8   `json:"status"`
}

// DepartmentReq 院系创建/更新
type DepartmentReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
}

// UserResetPwdReq 重置学员密码
type UserResetPwdReq struct {
	Password string `json:"password" binding:"required,min=6"`
}

// ClassReq 班级创建/更新
type ClassReq struct {
	Name        string `json:"name" binding:"required"`
	Cover       string `json:"cover"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
}

// QuestionReq 试题创建/更新
type QuestionReq struct {
	Type     int8   `json:"type" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Options  string `json:"options"`  // JSON: [{"label":"A","content":"..."}]
	Answer   string `json:"answer"`   // JSON: ["A"] 或 ["A","C"]
	Analysis string `json:"analysis"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status"`
}

// TestpaperReq 试卷创建/更新
type TestpaperReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        int8   `json:"type"`       // 1随时考 2课程完成后考
	TotalScore  int    `json:"totalScore"` // 固定100
	PassScore   int    `json:"passScore"`
	Duration    int    `json:"duration"` // 考试时长（分钟）
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
}

// TestpaperQuestionItem 试卷试题项（含分值）
type TestpaperQuestionItem struct {
	QuestionID uint `json:"questionId"`
	Score      int  `json:"score"`
}

// TestpaperQuestionsReq 设置试卷试题
type TestpaperQuestionsReq struct {
	Items []TestpaperQuestionItem `json:"items"`
}
