package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/config"
	"github.com/dezhishen/now-and-again/backend/pkg/logger"
	"github.com/dezhishen/now-and-again/backend/pkg/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// gormZapLogger implements gormlogger.Interface, routing through zap.
type gormZapLogger struct{}

func (l gormZapLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface { return l }
func (l gormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	logger.Debugf("gorm: "+msg, data...)
}
func (l gormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	logger.Warnf("gorm: "+msg, data...)
}
func (l gormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	// Convert data to key=val pairs for readability, truncating long values.
	parts := make([]string, 0, len(data))
	for i := 0; i+1 < len(data); i += 2 {
		key := fmt.Sprint(data[i])
		val := fmt.Sprint(data[i+1])
		if len(val) > 80 {
			val = val[:80] + "..."
		}
		parts = append(parts, key+"="+val)
	}
	extra := ""
	if len(parts) > 0 {
		extra = " " + fmt.Sprint(parts)
	}
	logger.Errorf("gorm: %s%s", msg, extra)
}
func (l gormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("gorm error: %v [%.2fms, %d rows]", err, float64(elapsed.Microseconds())/1000.0, rows)
		return
	}
	// Log slow queries at Warn level (>200ms)
	if elapsed > 200*time.Millisecond {
		logger.Warnf("gorm slow query [%.0fms, %d rows]: %s", float64(elapsed.Microseconds())/1000.0, rows, sql)
	}
}

func NewDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dialector := sqlite.Open(cfg.DSN)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormZapLogger{},
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	return db, nil
}

// Migrate runs auto-migration for all models (built-in + plugin-registered).
func Migrate(db *gorm.DB) error {
	logger.Infof("running database migrations...")
	models := []any{
		&UserModel{},
		&AccountModel{},
		&RoleModel{},
		&UserRoleModel{},
		&FamilyModel{},
		&FamilyMemberModel{},
		&FamilyGroupModel{},
		&FamilyGroupMemberModel{},
		&RefreshTokenModel{},
		&ApiKeyModel{},
		&FloorPlanModel{},
		&LocationModel{},
		&ImageModel{},
		&SystemSettingModel{},
		&TaskModel{},
		&TodoModel{},
		&TaskLogModel{},
		&IcsFeedModel{},
		&TaskTemplateModel{},
		&TaskTemplateSubscriptionModel{},
	}
	models = append(models, model.MigrationModels()...)
	return db.AutoMigrate(models...)
}
