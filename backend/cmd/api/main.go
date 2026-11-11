package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"autodidact-backend/internal/config"
	"autodidact-backend/internal/db"
	"autodidact-backend/internal/handler"
	"autodidact-backend/internal/repository"
	"autodidact-backend/internal/service"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting Autodidact API server on %s:%s...", cfg.Host, cfg.Port)
	log.Printf("Connecting to SQLite database at: %s", cfg.DBPath)

	database, err := db.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer database.Close()

	repo := repository.NewSyllabusRepo(database)
	svc := service.NewSyllabusService(repo)
	focusRepo := repository.NewFocusRepo(database)
	focusSvc := service.NewFocusService(focusRepo, repo, database)
	router := handler.NewRouter(cfg, svc, focusSvc)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Autodidact API server listening on http://%s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down API server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown forced: %v", err)
	}

	log.Println("Server exited successfully.")
}
