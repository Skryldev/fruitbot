package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fruitbot/api"
	"fruitbot/internal/client"

	"go.uber.org/zap"
)

var Version = "2.0.0"

func main() {
	var (
		port     = flag.Int("port", 8080, "HTTP server port")
		host     = flag.String("host", "0.0.0.0", "HTTP server host")
		apiKey   = flag.String("api-key", "", "API key for authentication (empty = no auth)")
		logLevel = flag.String("log-level", "info", "Log level: debug, info, warn, error")
	)
	flag.Parse()

	logger, err := createLogger(*logLevel)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	printBanner(Version, *host, *port)

	multiClient, err := client.NewMultiClient(&client.MultiClientConfig{
		Logger: logger,
	})
	if err != nil {
		logger.Fatal("Failed to create multi-client", zap.Error(err))
	}

	serverConfig := &api.ServerConfig{
		Host:        *host,
		Port:        *port,
		APIToken:    *apiKey,
		Logger:      logger,
		MultiClient: multiClient,
		Version:     Version,
	}

	server := api.NewServer(serverConfig)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigCh
		logger.Info("Received signal, shutting down...",
			zap.String("signal", sig.String()),
		)

		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Server shutdown error", zap.Error(err))
		}

		cancel()
	}()

	logger.Info("🚀 Server starting...")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}

func createLogger(level string) (*zap.Logger, error) {
	var config zap.Config

	switch level {
	case "debug":
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	return config.Build()
}

func printBanner(version, host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║        🍎 FruitBot API Server Go Edition 🍎        ║")
	fmt.Printf("║        Version: %s                              ║\n", version)
	fmt.Println("║                                                      ║")
	fmt.Printf("║        Mode: Multi-Session                          ║\n")
	fmt.Printf("║        Listening on: %-25s ║\n", addr)
	fmt.Println("║                                                      ║")
	fmt.Println("║  📚 API Endpoints:                                  ║")
	fmt.Printf("║  Health:  GET  http://%s/health              ║\n", addr)
	fmt.Printf("║  Load:    POST http://%s/api/player/load     ║\n", addr)
	fmt.Printf("║  Gold:    POST http://%s/api/cards/collect-gold║\n", addr)
	fmt.Println("║                                                      ║")
	fmt.Println("║  💡 Usage:                                          ║")
	fmt.Println("║  Send restore_key in request body to manage         ║")
	fmt.Println("║  multiple accounts automatically.                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()
}