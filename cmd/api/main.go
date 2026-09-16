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
	"football-app/internal/middleware"
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

	router := gin.Default()
	router.Use(middleware.ErrorHandler())

	router.GET("/health", func(c *gin.Context) {
		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			response.Error(c, http.StatusServiceUnavailable, "DB_UNREACHABLE", "database unreachable", nil)
			return
		}
		response.OK(c, http.StatusOK, gin.H{"status": "ok"})
	})

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
