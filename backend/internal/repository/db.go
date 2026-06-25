package repository

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dezhishen/now-and-again/backend/internal/config"
)

func NewDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "sqlite":
		dialector = sqlite.Open(cfg.DSN)
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
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

// Migrate runs auto-migration for all models.
func Migrate(db *gorm.DB) error {
	log.Println("running database migrations...")
	return db.AutoMigrate(
		&UserModel{},
		&FamilyModel{},
		&FamilyMemberModel{},
		&SubGroupModel{},
		&SubGroupMemberModel{},
		&TaskTypeModel{},
		&TaskModel{},
		&TaskAssigneeModel{},
		&TaskDependencyModel{},
		&TaskChainModel{},
		&TaskChainStepModel{},
		&TaskLogModel{},
		&InspectionModel{},
		&InspectionItemModel{},
		&NotificationChannelModel{},
		&NotificationTemplateModel{},
		&UserChannelConfigModel{},
		&NotificationModel{},
		&RefreshTokenModel{},
		&ApiKeyModel{},
	)
}

// Seed inserts default data (notification channels). Task types are seeded
// from main.go using the scheduler registry to avoid import cycles.
func Seed(db *gorm.DB) error {
	log.Println("seeding default data...")

	// Default notification channels.
	defaultChannels := []NotificationChannelModel{
		{Code: "email", Name: "邮件通知", Description: "通过 SMTP 发送邮件", IsActive: true},
		{Code: "push", Name: "App 推送", Description: "FCM / APNs 推送", IsActive: false},
		{Code: "wechat_webhook", Name: "企业微信机器人", Description: "通过 Webhook 发送到企业微信群", IsActive: false},
		{Code: "webhook", Name: "自定义 Webhook", Description: "发送到自定义 HTTP 端点", IsActive: false},
	}
	for _, ch := range defaultChannels {
		db.Where("code = ?", ch.Code).FirstOrCreate(&ch)
	}

	return nil
}
