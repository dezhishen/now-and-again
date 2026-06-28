package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/dezhishen/now-and-again/backend/internal/config"
	"github.com/dezhishen/now-and-again/backend/internal/handler"
	"github.com/dezhishen/now-and-again/backend/internal/logger"
	"github.com/dezhishen/now-and-again/backend/internal/middleware"
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/internal/service"
	"github.com/dezhishen/now-and-again/backend/internal/webui"
	"github.com/dezhishen/now-and-again/backend/pkg/scheduler"
	"github.com/gin-gonic/gin"
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
	todoSvc := service.NewTodoService(taskRepo, sched)
	logSvc := service.NewLogService(taskRepo)
	icsSvc := service.NewIcsService(icsRepo, taskRepo, apiKeyRepo, userRepo)

	// ── Seed admin ──────────────────────────────────────────────
	if _, err := repository.SeedAdmin(db); err != nil {
		logger.Warnf("warning: seed admin failed: %v", err)
	}

	// ── Bundle contracts ────────────────────────────────────────
	allContracts := service.NewAllContracts(userSvc, familySvc, apiKeySvc, floorPlanSvc, taskSvc, todoSvc, logSvc)

	// ── HTTP Router ─────────────────────────────────────────────
	router := gin.Default()
	router.Use(middleware.CORS())

	// Serve uploaded files
	router.Static("/uploads", cfg.UploadDir)

	imageHandler := handler.NewImageHandlers(imageRepo)
	settingsHandler := handler.NewSettingsHandlers(settingsRepo)
	taskHandler := &handler.TaskHandlers{TaskSvc: taskSvc, TodoSvc: todoSvc, LogSvc: logSvc}
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
