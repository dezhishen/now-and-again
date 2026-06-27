package repository

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/config"
	"github.com/dezhishen/now-and-again/backend/internal/logger"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
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
		Logger: gormlogger.Default.LogMode(gormlogger.Error),
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
	logger.Infof("running database migrations...")
	if err := db.AutoMigrate(
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
		&TaskTemplateModel{},
		&TodoModel{},
		&TaskLogModel{},
		&InspectionResultModel{},
		&CheckItemModel{},
		&CheckItemBranchModel{},
		&IcsFeedModel{},
	); err != nil {
		return err
	}

	// Backfill family_id for existing locations (new first-class entity migration)
	if err := backfillLocationFamilyID(db); err != nil {
		logger.Warnf("warning: backfill location family_id: %v", err)
	}
	return nil
}

func backfillLocationFamilyID(db *gorm.DB) error {
	// Update locations that have a floor_plan_id but no family_id yet
	return db.Exec(`
		UPDATE locations SET family_id = (
			SELECT fp.family_id FROM floor_plans fp WHERE fp.id = locations.floor_plan_id
		) WHERE family_id = '' OR family_id IS NULL
	`).Error
}

// Seed inserts default roles.
func Seed(db *gorm.DB) error {
	logger.Infof("seeding default data...")

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

	logger.Infof("seed complete")
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
