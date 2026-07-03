package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/dezhishen/now-and-again/backend/internal/config"
	"github.com/dezhishen/now-and-again/backend/internal/handler"
	"github.com/dezhishen/now-and-again/backend/internal/middleware"
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/internal/service"
	"github.com/dezhishen/now-and-again/backend/internal/webui"
	"github.com/dezhishen/now-and-again/backend/pkg/logger"
	"github.com/dezhishen/now-and-again/backend/pkg/scheduler"
	"github.com/dezhishen/now-and-again/backend/pkg/tasktemplate/builtin"
	"github.com/gin-gonic/gin"
)

func main() {
	// ── Logger (must be first — config.Load may fail and call Fatalf) ─────
	logDir := filepath.Join(os.Getenv("NA_DATA_DIR"), "logs")
	if logDir == "logs" || logDir == "/logs" {
		logDir = filepath.Join("data", "logs") // default when NA_DATA_DIR is empty
	}
	if _, err := logger.Init(logDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
	}
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("failed to load config: %v", err)
	}

	// ── Database ────────────────────────────────────────────────
	db, err := repository.NewDB(cfg.Database)
	if err != nil {
		logger.Fatalf("failed to connect database: %v", err)
	}
	if err := repository.Migrate(db); err != nil {
		logger.Fatalf("failed to migrate: %v", err)
	}
	repository.RunAll(db)

	// Load family defaults (env NA_FAMILY_DEFAULTS_PATH overrides YAML path, NA_FAMILY_DEFAULTS_INIT=false disables)
	defaultsPath := os.Getenv("NA_FAMILY_DEFAULTS_PATH")
	if defaultsPath == "" {
		defaultsPath = filepath.Join(cfg.DataDir, "family_defaults.yaml")
	}
	service.LoadFamilyDefaults(defaultsPath)

	// ── Repositories ────────────────────────────────────────────
	userRepo := repository.NewUserRepo(db)
	familyRepo := repository.NewFamilyRepo(db)
	apiKeyRepo := repository.NewApiKeyRepo(db)
	floorPlanRepo := repository.NewFloorPlanRepo(db)
	imageRepo := repository.NewImageRepo(db)
	settingsRepo := repository.NewSettingsRepo(db)
	taskRepo := repository.NewTaskRepo(db)
	icsRepo := repository.NewIcsRepo(db)
	taskTemplateRepo := repository.NewTaskTemplateRepo(db)

	// ── Services ────────────────────────────────────────────────
	userSvc := service.NewUserService(userRepo, settingsRepo, cfg.JWTSecret)
	familySvc := service.NewFamilyService(familyRepo, floorPlanRepo, userRepo)
	apiKeySvc := service.NewApiKeyService(apiKeyRepo)
	imageSvc := service.NewImageService(imageRepo, cfg.UploadDir, settingsRepo)
	floorPlanSvc := service.NewFloorPlanService(floorPlanRepo, familyRepo, userRepo, imageSvc, imageRepo)

	// Scheduler with DB log
	if err := scheduler.Init(); err != nil {
		logger.Fatalf("failed to init scheduler engine: %v", err)
	}
	scheduler.SetLogger(func(taskID, status, message string) {
		taskRepo.CreateLog(taskID, status, message)
	})
	taskSvc := service.NewTaskService(taskRepo)
	todoSvc := service.NewTodoService(taskRepo)
	logSvc := service.NewLogService(taskRepo)
	icsSvc := service.NewIcsService(icsRepo, taskRepo, apiKeyRepo, userRepo)
	calendarSvc := service.NewCalendarService(taskRepo)
	taskTemplateSvc := service.NewTaskTemplateService(taskTemplateRepo)

	// Set data dir for builtin provider (admins can place custom .yaml in ${DATA_DIR}/templates/)
	builtin.SetDataDir(cfg.DataDir)

	// Sync all providers at startup so the DB is populated.
	// Each provider logs its own errors; failures do not block startup.
	if err := taskTemplateSvc.SyncAll(context.Background()); err != nil {
		logger.Warnf("warning: initial task template sync failed: %v", err)
	}

	// ── Bundle contracts ────────────────────────────────────────
	allContracts := service.NewAllContracts(userSvc, familySvc, apiKeySvc, floorPlanSvc, taskSvc, todoSvc, logSvc, calendarSvc, taskTemplateSvc)

	// ── HTTP Router ─────────────────────────────────────────────
	router := gin.Default()
	router.Use(middleware.CORS())

	// Serve uploaded files
	router.Static("/uploads", cfg.UploadDir)

	imageHandler := handler.NewImageHandlers(imageRepo)
	settingsHandler := handler.NewSettingsHandlers(settingsRepo)
	taskHandler := &handler.TaskHandlers{Svc: taskSvc}
	todoHandler := &handler.TodoHandlers{Svc: todoSvc}
	logHandler := &handler.LogHandlers{Svc: logSvc}
	icsHandler := &handler.IcsHandlers{Svc: icsSvc}
	calendarHandler := &handler.CalendarHandlers{Svc: calendarSvc}
	locationHandler := &handler.LocationHandlers{C: floorPlanSvc}
	taskTemplateHandler := &handler.TaskTemplateHandlers{Svc: taskTemplateSvc}
	auth := router.Group("")
	auth.Use(middleware.JWTAuth(cfg.JWTSecret, apiKeyRepo))
	auth.Use(middleware.ScopeGuard())

	// Family-scoped routes: X-Family-Id header required (falls back to default)
	familyAuth := router.Group("")
	familyAuth.Use(middleware.JWTAuth(cfg.JWTSecret, apiKeyRepo))
	familyAuth.Use(middleware.ScopeGuard())
	familyAuth.Use(middleware.FamilyGuard(familySvc))

	// Admin-only routes (JWT user must have admin role; API key must have admin scope)
	adminAuth := auth.Group("")
	adminAuth.Use(middleware.AdminGuard(userSvc))

	// Family-owner-only routes (JWT user must be family owner; API key must have family:admin scope)
	ownerAuth := familyAuth.Group("")
	ownerAuth.Use(middleware.OwnerGuard(familySvc))

	handler.RegisterRoutes(router, auth, adminAuth, familyAuth, ownerAuth, allContracts, imageHandler, settingsHandler, taskHandler, todoHandler, logHandler, icsHandler, calendarHandler, locationHandler, taskTemplateHandler)

	// ── Task Template: admin-only / owner-only routes ────────────
	adminAuth.POST("/api/admin/task-templates/providers/:code/refresh", taskTemplateHandler.AdminRefreshProvider)

	// Admin subscription CRUD
	adminAuth.GET("/api/admin/task-template-subscriptions", taskTemplateHandler.AdminListSubscriptions)
	adminAuth.POST("/api/admin/task-template-subscriptions", taskTemplateHandler.AdminCreateSubscription)
	adminAuth.PUT("/api/admin/task-template-subscriptions/:id", taskTemplateHandler.AdminUpdateSubscription)
	adminAuth.DELETE("/api/admin/task-template-subscriptions/:id", taskTemplateHandler.AdminDeleteSubscription)

	// Family-owner template CRUD
	ownerAuth.POST("/api/task-templates", taskTemplateHandler.CreateFamily)
	ownerAuth.PUT("/api/task-templates/:code", taskTemplateHandler.UpdateFamily)
	ownerAuth.DELETE("/api/task-templates/:code", taskTemplateHandler.DeleteFamily)
	ownerAuth.POST("/api/task-templates/providers/:code/refresh", taskTemplateHandler.RefreshFamilyProvider)

	// Family-owner subscription CRUD
	ownerAuth.POST("/api/task-template-subscriptions", taskTemplateHandler.FamilyCreateSubscription)
	ownerAuth.PUT("/api/task-template-subscriptions/:id", taskTemplateHandler.FamilyUpdateSubscription)
	ownerAuth.DELETE("/api/task-template-subscriptions/:id", taskTemplateHandler.FamilyDeleteSubscription)

	// ── Frontend SPA ───────────────────────────────────────────
	webui.Serve(router)

	// ── Scheduler ──────────────────────────────────────────────
	scheduler.Start()
	defer scheduler.Stop()

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
