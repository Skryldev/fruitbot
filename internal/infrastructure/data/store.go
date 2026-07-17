package data

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

var embeddedFS embed.FS

type Store struct {
	baseDir string
	logger  *zap.Logger
	
	// In-memory cache with TTL
	cache    map[string]*cacheEntry
	cacheMu  sync.RWMutex
	cacheTTL time.Duration
	
	// File watcher for hot-reload
	watcher  *fileWatcher
	stopCh   chan struct{}
	
	// Metrics
	loadCount   uint64
	cacheHits   uint64
	cacheMisses uint64
}

type cacheEntry struct {
	data      interface{}
	timestamp time.Time
}

type StoreConfig struct {
	BaseDir  string
	CacheTTL time.Duration
	Logger   *zap.Logger
	
	// Hot reload settings
	EnableHotReload bool
	WatchInterval   time.Duration
}

func DefaultStoreConfig() *StoreConfig {
	return &StoreConfig{
		BaseDir:         "data",
		CacheTTL:        5 * time.Minute,
		Logger:          zap.NewNop(),
		EnableHotReload: false,
		WatchInterval:   30 * time.Second,
	}
}

func NewStore(cfg *StoreConfig) (*Store, error) {
	if cfg == nil {
		cfg = DefaultStoreConfig()
	}
	
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
	
	if cfg.BaseDir != "" {
		if err := os.MkdirAll(cfg.BaseDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
	}
	
	s := &Store{
		baseDir:  cfg.BaseDir,
		logger:   cfg.Logger,
		cache:    make(map[string]*cacheEntry),
		cacheTTL: cfg.CacheTTL,
		stopCh:   make(chan struct{}),
	}
	
	if cfg.EnableHotReload {
		s.watcher = newFileWatcher(cfg.WatchInterval)
		go s.watchLoop()
	}
	
	return s, nil
}

// ============================================================
// JSON Operations
// ============================================================

func (s *Store) SaveJSON(fileName string, data interface{}) error {
	if fileName == "" {
		return fmt.Errorf("file name is required")
	}
	
	filePath := s.resolvePath(fileName)
	
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	if err := os.Rename(tmpFile, filePath); err != nil {
		os.Remove(tmpFile) // Clean up temp file
		return fmt.Errorf("failed to rename file: %w", err)
	}
	
	s.cacheMu.Lock()
	delete(s.cache, fileName)
	s.cacheMu.Unlock()
	
	s.logger.Debug("Data saved", zap.String("file", fileName))
	return nil
}

func (s *Store) LoadJSON(fileName string) (interface{}, error) {
	if fileName == "" {
		return nil, fmt.Errorf("file name is required")
	}
	
	if s.cacheTTL > 0 {
		if cached, ok := s.getFromCache(fileName); ok {
			s.cacheHits++
			return cached, nil
		}
		s.cacheMisses++
	}
	
	s.loadCount++
	
	data, err := s.loadFromFile(fileName)
	if err == nil {
		s.setCache(fileName, data)
		return data, nil
	}
	
	data, err = s.loadFromEmbedded(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", fileName, err)
	}
	
	s.logger.Debug("Data loaded from embedded", zap.String("file", fileName))
	
	s.setCache(fileName, data)
	
	return data, nil
}

func LoadJSONTyped[T any](s *Store, fileName string) (*T, error) {
	data, err := s.LoadJSON(fileName)
	if err != nil {
		return nil, err
	}
	
	var result T
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to type: %w", err)
	}
	
	return &result, nil
}

func (s *Store) LoadCardData() (map[string]interface{}, error) {
	data, err := s.LoadJSON("cards.json")
	if err != nil {
		return nil, err
	}
	
	result, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid card data format")
	}
	
	return result, nil
}

// ============================================================
// File Operations
// ============================================================

func (s *Store) FileExists(fileName string) bool {
	filePath := s.resolvePath(fileName)
	_, err := os.Stat(filePath)
	return err == nil
}

func (s *Store) DeleteFile(fileName string) error {
	filePath := s.resolvePath(fileName)
	
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", fileName)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	s.cacheMu.Lock()
	delete(s.cache, fileName)
	s.cacheMu.Unlock()
	
	return nil
}

func (s *Store) ListFiles() ([]string, error) {
	if s.baseDir == "" {
		return nil, fmt.Errorf("base directory not set")
	}
	
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			files = append(files, entry.Name())
		}
	}
	
	return files, nil
}

