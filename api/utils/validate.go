package utils

import "regexp"

var phoneRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

// IsValidPhone 校验中国大陆手机号（11位，1开头，第二位3-9）
func IsValidPhone(phone string) bool {
	return phoneRegexp.MatchString(phone)
}
