package repository

import (
	"fmt"
	"log"
	"math/rand"
	"time"

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
		&LocationModel{},
	)
}

// Seed inserts default roles.
func Seed(db *gorm.DB) error {
	log.Println("seeding default data...")

	roles := []RoleModel{
		{Name: "admin", Description: "系统管理员"},
		{Name: "user", Description: "普通用户"},
	}
	for _, r := range roles {
		db.Where("name = ?", r.Name).FirstOrCreate(&r)
	}

	// Default system settings
	settingsDefaults := map[string]string{
		"storage.type": "local",
	}
	for k, v := range settingsDefaults {
		var existing SystemSettingModel
		if err := db.Where("key = ?", k).First(&existing).Error; err != nil {
			db.Create(&SystemSettingModel{Key: k, Value: v})
		}
	}

	log.Println("seed complete")
	return nil
}

// GenInviteCode generates an 8-char alphanumeric invite code.
func GenInviteCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}
