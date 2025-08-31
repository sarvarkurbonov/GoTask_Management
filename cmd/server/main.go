package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"GoTask_Management/internal/api"
	"GoTask_Management/internal/scheduler"
	"GoTask_Management/internal/storage"
	"GoTask_Management/internal/task"

	"github.com/spf13/viper"
)

func main() {
	log.Println("🚀 Starting GoTask Management Server...")

	// Load configuration with enhanced defaults
	setupConfig()

	// Initialize storage using the new factory pattern
	store, err := initializeStorage()
	if err != nil {
		log.Fatalf("❌ Failed to initialize storage: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("⚠️ Error closing storage: %v", err)
		}
	}()

	// Perform health check on storage
	if err := performStorageHealthCheck(store); err != nil {
		log.Fatalf("❌ Storage health check failed: %v", err)
	}

	// Initialize service
	taskService := task.NewService(store)

	// Start scheduler if enabled
	var sched *scheduler.Scheduler
	if viper.GetBool("scheduler.enabled") {
		sched = scheduler.New(taskService, viper.GetInt("scheduler.interval"))
		sched.Start()
		defer sched.Stop()
		log.Printf("📅 Scheduler started with %d second interval", viper.GetInt("scheduler.interval"))
	}

	// Initialize API server
	server := api.NewServer(taskService, viper.GetInt("server.port"))

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in a goroutine
	go func() {
		port := viper.GetInt("server.port")
		storageType := viper.GetString("storage.type")

		log.Printf("🌐 Server starting on port %d with %s storage", port, storageType)
		log.Printf("📡 API available at: http://localhost:%d", port)
		log.Printf("🏥 Health check: http://localhost:%d/health", port)

		if err := server.Start(); err != nil {
			log.Printf("❌ Server error: %v", err)
			cancel()
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Handle force quit (second Ctrl+C)
	forceQuit := make(chan os.Signal, 1)
	signal.Notify(forceQuit, syscall.SIGINT, syscall.SIGTERM)

	// Ensure we stop listening for signals when done
	defer signal.Stop(quit)
	defer signal.Stop(forceQuit)

	select {
	case <-quit:
		log.Println("🛑 Shutdown signal received...")

		// Start listening for second signal for force quit
		go func() {
			<-forceQuit
			log.Println("🚨 Force quit signal received, exiting immediately!")
			os.Exit(1)
		}()

	case <-ctx.Done():
		log.Println("🛑 Context cancelled...")
	}

	// Graceful shutdown with timeout
	log.Println("🔄 Shutting down server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Create a channel to signal when shutdown is complete
	shutdownComplete := make(chan error, 1)

	go func() {
		shutdownComplete <- server.Shutdown()
	}()

	// Wait for shutdown to complete or timeout
	select {
	case err := <-shutdownComplete:
		if err != nil {
			log.Printf("⚠️ Error during server shutdown: %v", err)
			os.Exit(1)
		} else {
			log.Println("✅ Server stopped gracefully")
		}
	case <-shutdownCtx.Done():
		log.Println("⚠️ Shutdown timeout exceeded, forcing exit")
		os.Exit(1)
	}

	// Final cleanup
	log.Println("🏁 Application shutdown complete")
}

// setupConfig initializes Viper configuration with enhanced defaults
func setupConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Server configuration
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "15s")
	viper.SetDefault("server.write_timeout", "15s")
	viper.SetDefault("server.idle_timeout", "60s")

	// Storage configuration
	viper.SetDefault("storage.type", "json")
	viper.SetDefault("storage.path", "tasks.json")

	// Database configuration
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "gotask")
	viper.SetDefault("database.user", "gotask_user")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.timezone", "UTC")

	// MongoDB configuration
	viper.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongodb.database", "gotask")
	viper.SetDefault("mongodb.collection", "tasks")
	viper.SetDefault("mongodb.connect_timeout", "10s")
	viper.SetDefault("mongodb.query_timeout", "5s")

	// Scheduler configuration
	viper.SetDefault("scheduler.enabled", true)
	viper.SetDefault("scheduler.interval", 300)

	// Logging configuration
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "text")

	// Allow environment variables to override config
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("📄 Config file not found, using defaults and environment variables: %v", err)
	} else {
		log.Printf("📄 Configuration loaded from: %s", viper.ConfigFileUsed())
	}
}

// initializeStorage creates and configures the storage backend
func initializeStorage() (storage.Storage, error) {
	storageType := viper.GetString("storage.type")

	// Create storage configuration
	config := &storage.StorageConfig{
		Type:     storage.StorageType(storageType),
		FilePath: viper.GetString("storage.path"),
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.name"),
		SSLMode:  viper.GetString("database.ssl_mode"),
		TimeZone: viper.GetString("database.timezone"),
		Charset:  viper.GetString("database.charset"),
		ParseTime: viper.GetBool("database.parse_time"),
		Loc:      viper.GetString("database.location"),
		URI:      viper.GetString("mongodb.uri"),
		Collection: viper.GetString("mongodb.collection"),
		ConnectTimeout: viper.GetDuration("mongodb.connect_timeout"),
		QueryTimeout:   viper.GetDuration("mongodb.query_timeout"),
	}

	// Validate configuration
	if err := storage.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid storage configuration: %w", err)
	}

	// Create storage instance
	store, err := storage.NewStorage(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	log.Printf("✅ Storage initialized: %s", storageType)
	return store, nil
}

// performStorageHealthCheck checks if the storage backend is healthy
func performStorageHealthCheck(store storage.Storage) error {
	if healthChecker, ok := store.(interface{ HealthCheck() error }); ok {
		if err := healthChecker.HealthCheck(); err != nil {
			return fmt.Errorf("storage health check failed: %w", err)
		}
		log.Println("✅ Storage health check passed")
	} else {
		log.Println("ℹ️ Storage does not support health checks")
	}
	return nil
}
