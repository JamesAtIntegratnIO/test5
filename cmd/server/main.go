package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "test5/internal/config"
    "test5/internal/database"
    "test5/internal/cache"
    "test5/internal/server"
    "test5/internal/logger"
)

func main() {
    // Initialize logger
    logger := logger.New()
    logger.Info("Starting test5 server")

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        logger.Fatalf("Failed to load configuration: %v", err)
    }

    // Initialize database
    db, err := database.Initialize(cfg.Database)
    if err != nil {
        logger.Fatalf("Failed to initialize database: %v", err)
    }
    defer database.Close()

    
    // Initialize Redis
    cache, err := cache.Initialize(cfg.Redis)
    if err != nil {
        logger.Fatalf("Failed to initialize Redis: %v", err)
    }
    defer cache.Close()
    
    // Initialize server
    srv := server.New(server.Config{
        Port:   cfg.Server.Port,
        Logger: logger,DB: db,Cache: cache,
    })

    // Start server
    if err := srv.Start(); err != nil {
        logger.Fatalf("Failed to start server: %v", err)
    }

    // Wait for interrupt signal to gracefully shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    logger.Info("Shutting down server...")

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        logger.Errorf("Server forced to shutdown: %v", err)
    }

    logger.Info("Server exited")
}
