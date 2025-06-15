package storage

import (
	"log"
	"time"

	"video-converter/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewMySQLDB 创建MySQL数据库连接
func NewMySQLDB(dsn string) (*gorm.DB, error) {
	// 配置GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return nil, err
	}

	// 获取底层sql.DB以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生存时间

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	// 自动迁移数据库表
	if err := autoMigrate(db); err != nil {
		return nil, err
	}

	log.Println("MySQL数据库连接成功")
	return db, nil
}

// 其余代码保持不变...

// autoMigrate 自动迁移数据库表结构
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.ConversionTask{},
		&model.User{},
	)
}

// CloseDB 关闭数据库连接
func CloseDB(db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("获取数据库连接失败: %v", err)
			return
		}

		if err := sqlDB.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		} else {
			log.Println("数据库连接已关闭")
		}
	}
}
