package admin

import (
	"errors"

	"training/global"
	"training/model/admin"
)

const (
	settingKeyLogo  = "site_logo"
	settingKeyName  = "site_name"
	defaultSiteName = "培训学习平台"
)

// GetSiteConfig 读取站点配置（logo 为空时返回空串，由前端回退到默认图）
func GetSiteConfig() (logoUrl, siteName string) {
	var items []admin.Setting
	global.DB.Where("`key` IN ?", []string{settingKeyLogo, settingKeyName}).Find(&items)
	m := make(map[string]string, len(items))
	for _, it := range items {
		m[it.Key] = it.Value
	}
	logoUrl = m[settingKeyLogo]
	siteName = m[settingKeyName]
	if siteName == "" {
		siteName = defaultSiteName
	}
	return
}

// UpdateSiteConfig 更新站点配置（upsert）
func UpdateSiteConfig(logoUrl, siteName string) error {
	if siteName == "" {
		return errors.New("站点名称不能为空")
	}
	pairs := map[string]string{
		settingKeyLogo: logoUrl,
		settingKeyName: siteName,
	}
	for k, v := range pairs {
		var s admin.Setting
		err := global.DB.Where("`key` = ?", k).First(&s).Error
		if err != nil {
			// 不存在则新建
			global.DB.Create(&admin.Setting{Key: k, Value: v})
			continue
		}
		global.DB.Model(&s).Update("value", v)
	}
	return nil
}
