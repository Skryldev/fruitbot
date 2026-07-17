package data

import (
	"os"
	"testing"
	"time"
)

func TestSaveLoadJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "data_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	store, err := NewStore(&StoreConfig{
		BaseDir:  tmpDir,
		CacheTTL: 0, // Disable cache for testing
	})
	if err != nil {
		t.Fatal(err)
	}
	
	// Test data
	testData := map[string]interface{}{
		"player": map[string]interface{}{
			"name": "TestPlayer",
			"id":   12345,
		},
		"score": 1000,
	}
	
	// Save
	if err := store.SaveJSON("test.json", testData); err != nil {
		t.Fatal(err)
	}
	
	// Load
	loaded, err := store.LoadJSON("test.json")
	if err != nil {
		t.Fatal(err)
	}
	
	loadedMap, ok := loaded.(map[string]interface{})
	if !ok {
		t.Fatal("loaded data is not a map")
	}
	
	player, ok := loadedMap["player"].(map[string]interface{})
	if !ok {
		t.Fatal("player is not a map")
	}
	
	if player["name"] != "TestPlayer" {
		t.Errorf("player name = %v, want TestPlayer", player["name"])
	}
}

func TestLoadJSONTyped(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "data_typed_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	store, _ := NewStore(&StoreConfig{
		BaseDir:  tmpDir,
		CacheTTL: 0,
	})
	
	type CardInfo struct {
		Name    string `json:"name"`
		Power   int    `json:"power"`
		Level   int    `json:"level"`
	}
	
	type CardsFile struct {
		Cards []CardInfo `json:"cards"`
	}
	
	testData := CardsFile{
		Cards: []CardInfo{
			{Name: "Fire", Power: 100, Level: 5},
			{Name: "Water", Power: 80, Level: 4},
		},
	}
	
	store.SaveJSON("cards.json", testData)
	
	result, err := LoadJSONTyped[CardsFile](store, "cards.json")
	if err != nil {
		t.Fatal(err)
	}
	
	if len(result.Cards) != 2 {
		t.Errorf("cards count = %d, want 2", len(result.Cards))
	}
	
	if result.Cards[0].Name != "Fire" {
		t.Errorf("first card = %s, want Fire", result.Cards[0].Name)
	}
}

func TestFileOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "data_ops_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	store, _ := NewStore(&StoreConfig{
		BaseDir:  tmpDir,
		CacheTTL: 0,
	})
	
	// Test file doesn't exist
	if store.FileExists("nonexistent.json") {
		t.Error("file should not exist")
	}
	
	// Save and check existence
	store.SaveJSON("test.json", map[string]string{"key": "value"})
	if !store.FileExists("test.json") {
		t.Error("file should exist")
	}
	
	// List files
	files, _ := store.ListFiles()
	if len(files) != 1 || files[0] != "test.json" {
		t.Errorf("ListFiles() = %v, want [test.json]", files)
	}
	
	// Delete file
	store.DeleteFile("test.json")
	if store.FileExists("test.json") {
		t.Error("file should be deleted")
	}
}

func TestCacheBehavior(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "data_cache_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	store, _ := NewStore(&StoreConfig{
		BaseDir:  tmpDir,
		CacheTTL: 1 * time.Second,
	})
	
	testData := map[string]string{"cached": "value"}
	store.SaveJSON("cached.json", testData)
	
	// First load (cache miss)
	store.LoadJSON("cached.json")
	
	stats := store.CacheStats()
	if stats.Hits != 0 || stats.Misses != 1 {
		t.Errorf("Stats after first load: hits=%d misses=%d, want hits=0 misses=1", stats.Hits, stats.Misses)
	}
	
	// Second load (cache hit)
	store.LoadJSON("cached.json")
	
	stats = store.CacheStats()
	if stats.Hits != 1 {
		t.Errorf("Cache hits = %d, want 1", stats.Hits)
	}
	
	// Wait for cache to expire
	time.Sleep(1100 * time.Millisecond)
	
	// Third load (cache miss due to TTL)
	store.LoadJSON("cached.json")
	
	stats = store.CacheStats()
	if stats.Misses != 2 {
		t.Errorf("Cache misses = %d, want 2", stats.Misses)
	}
}

func TestEmbeddedData(t *testing.T) {
	
	tmpDir, err := os.MkdirTemp("", "data_embed_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	t.Skip("Skipping embedded test - requires actual embedded files")
}

// Benchmarks
func BenchmarkSaveJSON(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "data_bench")
	defer os.RemoveAll(tmpDir)
	
	store, _ := NewStore(&StoreConfig{BaseDir: tmpDir, CacheTTL: 0})
	
	data := map[string]interface{}{
		"player": map[string]interface{}{
			"name": "BenchPlayer",
			"id":   12345,
		},
	}
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		store.SaveJSON("bench.json", data)
	}
}

func BenchmarkLoadJSON_WithCache(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "data_bench")
	defer os.RemoveAll(tmpDir)
	
	store, _ := NewStore(&StoreConfig{BaseDir: tmpDir, CacheTTL: time.Hour})
	
	data := map[string]string{"key": "value"}
	store.SaveJSON("bench.json", data)
	
	// Prime the cache
	store.LoadJSON("bench.json")
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		store.LoadJSON("bench.json")
	}
}

func BenchmarkLoadJSON_NoCache(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "data_bench")
	defer os.RemoveAll(tmpDir)
	
	store, _ := NewStore(&StoreConfig{BaseDir: tmpDir, CacheTTL: 0})
	
	data := map[string]string{"key": "value"}
	store.SaveJSON("bench.json", data)
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		store.LoadJSON("bench.json")
	}
}