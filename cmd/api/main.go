package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"football-app/internal/config"
	"football-app/internal/handler"
	"football-app/internal/middleware"
	"football-app/internal/repository"
	"football-app/internal/service"
	"football-app/pkg/response"
)

func main() {
	cfg := config.Load()

	db, err := config.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	playerRepo := repository.NewPlayerRepository(db)
	matchRepo := repository.NewMatchRepository(db)
	reportRepo := repository.NewReportRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.TTL)
	teamService := service.NewTeamService(teamRepo)
	playerService := service.NewPlayerService(playerRepo, teamRepo)
	matchService := service.NewMatchService(matchRepo, teamRepo, playerRepo)
	reportService := service.NewReportService(reportRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	teamHandler := handler.NewTeamHandler(teamService, playerService)
	playerHandler := handler.NewPlayerHandler(playerService)
	matchHandler := handler.NewMatchHandler(matchService)
	reportHandler := handler.NewReportHandler(reportService)

	router := gin.Default()
	router.Use(middleware.ErrorHandler())

	router.GET("/health", func(c *gin.Context) {
		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			response.Error(c, http.StatusServiceUnavailable, "DB_UNREACHABLE", "database unreachable", nil)
			return
		}
		response.OK(c, http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	v1.POST("/auth/login", authHandler.Login)

	protected := v1.Group("")
	protected.Use(middleware.RequireAuth(cfg.JWT.Secret))

	protected.GET("/teams", teamHandler.List)
	protected.POST("/teams", teamHandler.Create)
	protected.GET("/teams/:id", teamHandler.Get)
	protected.PUT("/teams/:id", teamHandler.Update)
	protected.DELETE("/teams/:id", teamHandler.Delete)
	protected.GET("/teams/:id/players", teamHandler.ListPlayers)

	protected.GET("/players", playerHandler.List)
	protected.POST("/players", playerHandler.Create)
	protected.GET("/players/:id", playerHandler.Get)
	protected.PUT("/players/:id", playerHandler.Update)
	protected.DELETE("/players/:id", playerHandler.Delete)

	protected.GET("/matches", matchHandler.List)
	protected.POST("/matches", matchHandler.Create)
	protected.GET("/matches/:id", matchHandler.Get)
	protected.PUT("/matches/:id", matchHandler.Update)
	protected.DELETE("/matches/:id", matchHandler.Delete)
	protected.POST("/matches/:id/result", matchHandler.ReportResult)

	protected.GET("/reports/matches", reportHandler.ListMatchReports)
	protected.GET("/reports/matches/:id", reportHandler.MatchReport)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Println("server exited")
}
