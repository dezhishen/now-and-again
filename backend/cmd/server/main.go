package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dezhishen/now-and-again/backend/internal/config"
	"github.com/dezhishen/now-and-again/backend/internal/handler"
	"github.com/dezhishen/now-and-again/backend/internal/middleware"
	"github.com/dezhishen/now-and-again/backend/internal/notifier"
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/internal/scheduler"
	_ "github.com/dezhishen/now-and-again/backend/internal/scheduler/handlers"
	"github.com/dezhishen/now-and-again/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ── Database ────────────────────────────────────────────────
	db, err := repository.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	if err := repository.Migrate(db); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}
	if err := repository.Seed(db); err != nil {
		log.Printf("warning: seed failed: %v", err)
	}

	// Seed task scheduling types from the scheduler registry.
	for _, h := range scheduler.All() {
		db.Where("code = ?", h.Code()).FirstOrCreate(&repository.ScheduleTypeModel{
			Code: h.Code(), Name: h.Name(), Category: h.Category(),
			DefaultPriority: h.DefaultPriority(), Icon: h.Icon(),
			IsActive: true,
		})
	}

	// ── Repositories ────────────────────────────────────────────
	userRepo := repository.NewUserRepo(db)
	familyRepo := repository.NewFamilyRepo(db)
	subGroupRepo := repository.NewSubGroupRepo(db)
	taskRepo := repository.NewTaskRepo(db)
	chainRepo := repository.NewChainRepo(db)
	inspectionRepo := repository.NewInspectionRepo(db)
	logRepo := repository.NewLogRepo(db)
	notifRepo := repository.NewNotificationRepo(db)
	apiKeyRepo := repository.NewApiKeyRepo(db)

	// ── Notifier Engine ─────────────────────────────────────────
	notifEngine := notifier.NewNotificationEngine(notifRepo, userRepo)

	// ── Services ────────────────────────────────────────────────
	userSvc := service.NewUserService(userRepo, cfg.JWTSecret)
	familySvc := service.NewFamilyService(familyRepo, userRepo)
	subGroupSvc := service.NewSubGroupService(subGroupRepo, familyRepo)
	taskSvc := service.NewTaskService(taskRepo, logRepo, notifEngine)
	chainSvc := service.NewChainService(chainRepo, taskRepo, logRepo, notifEngine)
	inspectionSvc := service.NewInspectionService(inspectionRepo, taskRepo, logRepo, notifEngine)
	logSvc := service.NewLogService(logRepo)
	notifSvc := service.NewNotificationService(notifRepo, userRepo)
	apiKeySvc := service.NewApiKeyService(apiKeyRepo)

	_ = subGroupSvc
	_ = logSvc
	_ = notifSvc

	// ── Bundle contracts ───────────────────────────────────────
	allContracts := service.NewAllContracts(
		userSvc, familySvc, subGroupSvc, taskSvc,
		chainSvc, inspectionSvc, logSvc, notifSvc, apiKeySvc,
	)

	// ── Scheduler ───────────────────────────────────────────────
	sched := scheduler.NewEngine(taskRepo, notifEngine)
	go sched.Start()

	// ── HTTP Router ─────────────────────────────────────────────
	router := gin.Default()
	router.Use(middleware.CORS())

	auth := router.Group("")
	auth.Use(middleware.JWTAuth(cfg.JWTSecret, apiKeyRepo))
	handler.RegisterRoutes(router, auth, allContracts)

	// ── Graceful Shutdown ───────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Printf("server listening on %s", addr)
		if err := router.Run(addr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")
	sched.Stop()
}
