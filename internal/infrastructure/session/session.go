package session

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/nacl/secretbox"
)

type PlayerInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	InviteKey string `json:"invite_key"`
}

type SessionData struct {
	RestoreKey  string      `json:"restore_key"`
	Passport    string      `json:"passport"`
	UDID        string      `json:"udid"`
	MobileModel string      `json:"mobile_model"`
	Player      *PlayerInfo `json:"player"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (sd *SessionData) Validate() error {
	if sd.Passport == "" {
		return fmt.Errorf("passport is required")
	}
	if sd.UDID == "" {
		return fmt.Errorf("UDID is required")
	}
	if sd.Player == nil {
		return fmt.Errorf("player info is required")
	}
	if sd.Player.ID <= 0 {
		return fmt.Errorf("player ID must be positive")
	}
	return nil
}

type StorageBackend interface {
	Save(ctx context.Context, name string, data *SessionData) error
	Load(ctx context.Context, name string) (*SessionData, error)
	Exists(ctx context.Context, name string) (bool, error)
	Delete(ctx context.Context, name string) error
	List(ctx context.Context) ([]string, error)
	Close() error
}

// ============================================================
// File Storage Backend
// ============================================================

type FileStorage struct {
	directory string
	ext       string
	logger    *zap.Logger
	mu        sync.RWMutex
	
	secretKey *[32]byte
}

type FileStorageConfig struct {
	Directory string
	Extension string
	SecretKey []byte
	Logger    *zap.Logger
}

func NewFileStorage(cfg *FileStorageConfig) (*FileStorage, error) {
	if cfg == nil {
		cfg = &FileStorageConfig{
			Directory: "sessions",
			Extension: ".fb",
		}
	}
	
	if cfg.Directory == "" {
		cfg.Directory = "sessions"
	}
	if cfg.Extension == "" {
		cfg.Extension = ".fb"
	}
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
	
	if err := os.MkdirAll(cfg.Directory, 0700); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}
	
	fs := &FileStorage{
		directory: cfg.Directory,
		ext:       cfg.Extension,
		logger:    cfg.Logger,
	}
	
	if len(cfg.SecretKey) > 0 {
		if len(cfg.SecretKey) != 32 {
			return nil, fmt.Errorf("secret key must be 32 bytes, got %d", len(cfg.SecretKey))
		}
		var key [32]byte
		copy(key[:], cfg.SecretKey)
		fs.secretKey = &key
	}
	
	return fs, nil
}

func (fs *FileStorage) filePath(name string) string {
	return filepath.Join(fs.directory, name+fs.ext)
}

func (fs *FileStorage) Save(ctx context.Context, name string, data *SessionData) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	if data == nil {
		return fmt.Errorf("session data is nil")
	}
	
	data.UpdatedAt = time.Now()
	if data.CreatedAt.IsZero() {
		data.CreatedAt = data.UpdatedAt
	}
	
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	
	if fs.secretKey != nil {
		jsonData, err = fs.encrypt(jsonData)
		if err != nil {
			return fmt.Errorf("failed to encrypt session: %w", err)
		}
	}
	
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	tmpPath := fs.filePath(name + ".tmp")
	if err := os.WriteFile(tmpPath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}
	
	if err := os.Rename(tmpPath, fs.filePath(name)); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename session file: %w", err)
	}
	
	fs.logger.Debug("Session saved", zap.String("name", name))
	return nil
}

func (fs *FileStorage) Load(ctx context.Context, name string) (*SessionData, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	
	fs.mu.RLock()
	data, err := os.ReadFile(fs.filePath(name))
	fs.mu.RUnlock()
	
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session file not found: %s", name)
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}
	
	if fs.secretKey != nil {
		data, err = fs.decrypt(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt session: %w", err)
		}
	}
	
	var sessionData SessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	
	return &sessionData, nil
}

func (fs *FileStorage) Exists(ctx context.Context, name string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}
	
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	
	_, err := os.Stat(fs.filePath(name))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (fs *FileStorage) Delete(ctx context.Context, name string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	if err := os.Remove(fs.filePath(name)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	
	return nil
}

func (fs *FileStorage) List(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	
	files, err := filepath.Glob(filepath.Join(fs.directory, "*"+fs.ext))
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	
	names := make([]string, 0, len(files))
	for _, f := range files {
		base := filepath.Base(f)
		name := base[:len(base)-len(fs.ext)]
		names = append(names, name)
	}
	
	return names, nil
}

func (fs *FileStorage) Close() error {
	return nil
}

func (fs *FileStorage) encrypt(plaintext []byte) ([]byte, error) {
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	
	encrypted := secretbox.Seal(nonce[:], plaintext, &nonce, fs.secretKey)
	return encrypted, nil
}

func (fs *FileStorage) decrypt(encrypted []byte) ([]byte, error) {
	var nonce [24]byte
	copy(nonce[:], encrypted[:24])
	
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &nonce, fs.secretKey)
	if !ok {
		return nil, fmt.Errorf("failed to decrypt session data")
	}
	
	return decrypted, nil
}

// ============================================================
// Memory Storage Backend (for testing/caching)
// ============================================================

type MemoryStorage struct {
	data  map[string]*SessionData
	mu    sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]*SessionData),
	}
}

func (ms *MemoryStorage) Save(ctx context.Context, name string, data *SessionData) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	data.UpdatedAt = time.Now()
	if data.CreatedAt.IsZero() {
		data.CreatedAt = data.UpdatedAt
	}
	
	ms.mu.Lock()
	ms.data[name] = data
	ms.mu.Unlock()
	
	return nil
}

func (ms *MemoryStorage) Load(ctx context.Context, name string) (*SessionData, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	
	ms.mu.RLock()
	data, ok := ms.data[name]
	ms.mu.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("session not found: %s", name)
	}
	
	return data, nil
}

func (ms *MemoryStorage) Exists(ctx context.Context, name string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}
	
	ms.mu.RLock()
	_, ok := ms.data[name]
	ms.mu.RUnlock()
	
	return ok, nil
}

func (ms *MemoryStorage) Delete(ctx context.Context, name string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	ms.mu.Lock()
	delete(ms.data, name)
	ms.mu.Unlock()
	
	return nil
}

func (ms *MemoryStorage) List(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	names := make([]string, 0, len(ms.data))
	for name := range ms.data {
		names = append(names, name)
	}
	
	return names, nil
}

func (ms *MemoryStorage) Close() error {
	return nil
}

// ============================================================
// Session Manager
// ============================================================

type SessionManager struct {
	storage     StorageBackend
	sessionName string
	logger      *zap.Logger
	
	cache   *SessionData
	cacheMu sync.RWMutex
}

type SessionConfig struct {
	Storage     StorageBackend
	SessionName string
	Logger      *zap.Logger
}

func NewSessionManager(cfg *SessionConfig) (*SessionManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.Storage == nil {
		return nil, fmt.Errorf("storage backend is required")
	}
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
	
	return &SessionManager{
		storage:     cfg.Storage,
		sessionName: cfg.SessionName,
		logger:      cfg.Logger,
	}, nil
}

func (sm *SessionManager) SessionName() string {
	return sm.sessionName
}

func (sm *SessionManager) SetSessionName(name string) {
	sm.sessionName = name
	sm.cacheMu.Lock()
	sm.cache = nil
	sm.cacheMu.Unlock()
}

func (sm *SessionManager) DoesSessionExist(ctx context.Context) (bool, error) {
	if sm.sessionName == "" {
		return false, nil
	}
	return sm.storage.Exists(ctx, sm.sessionName)
}

func (sm *SessionManager) LoadSessionData(ctx context.Context) (*SessionData, error) {
	if sm.sessionName == "" {
		return nil, fmt.Errorf("session name is not set")
	}
	
	sm.cacheMu.RLock()
	if sm.cache != nil {
		cached := *sm.cache
		sm.cacheMu.RUnlock()
		return &cached, nil
	}
	sm.cacheMu.RUnlock()
	
	data, err := sm.storage.Load(ctx, sm.sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}
	
	sm.cacheMu.Lock()
	sm.cache = data
	sm.cacheMu.Unlock()
	
	sm.logger.Info("Session loaded", 
		zap.String("name", sm.sessionName),
		zap.Int("player_id", data.Player.ID),
	)
	
	return data, nil
}

func (sm *SessionManager) SaveSession(ctx context.Context, playerInfo *PlayerInfo) error {
	if sm.sessionName == "" {
		return fmt.Errorf("session name is not set")
	}
	
	sessionData, err := sm.LoadSessionData(ctx)
	if err != nil {
		sessionData = &SessionData{}
	}
	
	if playerInfo != nil {
		sessionData.Player = playerInfo
	}
	
	if sessionData.CreatedAt.IsZero() {
		sessionData.CreatedAt = time.Now()
	}
	sessionData.UpdatedAt = time.Now()
	
	if err := sm.storage.Save(ctx, sm.sessionName, sessionData); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	
	sm.cacheMu.Lock()
	sm.cache = sessionData
	sm.cacheMu.Unlock()
	
	sm.logger.Info("Session saved",
		zap.String("name", sm.sessionName),
		zap.Int("player_id", sessionData.Player.ID),
	)
	
	return nil
}

func (sm *SessionManager) DeleteSession(ctx context.Context) error {
	if sm.sessionName == "" {
		return fmt.Errorf("session name is not set")
	}
	
	sm.cacheMu.Lock()
	sm.cache = nil
	sm.cacheMu.Unlock()
	
	return sm.storage.Delete(ctx, sm.sessionName)
}

func (sm *SessionManager) ListSessions(ctx context.Context) ([]string, error) {
	return sm.storage.List(ctx)
}

func (sm *SessionManager) UpdateRestoreKey(ctx context.Context, restoreKey string) error {
	data, err := sm.LoadSessionData(ctx)
	if err != nil {
		return err
	}
	
	data.RestoreKey = restoreKey
	return sm.storage.Save(ctx, sm.sessionName, data)
}

func (sm *SessionManager) UpdatePassport(ctx context.Context, passport string) error {
	data, err := sm.LoadSessionData(ctx)
	if err != nil {
		return err
	}
	
	data.Passport = passport
	return sm.storage.Save(ctx, sm.sessionName, data)
}

func (sm *SessionManager) Close() error {
	return sm.storage.Close()
}