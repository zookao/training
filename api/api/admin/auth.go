package admin

import (
	"training/global"
	"training/middleware"
	adminModel "training/model/admin"
	"training/model/common"
	adminService "training/service/admin"
	"training/utils"

	"github.com/gin-gonic/gin"
)

// Login 管理员登录
func Login(c *gin.Context) {
	var req adminModel.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	token, err := adminService.Login(req, c.ClientIP())
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, adminModel.LoginRes{Token: token, Expires: global.Config.JWT.ExpiresTime})
}

// UserInfo 当前管理员信息
func UserInfo(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	res, err := adminService.UserInfo(uid)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// Menus 当前管理员菜单树
func Menus(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	menus, err := adminService.CurrentMenus(uid)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, menus)
}

// Logout 退出登录（将当前 token 加入黑名单）
func Logout(c *gin.Context) {
	token := middleware.GetToken(c)
	claims := middleware.GetClaims(c)
	if token != "" && claims != nil && claims.ExpiresAt != nil {
		utils.BlacklistToken(token, claims.ExpiresAt.Time)
	}
	common.OKMsg(c, "退出成功")
}
