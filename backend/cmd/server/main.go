package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/config"
	"github.com/dezhishen/now-and-again/backend/internal/handler"
	"github.com/dezhishen/now-and-again/backend/internal/logger"
	"github.com/dezhishen/now-and-again/backend/internal/middleware"
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/internal/scheduler"
	"github.com/dezhishen/now-and-again/backend/internal/service"
	"github.com/dezhishen/now-and-again/backend/internal/webui"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("failed to load config: %v", err)
	}

	// ── Logger ─────────────────────────────────────────────────
	if _, err := logger.Init(filepath.Join(cfg.BaseDir(), "logs")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
	}
	defer logger.Sync()

	// ── Database ────────────────────────────────────────────────
	db, err := repository.NewDB(cfg.Database)
	if err != nil {
		logger.Fatalf("failed to connect database: %v", err)
	}
	if err := repository.Migrate(db); err != nil {
		logger.Fatalf("failed to migrate: %v", err)
	}
	if err := repository.Seed(db); err != nil {
		logger.Warnf("warning: seed failed: %v", err)
	}

	// ── Repositories ────────────────────────────────────────────
	userRepo := repository.NewUserRepo(db)
	familyRepo := repository.NewFamilyRepo(db)
	apiKeyRepo := repository.NewApiKeyRepo(db)
	floorPlanRepo := repository.NewFloorPlanRepo(db)
	imageRepo := repository.NewImageRepo(db)
	settingsRepo := repository.NewSettingsRepo(db)
	taskRepo := repository.NewTaskRepo(db)
	icsRepo := repository.NewIcsRepo(db)

	// ── Services ────────────────────────────────────────────────
	userSvc := service.NewUserService(userRepo, cfg.JWTSecret)
	familySvc := service.NewFamilyService(familyRepo, userRepo)
	apiKeySvc := service.NewApiKeyService(apiKeyRepo)
	imageSvc := service.NewImageService(imageRepo, cfg.UploadDir, settingsRepo)
	floorPlanSvc := service.NewFloorPlanService(floorPlanRepo, userRepo, imageSvc, imageRepo)

	// Scheduler with DB log
	sched, err := scheduler.New(func(taskID, status, message string) {
		taskRepo.CreateLog(taskID, status, message)
	})
	if err != nil {
		logger.Fatalf("failed to create scheduler: %v", err)
	}
	taskSvc := service.NewTaskService(taskRepo, sched)
	icsSvc := service.NewIcsService(icsRepo, taskRepo, apiKeyRepo, userRepo)

	// ── Seed admin ──────────────────────────────────────────────
	seedAdmin(db)

	// ── Bundle contracts ────────────────────────────────────────
	allContracts := service.NewAllContracts(userSvc, familySvc, apiKeySvc, floorPlanSvc)

	// ── HTTP Router ─────────────────────────────────────────────
	router := gin.Default()
	router.Use(middleware.CORS())

	// Serve uploaded files
	router.Static("/uploads", cfg.UploadDir)

	imageHandler := handler.NewImageHandlers(imageRepo)
	settingsHandler := handler.NewSettingsHandlers(settingsRepo)
	taskHandler := &handler.TaskHandlers{Svc: taskSvc}
	icsHandler := &handler.IcsHandlers{Svc: icsSvc}
	auth := router.Group("")
	auth.Use(middleware.JWTAuth(cfg.JWTSecret, apiKeyRepo))
	auth.Use(middleware.ScopeGuard())
	handler.RegisterRoutes(router, auth, allContracts, imageHandler, settingsHandler, taskHandler, icsHandler)

	// ── Frontend SPA ───────────────────────────────────────────
	webui.Serve(router)

	// ── Scheduler ──────────────────────────────────────────────
	sched.Start()
	defer sched.Stop()

	// ── Graceful Shutdown ───────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		logger.Infof("server listening on %s", addr)
		if err := router.Run(addr); err != nil {
			logger.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	logger.Infof("shutting down...")
}

// seedAdmin creates a default admin user if none exists.
func seedAdmin(db *gorm.DB) {
	var count int64
	if err := db.Model(&repository.UserModel{}).Count(&count).Error; err != nil || count > 0 {
		return
	}

	password := os.Getenv("ADMIN_DEFAULT_PASSWORD")
	if password == "" {
		password = randomPassword(12)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("failed to hash admin password: %v", err)
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		user := &repository.UserModel{
			DisplayName: "管理员",
			Email:       "admin@now-and-again.local",
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		acc := &repository.AccountModel{
			UserID:       user.ID,
			Provider:     "local",
			Username:     "admin",
			PasswordHash: string(hash),
		}
		if err := tx.Create(acc).Error; err != nil {
			return err
		}

		var adminRole repository.RoleModel
		if err := tx.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
			return err
		}
		ur := &repository.UserRoleModel{UserID: user.ID, RoleID: adminRole.ID}
		return tx.Create(ur).Error
	})
	if err != nil {
		logger.Errorf("failed to seed admin: %v", err)
		return
	}

	logger.Infof("========================================")
	logger.Infof("  Default admin account created")
	logger.Infof("  Username: admin")
	logger.Infof("  Password: %s", password)
	logger.Infof("  Change it after first login!")
	logger.Infof("========================================")
}

func randomPassword(length int) string {
	const chars = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}