// ============================================================
// Cache Management
// ============================================================

func (s *Store) getFromCache(fileName string) (interface{}, bool) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	
	entry, exists := s.cache[fileName]
	if !exists {
		return nil, false
	}
	
	if time.Since(entry.timestamp) > s.cacheTTL {
		return nil, false
	}
	
	return entry.data, true
}

func (s *Store) setCache(fileName string, data interface{}) {
	if s.cacheTTL <= 0 {
		return
	}
	
	s.cacheMu.Lock()
	s.cache[fileName] = &cacheEntry{
		data:      data,
		timestamp: time.Now(),
	}
	s.cacheMu.Unlock()
}

func (s *Store) InvalidateCache() {
	s.cacheMu.Lock()
	s.cache = make(map[string]*cacheEntry)
	s.cacheMu.Unlock()
}

func (s *Store) ClearExpiredCache() {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	
	now := time.Now()
	for key, entry := range s.cache {
		if now.Sub(entry.timestamp) > s.cacheTTL {
			delete(s.cache, key)
		}
	}
}

// ============================================================
// Cache Statistics
// ============================================================

type CacheStats struct {
	Size       int
	Hits       uint64
	Misses     uint64
	LoadCount  uint64
	HitRatio   float64
}

func (s *Store) CacheStats() CacheStats {
	s.cacheMu.RLock()
	size := len(s.cache)
	s.cacheMu.RUnlock()
	
	total := s.cacheHits + s.cacheMisses
	var hitRatio float64
	if total > 0 {
		hitRatio = float64(s.cacheHits) / float64(total)
	}
	
	return CacheStats{
		Size:      size,
		Hits:      s.cacheHits,
		Misses:    s.cacheMisses,
		LoadCount: s.loadCount,
		HitRatio:  hitRatio,
	}
}

// ============================================================
// Internal Helpers
// ============================================================

func (s *Store) resolvePath(fileName string) string {
	if s.baseDir == "" {
		return fileName
	}
	return filepath.Join(s.baseDir, fileName)
}

func (s *Store) loadFromFile(fileName string) (interface{}, error) {
	filePath := s.resolvePath(fileName)
	
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var data interface{}
	if err := json.Unmarshal(fileData, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	return data, nil
}

func (s *Store) loadFromEmbedded(fileName string) (interface{}, error) {
	fileData, err := embeddedFS.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	
	var data interface{}
	if err := json.Unmarshal(fileData, &data); err != nil {
		return nil, fmt.Errorf("failed to parse embedded JSON: %w", err)
	}
	
	return data, nil
}

// ============================================================
// Hot Reload Watcher
// ============================================================

type fileWatcher struct {
	interval  time.Duration
	modTimes  map[string]time.Time
	mu        sync.Mutex
	callbacks []func(string)
}

func newFileWatcher(interval time.Duration) *fileWatcher {
	return &fileWatcher{
		interval: interval,
		modTimes: make(map[string]time.Time),
	}
}

func (fw *fileWatcher) addCallback(cb func(string)) {
	fw.mu.Lock()
	fw.callbacks = append(fw.callbacks, cb)
	fw.mu.Unlock()
}

func (s *Store) watchLoop() {
	if s.watcher == nil {
		return
	}
	
	ticker := time.NewTicker(s.watcher.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.checkForChanges()
		}
	}
}

func (s *Store) checkForChanges() {
	files, err := s.ListFiles()
	if err != nil {
		s.logger.Error("Failed to list files for watching", zap.Error(err))
		return
	}
	
	for _, file := range files {
		filePath := s.resolvePath(file)
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}
		
		s.watcher.mu.Lock()
		lastMod, exists := s.watcher.modTimes[file]
		currentMod := info.ModTime()
		
		if !exists || currentMod.After(lastMod) {
			s.watcher.modTimes[file] = currentMod
			
			s.cacheMu.Lock()
			delete(s.cache, file)
			s.cacheMu.Unlock()
			
			s.logger.Info("File changed, cache invalidated", zap.String("file", file))
			
			for _, cb := range s.watcher.callbacks {
				cb(file)
			}
		}
		s.watcher.mu.Unlock()
	}
}

func (s *Store) OnFileChange(callback func(fileName string)) {
	if s.watcher != nil {
		s.watcher.addCallback(callback)
	}
}

func (s *Store) Stop() {
	close(s.stopCh)
}