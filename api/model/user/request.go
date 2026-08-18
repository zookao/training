package user

// RegisterReq 学员注册
type RegisterReq struct {
	Username     string `json:"username" binding:"required,min=3,max=50"`
	Password     string `json:"password" binding:"required,min=6"`
	Nickname     string `json:"nickname"`
	StudentNo    string `json:"studentNo"`
	DepartmentID uint   `json:"departmentId"`
	Phone        string `json:"phone" binding:"required"`
}

// LoginReq 学员登录（账号或手机号）
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRes 登录响应
type LoginRes struct {
	Token   string `json:"token"`
	Expires int    `json:"expires"` // 小时
}

// UpdateProfileReq 更新资料（学员自助）
// 注意：department_id 由管理员维护，学员不可自助修改，故不在此结构体中暴露。
type UpdateProfileReq struct {
	Nickname  string `json:"nickname"`
	StudentNo string `json:"studentNo"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// ChangePwdReq 修改密码
type ChangePwdReq struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// UserInfoRes 当前学员信息
type UserInfoRes struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	StudentNo    string `json:"studentNo"`
	DepartmentID uint   `json:"departmentId"`
	Avatar       string `json:"avatar"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
}

// ProgressReq 上报学习进度
type ProgressReq struct {
	VideoID  uint `json:"videoId" binding:"required"`
	CourseID uint `json:"courseId" binding:"required"`
	ClassID  uint `json:"classId"`
	Position int  `json:"position"`  // 当前秒
	Duration int  `json:"duration"`  // 总秒
	Completed bool `json:"completed"` // 前端标记已播放结束
}
