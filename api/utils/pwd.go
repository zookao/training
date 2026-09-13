package utils

import "golang.org/x/crypto/bcrypt"

// dummyHash 固定的无效密码哈希：登录时账号不存在也执行一次 bcrypt 比对，
// 抹平"账号不存在"与"密码错误"之间的响应时间差，防止通过计时枚举用户名
const dummyHash = "$2a$10$vAHwpIpPS5pjeJSx3XiBm.SKS5Mii6qbjXl/vD1ufp6GURZalbjde"

// HashPassword 加密密码
func HashPassword(pwd string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword 校验密码
func CheckPassword(hash, pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)) == nil
}

// CheckDummyPassword 对哑哈希做一次无意义比对（结果恒 false），仅用于对齐耗时
func CheckDummyPassword(pwd string) {
	_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(pwd))
}
