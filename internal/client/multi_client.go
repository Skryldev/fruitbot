package client

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"fruitbot/internal/infrastructure/config"
	"fruitbot/internal/infrastructure/session"

	"go.uber.org/zap"
)

// ============================================================
// Account Types
// ============================================================

// AccountClient wraps a Client for a specific account
type AccountClient struct {
	Name   string
	Client *Client
	Opts   *AccountOptions
	mu     sync.RWMutex
}

// AccountOptions holds account-specific options
type AccountOptions struct {
	RestoreKey  string
	Passport    string
	UDID        string
	MobileModel string
	DeviceName  string
	StoreType   string
}

// ============================================================
// Multi-Client
// ============================================================

// MultiClient manages multiple game accounts concurrently
type MultiClient struct {
	accounts    map[string]*AccountClient
	mu          sync.RWMutex
	sessionPool *session.SessionPool
	logger      *zap.Logger

	// Shared resources
	deviceFingerprinter *config.DeviceFingerprinter

	// Metrics
	activeAccounts int64
	totalRequests  int64
}

// MultiClientConfig holds multi-client configuration
type MultiClientConfig struct {
	Logger     *zap.Logger
	BaseURL    string
	EncVersion int
}

// NewMultiClient creates a new multi-account client
func NewMultiClient(cfg *MultiClientConfig) (*MultiClient, error) {
	if cfg == nil {
		cfg = &MultiClientConfig{}
	}

	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}

	// Create session pool
	pool, err := session.NewSessionPool(&session.PoolConfig{
		Logger:          cfg.Logger,
		MaxSessions:     100,
		CleanupInterval: 10 * time.Minute,
		SessionTTL:      24 * time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session pool: %w", err)
	}

	return &MultiClient{
		accounts:            make(map[string]*AccountClient),
		sessionPool:         pool,
		logger:              cfg.Logger,
		deviceFingerprinter: config.NewDeviceFingerprinter(),
	}, nil
}

