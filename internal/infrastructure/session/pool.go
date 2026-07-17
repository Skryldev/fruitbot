package session

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type SessionPool struct {
	sessions map[string]*ManagedSession
	mu       sync.RWMutex
	storage  StorageBackend
	logger   *zap.Logger

	// Metrics
	activeSessions  int64
	totalSessions   int64
	expiredSessions int64

	// Cleanup
	cleanupInterval time.Duration
	sessionTTL      time.Duration
	stopCh          chan struct{}
}

type ManagedSession struct {
	Name      string
	Session   *SessionManager
	Options   *SessionOptions
	LastUsed  time.Time
	CreatedAt time.Time
	IsActive  bool
	mu        sync.RWMutex
}

type SessionOptions struct {
	RestoreKey  string
	Passport    string
	UDID        string
	MobileModel string
	DeviceName  string
	StoreType   string
	BaseURL     string
}

type PoolConfig struct {
	Storage         StorageBackend
	Logger          *zap.Logger
	MaxSessions     int
	CleanupInterval time.Duration
	SessionTTL      time.Duration
}

func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		Logger:          zap.NewNop(),
		MaxSessions:     100,
		CleanupInterval: 10 * time.Minute,
		SessionTTL:      24 * time.Hour,
	}
}

func NewSessionPool(cfg *PoolConfig) (*SessionPool, error) {
	if cfg == nil {
		cfg = DefaultPoolConfig()
	}

	if cfg.Storage == nil {
		storage, err := NewFileStorage(&FileStorageConfig{
			Directory: "sessions",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create storage: %w", err)
		}
		cfg.Storage = storage
	}

	sp := &SessionPool{
		sessions:        make(map[string]*ManagedSession),
		storage:         cfg.Storage,
		logger:          cfg.Logger,
		cleanupInterval: cfg.CleanupInterval,
		sessionTTL:      cfg.SessionTTL,
		stopCh:          make(chan struct{}),
	}

	go sp.cleanupLoop()

	if err := sp.loadExistingSessions(); err != nil {
		sp.logger.Warn("Failed to load existing sessions", zap.Error(err))
	}

	return sp, nil
}

func (sp *SessionPool) GetSession(ctx context.Context, name string, opts *SessionOptions) (*ManagedSession, error) {
	sp.mu.RLock()
	if session, exists := sp.sessions[name]; exists {
		sp.mu.RUnlock()
		session.mu.Lock()
		session.LastUsed = time.Now()
		session.IsActive = true

		if opts != nil {
			session.Options = opts
		}
		session.mu.Unlock()
		return session, nil
	}
	sp.mu.RUnlock()

	return sp.createSession(ctx, name, opts)
}

func (sp *SessionPool) createSession(ctx context.Context, name string, opts *SessionOptions) (*ManagedSession, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if session, exists := sp.sessions[name]; exists {
		return session, nil
	}

	sessionMgr, err := NewSessionManager(&SessionConfig{
		Storage:     sp.storage,
		SessionName: name,
		Logger:      sp.logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session manager: %w", err)
	}

	exists, _ := sessionMgr.DoesSessionExist(ctx)

	if exists {
		sessionData, err := sessionMgr.LoadSessionData(ctx)
		if err != nil {
			sp.logger.Warn("Failed to load session data, creating new",
				zap.String("name", name),
				zap.Error(err),
			)
		} else if opts == nil || opts.RestoreKey == "" {
			opts = &SessionOptions{
				RestoreKey:  sessionData.RestoreKey,
				Passport:    sessionData.Passport,
				UDID:        sessionData.UDID,
				MobileModel: sessionData.MobileModel,
			}
		}
	}

	if opts == nil {
		opts = &SessionOptions{}
	}

	managed := &ManagedSession{
		Name:      name,
		Session:   sessionMgr,
		Options:   opts,
		LastUsed:  time.Now(),
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	sp.sessions[name] = managed
	sp.totalSessions++
	sp.activeSessions++

	sp.logger.Info("Session created",
		zap.String("name", name),
		zap.Bool("existing", exists),
	)

	return managed, nil
}

func (sp *SessionPool) ReleaseSession(name string) {
	sp.mu.RLock()
	session, exists := sp.sessions[name]
	sp.mu.RUnlock()

	if !exists {
		return
	}

	session.mu.Lock()
	session.IsActive = false
	session.mu.Unlock()

	atomic.AddInt64(&sp.activeSessions, -1)
}

func (sp *SessionPool) RemoveSession(ctx context.Context, name string) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	session, exists := sp.sessions[name]
	if !exists {
		return fmt.Errorf("session not found: %s", name)
	}

	if err := session.Session.DeleteSession(ctx); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	delete(sp.sessions, name)
	sp.totalSessions--

	if session.IsActive {
		sp.activeSessions--
	}

	sp.logger.Info("Session removed", zap.String("name", name))
	return nil
}

func (sp *SessionPool) ListSessions() []string {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	names := make([]string, 0, len(sp.sessions))
	for name := range sp.sessions {
		names = append(names, name)
	}
	return names
}

func (sp *SessionPool) GetActiveSessions() []*ManagedSession {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	active := make([]*ManagedSession, 0)
	for _, session := range sp.sessions {
		session.mu.RLock()
		if session.IsActive {
			active = append(active, session)
		}
		session.mu.RUnlock()
	}
	return active
}

type PoolStats struct {
	TotalSessions   int64
	ActiveSessions  int64
	ExpiredSessions int64
	SessionNames    []string
}

func (sp *SessionPool) Stats() PoolStats {
	return PoolStats{
		TotalSessions:   atomic.LoadInt64(&sp.totalSessions),
		ActiveSessions:  atomic.LoadInt64(&sp.activeSessions),
		ExpiredSessions: atomic.LoadInt64(&sp.expiredSessions),
		SessionNames:    sp.ListSessions(),
	}
}

func (sp *SessionPool) cleanupLoop() {
	ticker := time.NewTicker(sp.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sp.stopCh:
			return
		case <-ticker.C:
			sp.cleanup()
		}
	}
}

func (sp *SessionPool) cleanup() {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	now := time.Now()

	for name, session := range sp.sessions {
		session.mu.RLock()
		inactive := !session.IsActive
		lastUsed := session.LastUsed
		session.mu.RUnlock()

		if inactive && now.Sub(lastUsed) > sp.sessionTTL {
			sp.logger.Info("Cleaning up expired session",
				zap.String("name", name),
				zap.Duration("idle", now.Sub(lastUsed)),
			)

			delete(sp.sessions, name)
			sp.totalSessions--
			sp.expiredSessions++
		}
	}
}

func (sp *SessionPool) loadExistingSessions() error {
	ctx := context.Background()

	names, err := sp.storage.List(ctx)
	if err != nil {
		return err
	}

	for _, name := range names {
		sp.sessions[name] = &ManagedSession{
			Name:      name,
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
			IsActive:  false,
		}
		sp.totalSessions++
	}

	sp.logger.Info("Loaded existing sessions",
		zap.Int("count", len(names)),
	)

	return nil
}

func (sp *SessionPool) Close() error {
	close(sp.stopCh)
	return sp.storage.Close()
}