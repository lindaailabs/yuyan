// Package server 是 apps/server 模块根包：承载 migration 嵌入与执行。
// migrations/ 目录按 LLM_DEV_GUIDE.md §3.3 约定使用时间戳前缀 SQL 文件，历史禁止修改删除。
package server

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrateUp 在给定连接上执行全部未应用的 migration（幂等：无变更时返回 nil）。
func MigrateUp(sqlDB *sql.DB) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	drv, err := migratemysql.WithInstance(sqlDB, &migratemysql.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", src, "mysql", drv)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
