package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ZyqAlwaysCool/anvil/internal/platform/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMySQL 创建 GORM MySQL 连接并执行 Ping。
func NewMySQL(cfg config.MySQLConfig) (*gorm.DB, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.DSN == "" {
		return nil, fmt.Errorf("mysql dsn is required")
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("mysql open failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("mysql get sql db: %w", err)
	}
	if err := pingSQL(sqlDB, cfg.Timeout); err != nil {
		return nil, fmt.Errorf("mysql ping failed: %w", err)
	}
	return db, nil
}

func CloseMySQL(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func pingSQL(db *sql.DB, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return db.PingContext(ctx)
}
