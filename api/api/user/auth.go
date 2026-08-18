package user

import (
	"training/global"
	"training/middleware"
	"training/model/common"
	userModel "training/model/user"
	userService "training/service/user"
	"training/utils"

	"github.com/gin-gonic/gin"
)

// Register 学员注册
func Register(c *gin.Context) {
	var req userModel.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := userService.Register(req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "注册成功")
}

// Login 学员登录
func Login(c *gin.Context) {
	var req userModel.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	token, err := userService.Login(req, c.ClientIP())
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, userModel.LoginRes{Token: token, Expires: global.Config.JWT.ExpiresTime})
}

// UserInfo 当前学员信息
func UserInfo(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	res, err := userService.UserInfo(uid)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OK(c, res)
}

// UpdateProfile 更新资料
func UpdateProfile(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	var req userModel.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := userService.UpdateProfile(uid, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "更新成功")
}

// ChangePassword 修改密码
func ChangePassword(c *gin.Context) {
	uid := middleware.GetClaims(c).UserID
	var req userModel.ChangePwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := userService.ChangePassword(uid, req); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OKMsg(c, "修改成功")
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
