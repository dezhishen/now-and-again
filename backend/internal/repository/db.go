package repository

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/dezhishen/now-and-again/backend/internal/config"
	"github.com/dezhishen/now-and-again/backend/internal/logger"
	"github.com/dezhishen/now-and-again/backend/pkg/timeutil"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
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
		&TaskModel{},
		&TodoModel{},
		&TaskLogModel{},
		&InspectionResultModel{},
		&CheckItemModel{},
		&CheckItemBranchModel{},
		&IcsFeedModel{},
	)
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

// SeedAdmin creates a default admin user if none exists.
func SeedAdmin(db *gorm.DB) (password string, err error) {
	var count int64
	if err := db.Model(&UserModel{}).Count(&count).Error; err != nil || count > 0 {
		return "", nil
	}

	password = os.Getenv("ADMIN_DEFAULT_PASSWORD")
	if password == "" {
		password = randomPassword(12)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash admin password: %w", err)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		user := &UserModel{DisplayName: "管理员", Email: "admin@now-and-again.local"}
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		acc := &AccountModel{
			UserID: user.ID, Provider: "local",
			Username: "admin", PasswordHash: string(hash),
		}
		if err := tx.Create(acc).Error; err != nil {
			return err
		}
		var adminRole RoleModel
		if err := tx.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
			return err
		}
		return tx.Create(&UserRoleModel{UserID: user.ID, RoleID: adminRole.ID}).Error
	})
	if err != nil {
		return "", err
	}

	logger.Infof("========================================")
	logger.Infof("  Default admin account created")
	logger.Infof("  Username: admin")
	logger.Infof("  Password: %s", password)
	logger.Infof("========================================")
	return password, nil
}

func randomPassword(length int) string {
	const chars = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// GenInviteCode generates an 8-char alphanumeric invite code.
func GenInviteCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	rng := rand.New(rand.NewSource(timeutil.Now().UnixNano()))
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}