// AddAccount adds a new account
// restore_key is the main account identifier (required)
func (mc *MultiClient) AddAccount(ctx context.Context, name string, opts *AccountOptions) (*AccountClient, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check if account already exists
	if existing, ok := mc.accounts[name]; ok {
		existing.mu.Lock()
		existing.Opts = opts // Update options
		existing.mu.Unlock()
		return existing, nil
	}

	// Get or create session
	sessionOpts := &session.SessionOptions{
		RestoreKey:  opts.RestoreKey,
		Passport:    opts.Passport,
		UDID:        opts.UDID,
		MobileModel: opts.MobileModel,
		DeviceName:  opts.DeviceName,
		StoreType:   opts.StoreType,
	}

	_, err := mc.sessionPool.GetSession(ctx, name, sessionOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Build client options - استفاده از WithRestoreKey به جای WithCredentials
	var clientOpts []Option
	
	// Always set restore_key as the main identifier
	if opts.RestoreKey != "" {
		clientOpts = append(clientOpts, WithRestoreKey(opts.RestoreKey))
	}
	
	// Set session name (usually same as account name)
	clientOpts = append(clientOpts, WithSessionName(name))

	// Set device info if provided
	if opts.MobileModel != "" {
		clientOpts = append(clientOpts, WithDeviceInfo(
			opts.MobileModel,
			opts.DeviceName,
			opts.StoreType,
		))
	}

	clientOpts = append(clientOpts, WithLogger(mc.logger))

	gameClient, err := NewClient(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	account := &AccountClient{
		Name:   name,
		Client: gameClient,
		Opts:   opts,
	}

	mc.accounts[name] = account
	atomic.AddInt64(&mc.activeAccounts, 1)

	mc.logger.Info("Account added",
		zap.String("name", name),
		zap.String("restore_key", func() string {
			if len(opts.RestoreKey) > 8 {
				return opts.RestoreKey[:8] + "..."
			}
			return opts.RestoreKey
		}()),
		zap.Int("total_accounts", len(mc.accounts)),
	)

	return account, nil
}

// GetAccount returns an account by name
func (mc *MultiClient) GetAccount(name string) (*AccountClient, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	account, ok := mc.accounts[name]
	if !ok {
		return nil, fmt.Errorf("account not found: %s", name)
	}

	return account, nil
}

// RemoveAccount removes an account
func (mc *MultiClient) RemoveAccount(ctx context.Context, name string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	account, ok := mc.accounts[name]
	if !ok {
		return fmt.Errorf("account not found: %s", name)
	}

	// Close the client
	account.Client.Close()

	// Release session
	mc.sessionPool.ReleaseSession(name)

	delete(mc.accounts, name)
	atomic.AddInt64(&mc.activeAccounts, -1)

	mc.logger.Info("Account removed",
		zap.String("name", name),
		zap.Int("remaining", len(mc.accounts)),
	)

	return nil
}

// ListAccounts returns all account names
func (mc *MultiClient) ListAccounts() []string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	names := make([]string, 0, len(mc.accounts))
	for name := range mc.accounts {
		names = append(names, name)
	}
	return names
}

// ExecuteOnAll executes a function on all accounts concurrently
func (mc *MultiClient) ExecuteOnAll(ctx context.Context, fn func(*AccountClient) error) map[string]error {
	mc.mu.RLock()
	accounts := make([]*AccountClient, 0, len(mc.accounts))
	for _, acc := range mc.accounts {
		accounts = append(accounts, acc)
	}
	mc.mu.RUnlock()

	results := make(map[string]error, len(accounts))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, acc := range accounts {
		wg.Add(1)
		go func(ac *AccountClient) {
			defer wg.Done()

			err := fn(ac)

			mu.Lock()
			results[ac.Name] = err
			mu.Unlock()
		}(acc)
	}

	wg.Wait()
	return results
}

// ExecuteOnAccount executes a function on a specific account
func (mc *MultiClient) ExecuteOnAccount(ctx context.Context, name string, fn func(*AccountClient) error) error {
	account, err := mc.GetAccount(name)
	if err != nil {
		return err
	}

	return fn(account)
}

// LoadAllPlayers loads all players concurrently
func (mc *MultiClient) LoadAllPlayers(ctx context.Context) map[string]error {
	return mc.ExecuteOnAll(ctx, func(ac *AccountClient) error {
		_, err := ac.Client.LoadPlayer(ctx, &LoadPlayerParams{
			SaveSession: true,
		})
		return err
	})
}

// SendMessageToAllTribes sends a message to all tribes
func (mc *MultiClient) SendMessageToAllTribes(ctx context.Context, message string) map[string]error {
	return mc.ExecuteOnAll(ctx, func(ac *AccountClient) error {
		_, err := ac.Client.SendTribeMessage(ctx, message)
		return err
	})
}

// CollectGoldOnAll collects gold on all accounts
func (mc *MultiClient) CollectGoldOnAll(ctx context.Context) map[string]error {
	return mc.ExecuteOnAll(ctx, func(ac *AccountClient) error {
		_, err := ac.Client.CollectMinedGold(ctx)
		return err
	})
}

// ============================================================
// Statistics
// ============================================================

// MultiClientStats holds multi-client statistics
type MultiClientStats struct {
	TotalAccounts  int64    `json:"total_accounts"`
	ActiveAccounts int64    `json:"active_accounts"`
	TotalRequests  int64    `json:"total_requests"`
	Accounts       []string `json:"accounts"`
}

// Stats returns statistics
func (mc *MultiClient) Stats() MultiClientStats {
	return MultiClientStats{
		TotalAccounts:  int64(len(mc.accounts)),
		ActiveAccounts: atomic.LoadInt64(&mc.activeAccounts),
		TotalRequests:  atomic.LoadInt64(&mc.totalRequests),
		Accounts:       mc.ListAccounts(),
	}
}

// Close closes all accounts and the session pool
func (mc *MultiClient) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for name, acc := range mc.accounts {
		acc.Client.Close()
		mc.sessionPool.ReleaseSession(name)
	}

	return mc.sessionPool.Close()
}