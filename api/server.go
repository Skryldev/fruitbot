package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"fruitbot/api/handlers"
	"fruitbot/api/middleware"
	"fruitbot/internal/client"

	"go.uber.org/zap"
)

type ServerConfig struct {
	Host        string
	Port        int
	APIToken    string
	Logger      *zap.Logger
	MultiClient *client.MultiClient
	Version     string
	
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type Server struct {
	cfg    *ServerConfig
	http   *http.Server
	logger *zap.Logger
	multi  *client.MultiClient
}

func NewServer(cfg *ServerConfig) *Server {
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 30 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 30 * time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 120 * time.Second
	}

	s := &Server{
		cfg:    cfg,
		logger: cfg.Logger,
		multi:  cfg.MultiClient,
	}

	mux := http.NewServeMux()

	var handler http.Handler = mux
	
	handler = middleware.Recovery(cfg.Logger)(handler)
	
	handler = middleware.Logging(cfg.Logger)(handler)
	
	handler = middleware.CORS()(handler)
	
	if cfg.APIToken != "" {
		handler = middleware.APIAuth(cfg.APIToken)(handler)
		cfg.Logger.Info("API authentication enabled")
	}

	h := handlers.NewHandlers(cfg.MultiClient, cfg.Logger, cfg.Version)
	mh := handlers.NewMultiAccountHandlers(cfg.MultiClient, cfg.Logger)

	registerRoutes(mux, h, mh)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"ok","version":"%s","message":"FruitBot API Server"}`, cfg.Version)
			return
		}
		http.NotFound(w, r)
	})

	s.http = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		
		MaxHeaderBytes: 1 << 20,
	}

	return s
}

func (s *Server) ListenAndServe() error {
	s.logger.Info("🚀 Server starting",
		zap.String("addr", s.http.Addr),
		zap.String("version", s.cfg.Version),
		zap.Bool("auth_enabled", s.cfg.APIToken != ""),
		zap.Duration("read_timeout", s.cfg.ReadTimeout),
		zap.Duration("write_timeout", s.cfg.WriteTimeout),
	)
	
	s.logger.Debug("Available endpoints",
		zap.Strings("routes", []string{
			"GET  /health",
			"GET  /api/status",
			"GET  /api/stats",
			"POST /api/player/load",
			"GET  /api/player/info",
			"POST /api/cards/collect-gold",
			"POST /api/store/buy-pack",
			"POST /api/tribe/message",
			"GET  /api/accounts",
			"POST /api/accounts",
			"DELETE /api/accounts/{name}",
			"POST /api/accounts/load-all",
			"POST /api/accounts/collect-gold-all",
			"POST /api/accounts/message-all",
			"GET  /api/accounts/stats",
		}),
	)
	
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("🛑 Shutting down server...")
	
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	if s.multi != nil {
		s.logger.Info("Closing multi-client connections...")
		if err := s.multi.Close(); err != nil {
			s.logger.Error("Error closing multi-client", zap.Error(err))
		}
	}
	
	if err := s.http.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}
	
	s.logger.Info("✅ Server stopped gracefully")
	return nil
}

func (s *Server) Addr() string {
	return s.http.Addr
}

func (s *Server) IsReady() bool {
	return s.multi != nil && s.multi.Stats().TotalAccounts >= 0
}