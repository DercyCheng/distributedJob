package database

import (
	"fmt"
	"go-job/internal/models"
	"go-job/pkg/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.Config) error {
	dsn := cfg.GetMySQLDSN()

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.Database.MySQL.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MySQL.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.MySQL.ConnMaxLifetime * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 自动迁移
	if err := migrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	return nil
}

// migrate 自动迁移数据库表
func migrate() error {
	return db.AutoMigrate(
		&models.Job{},
		&models.JobExecution{},
		&models.Worker{},
		&models.JobSchedule{},
	)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}

// Close 关闭数据库连接
func Close() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Transaction 执行事务
func Transaction(fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}

// IsConnected 检查数据库连接状态
func IsConnected() bool {
	if db == nil {
		return false
	}

	sqlDB, err := db.DB()
	if err != nil {
		return false
	}

	return sqlDB.Ping() == nil
}
