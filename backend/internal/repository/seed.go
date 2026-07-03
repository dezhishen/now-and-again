package repository

import (
	"math/rand"
	"os"

	"github.com/dezhishen/now-and-again/backend/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ─── Unified startup initialization ──────────────────────────────
//
// RunAll is the single entry point for all seed/init logic.
// Every operation is idempotent — it will not overwrite or duplicate
// existing data, and it is safe to call on every startup.

func RunAll(db *gorm.DB) {
	logger.Infof("running startup initialization...")

	seedRoles(db)
	seedDefaultSettings(db)
	backfillSettingsScope(db)
	seedAdminUser(db)
	seedDefaultTaskTemplateSubscription(db)

	logger.Infof("startup initialization complete")
}

// ─── Roles ────────────────────────────────────────────────────────

func seedRoles(db *gorm.DB) {
	roles := []RoleModel{
		{Name: "admin", Description: "系统管理员"},
		{Name: "user", Description: "普通用户"},
	}
	for _, r := range roles {
		db.Where("name = ?", r.Name).FirstOrCreate(&r)
	}
}

// ─── Default Settings ─────────────────────────────────────────────
//
// Only creates settings that don't already exist. Existing values
// (including user modifications) are NEVER overwritten.

func seedDefaultSettings(db *gorm.DB) {
	defaults := map[string]string{
		"storage.type": "local",
	}

	for k, v := range defaults {
		var existing SystemSettingModel
		if err := db.Where("key = ?", k).First(&existing).Error; err != nil {
			db.Create(&SystemSettingModel{Key: k, Value: v, Scope: "admin"})
		}
	}

	// Default password: auto-generate only if not already set
	var pwdSetting SystemSettingModel
	if err := db.Where("key = ?", "default_password").First(&pwdSetting).Error; err != nil {
		pwd := randomString(12)
		db.Create(&SystemSettingModel{Key: "default_password", Value: pwd, Scope: "admin"})
	}
}

// backfillSettingsScope ensures all existing settings have a scope.
// Safe to run on every startup (only touches rows with empty scope).
func backfillSettingsScope(db *gorm.DB) {
	db.Model(&SystemSettingModel{}).Where("scope = ?", "").Update("scope", "admin")
}

// ─── Admin User ───────────────────────────────────────────────────
//
// Creates the default admin account only if no users exist yet.
// Prints credentials to the log on first creation.

func seedAdminUser(db *gorm.DB) {
	var count int64
	if err := db.Model(&UserModel{}).Count(&count).Error; err != nil {
		logger.Warnf("seed admin: count users failed: %v", err)
		return
	}
	if count > 0 {
		return
	}

	password := os.Getenv("NA_ADMIN_DEFAULT_PASSWORD")
	if password == "" {
		password = os.Getenv("ADMIN_DEFAULT_PASSWORD") // backward compat
	}
	if password == "" {
		password = randomString(12)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Warnf("seed admin: hash password failed: %v", err)
		return
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
		logger.Warnf("seed admin: %v", err)
		return
	}

	logger.Infof("========================================")
	logger.Infof("  Default admin account created")
	logger.Infof("  Username: admin")
	logger.Infof("  Password: %s", password)
	logger.Infof("========================================")
}

// ─── Default Task Template Subscription ───────────────────────────
//
// Creates a system-level HTTP subscription pointing to the official
// GitHub repository. Idempotent — skips if the subscription already exists.

func seedDefaultTaskTemplateSubscription(db *gorm.DB) {
	subs := []struct {
		url  string
		name string
	}{
		{
			url:  "https://raw.githubusercontent.com/dezhishen/now-and-again/main/templates/daily_inspection.yaml",
			name: "官方模板-巡检",
		},
		{
			url:  "https://raw.githubusercontent.com/dezhishen/now-and-again/main/templates/household.yaml",
			name: "官方模板-家庭事务",
		},
	}
	for _, s := range subs {
		sub := TaskTemplateSubscriptionModel{
			ProviderCode:         "http",
			URL:                  s.url,
			Name:                 s.name,
			AutoRefresh:          true,
			RefreshIntervalHours: 24,
			Enabled:              true,
		}
		result := db.Where("provider_code = ? AND url = ?", "http", sub.URL).
			FirstOrCreate(&sub)
		if result.RowsAffected > 0 {
			logger.Infof("seed: created default task template subscription: %s", sub.Name)
		}
	}
}

// ─── Helpers ──────────────────────────────────────────────────────

func randomString(length int) string {
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
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// randomPassword is an alias kept for backward compatibility.
func randomPassword(length int) string {
	return randomString(length)
}
