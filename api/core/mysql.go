package core

import (
	"fmt"
	"log"
	"time"

	"training/global"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Mysql 建立 MySQL 连接（首次启动自动创建数据库）
func Mysql() *gorm.DB {
	c := global.Config.MySQL

	// 1. 先连到 MySQL 服务器（不指定数据库），自动创建数据库
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Charset)
	rootDB, err := gorm.Open(mysql.Open(rootDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("[MySQL] 连接服务器失败（请检查 host/port/账号密码）: %v", err)
	}
	// 自动创建数据库（已存在则跳过）
	createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET %s", c.Database, c.Charset)
	if err := rootDB.Exec(createSQL).Error; err != nil {
		log.Fatalf("[MySQL] 创建数据库失败（请检查账号是否有 CREATE 权限）: %v", err)
	}
	sqlRoot, _ := rootDB.DB()
	_ = sqlRoot.Close()

	// 2. 连接到目标数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("[MySQL] 连接失败: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)  // 连接最长存活 5 分钟，避免 MySQL wait_timeout 关闭后用到死连接
	sqlDB.SetConnMaxIdleTime(60 * time.Second) // 空闲连接 60 秒回收，须小于 MySQL wait_timeout(100s)
	log.Println("[MySQL] 连接成功:", c.Database)
	return db
}
