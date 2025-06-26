package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"swiftwallet/internal/config"
	"swiftwallet/internal/db"
	"swiftwallet/internal/repository"
	"swiftwallet/internal/router"
	"swiftwallet/internal/service"
)

func main() {
	_ = godotenv.Load("config.env")

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := db.NewPool(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	r := router.New(service.New(repository.New(pool)))

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("swiftwallet listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown initiated")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown error: %v", err)
	}
}
