package user

// ClassProgressItem 我的班级（含完成度）
type ClassProgressItem struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Cover       string `json:"cover"`
	Description string `json:"description"`
	CourseCount int    `json:"courseCount"`
	Percent     int    `json:"percent"`
}

// CourseProgressItem 课程进度项（班级详情内）
type CourseProgressItem struct {
	ID              uint   `json:"id"`
	Title           string `json:"title"`
	Cover           string `json:"cover"`
	Description     string `json:"description"`
	VideoCount      int    `json:"videoCount"`
	CompletedVideos int    `json:"completedVideos"`
	Percent         int    `json:"percent"`
}

// VideoLearnItem 学习页视频项（含该学员进度）
type VideoLearnItem struct {
	ID               uint   `json:"id"`
	URL              string `json:"url"`
	Thumbnail        string `json:"thumbnail"`
	Courseware        string `json:"courseware"`        // 课件URL（PPTX）
	CoursewarePages   string `json:"coursewarePages"`   // 课件每页时长 JSON: [10,60,300]（秒）
	CoursewarePDF     string `json:"coursewarePdf"`     // 课件 PDF URL（用于在线预览）
	Title            string `json:"title"`
	Description      string `json:"description"`
	Sort             int    `json:"sort"`
	Duration         int    `json:"duration"`
	Position         int    `json:"position"`
	MaxPosition      int    `json:"maxPosition"`
	Percent          int    `json:"percent"`
	Completed        bool   `json:"completed"`
	NextCheckPosition int   `json:"nextCheckPosition"` // 下次校验触发的视频位置（秒），0=不校验
	CheckPending      bool  `json:"checkPending"`      // 是否有待完成的滑动校验
}

// CourseLearnRes 课程学习详情
type CourseLearnRes struct {
	ID          uint             `json:"id"`
	Title       string           `json:"title"`
	Cover       string           `json:"cover"`
	Description string           `json:"description"`
	Videos      []VideoLearnItem `json:"videos"`
}

// ClassDetailRes 班级详情（含课程进度）
type ClassDetailRes struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	Cover       string               `json:"cover"`
	Description string               `json:"description"`
	Courses     []CourseProgressItem `json:"courses"`
	Percent     int                  `json:"percent"`
}

// ProgressRes 上报进度响应
type ProgressRes struct {
	Percent           int  `json:"percent"`
	Completed         bool `json:"completed"`
	MaxPosition       int  `json:"maxPosition"`
	CheckPending      bool `json:"checkPending"`
	NextCheckPosition int  `json:"nextCheckPosition"`
}

// ClassLearningReportRes 班级学习报告（admin 视角）
type ClassLearningReportRes struct {
	ClassID   uint                `json:"classId"`
	ClassName string              `json:"className"`
	Students  []StudentReportItem `json:"students"`
}

// StudentReportItem 班级学习报告-学员项
type StudentReportItem struct {
	UserID    uint   `json:"userId"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	StudentNo string `json:"studentNo"`
	Percent   int    `json:"percent"` // 班级完成%
}

// StudentLearningDetailRes 学员详细学习数据（admin 视角）
type StudentLearningDetailRes struct {
	UserID   uint               `json:"userId"`
	Username string             `json:"username"`
	Nickname string             `json:"nickname"`
	Courses  []CourseReportItem `json:"courses"`
}

// CourseReportItem 学员学习详情-课程项
type CourseReportItem struct {
	CourseID        uint              `json:"courseId"`
	Title           string            `json:"title"`
	Percent         int               `json:"percent"`
	CompletedVideos int               `json:"completedVideos"`
	VideoCount      int               `json:"videoCount"`
	Videos          []VideoReportItem `json:"videos"`
}

// VideoReportItem 学员学习详情-视频项
type VideoReportItem struct {
	VideoID     uint   `json:"videoId"`
	Title       string `json:"title"`
	Percent     int    `json:"percent"`
	Completed   bool   `json:"completed"`
	Duration    int    `json:"duration"`
	MaxPosition int    `json:"maxPosition"`
}
