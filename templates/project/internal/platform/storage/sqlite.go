//go:build ignore

package storage

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"anvil-scaffold-template/internal/platform/config"
)

// NewSQLite 创建 GORM SQLite 连接并执行 Ping。
func NewSQLite(cfg config.SQLiteConfig) (*gorm.DB, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.DSN == "" {
		return nil, fmt.Errorf("sqlite dsn is required")
	}

	db, err := gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("sqlite open failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("sqlite get sql db: %w", err)
	}
	if err := pingSQL(sqlDB, cfg.Timeout); err != nil {
		return nil, fmt.Errorf("sqlite ping failed: %w", err)
	}
	return db, nil
}

func CloseSQLite(db *gorm.DB) error {
	return CloseMySQL(db)
}
