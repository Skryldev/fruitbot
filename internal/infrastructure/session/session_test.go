package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileStorage_SaveLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "session_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	fs, err := NewFileStorage(&FileStorageConfig{
		Directory: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	
	ctx := context.Background()
	sessionName := "test_session"
	
	// Test save
	data := &SessionData{
		RestoreKey:  "test_restore_key",
		Passport:    "test_passport",
		UDID:        "test_udid",
		MobileModel: "iPhone 15",
		Player: &PlayerInfo{
			ID:        12345,
			Name:      "TestPlayer",
			InviteKey: "INVITE123",
		},
	}
	
	if err := fs.Save(ctx, sessionName, data); err != nil {
		t.Fatal(err)
	}
	
	// Verify file exists
	exists, err := fs.Exists(ctx, sessionName)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("session file should exist")
	}
	
	// Test load
	loaded, err := fs.Load(ctx, sessionName)
	if err != nil {
		t.Fatal(err)
	}
	
	if loaded.Passport != data.Passport {
		t.Errorf("Passport = %s, want %s", loaded.Passport, data.Passport)
	}
	if loaded.Player.ID != data.Player.ID {
		t.Errorf("Player ID = %d, want %d", loaded.Player.ID, data.Player.ID)
	}
}

func TestFileStorage_Encryption(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "session_enc_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create 32-byte test key
	secretKey := make([]byte, 32)
	for i := range secretKey {
		secretKey[i] = byte(i)
	}
	
	fs, err := NewFileStorage(&FileStorageConfig{
		Directory: tmpDir,
		SecretKey: secretKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	
	ctx := context.Background()
	
	data := &SessionData{
		Passport: "encrypted_test",
		Player:   &PlayerInfo{ID: 1, Name: "Test"},
	}
	
	if err := fs.Save(ctx, "enc_test", data); err != nil {
		t.Fatal(err)
	}
	
	// Verify file content is encrypted (not plain JSON)
	fileContent, _ := os.ReadFile(filepath.Join(tmpDir, "enc_test.fb"))
	if string(fileContent[:1]) == "{" {
		t.Error("file should be encrypted (not plain JSON)")
	}
	
	// Load and verify decryption
	loaded, err := fs.Load(ctx, "enc_test")
	if err != nil {
		t.Fatal(err)
	}
	
	if loaded.Passport != data.Passport {
		t.Error("decrypted data doesn't match original")
	}
}

func TestMemoryStorage(t *testing.T) {
	ms := NewMemoryStorage()
	ctx := context.Background()
	
	data := &SessionData{
		Passport: "memory_test",
		Player:   &PlayerInfo{ID: 1, Name: "MemPlayer"},
	}
	
	if err := ms.Save(ctx, "mem_session", data); err != nil {
		t.Fatal(err)
	}
	
	exists, _ := ms.Exists(ctx, "mem_session")
	if !exists {
		t.Error("session should exist in memory")
	}
	
	loaded, _ := ms.Load(ctx, "mem_session")
	if loaded.Passport != data.Passport {
		t.Error("loaded data doesn't match")
	}
	
	// Test delete
	ms.Delete(ctx, "mem_session")
	exists, _ = ms.Exists(ctx, "mem_session")
	if exists {
		t.Error("session should be deleted")
	}
}

func TestSessionManager(t *testing.T) {
	ms := NewMemoryStorage()
	
	sm, err := NewSessionManager(&SessionConfig{
		Storage:     ms,
		SessionName: "test_player",
	})
	if err != nil {
		t.Fatal(err)
	}
	
	ctx := context.Background()
	
	// Check session doesn't exist
	exists, _ := sm.DoesSessionExist(ctx)
	if exists {
		t.Error("session should not exist yet")
	}
	
	// Save session
	err = sm.SaveSession(ctx, &PlayerInfo{
		ID:        12345,
		Name:      "TestPlayer",
		InviteKey: "INVITE123",
	})
	if err != nil {
		t.Fatal(err)
	}
	
	// Check session exists
	exists, _ = sm.DoesSessionExist(ctx)
	if !exists {
		t.Error("session should exist now")
	}
	
	// Load session
	data, err := sm.LoadSessionData(ctx)
	if err != nil {
		t.Fatal(err)
	}
	
	if data.Player.ID != 12345 {
		t.Errorf("Player ID = %d, want 12345", data.Player.ID)
	}
	
	// List sessions
	names, _ := sm.ListSessions(ctx)
	if len(names) != 1 || names[0] != "test_player" {
		t.Errorf("ListSessions() = %v, want [test_player]", names)
	}
	
	// Delete session
	sm.DeleteSession(ctx)
	exists, _ = sm.DoesSessionExist(ctx)
	if exists {
		t.Error("session should be deleted")
	}
}

// Benchmarks
func BenchmarkFileStorage_Save(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "session_bench")
	defer os.RemoveAll(tmpDir)
	
	fs, _ := NewFileStorage(&FileStorageConfig{Directory: tmpDir})
	ctx := context.Background()
	
	data := &SessionData{
		Passport: "bench_test",
		Player:   &PlayerInfo{ID: 1, Name: "Bench"},
	}
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		fs.Save(ctx, "bench_session", data)
	}
}

func BenchmarkFileStorage_Load(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "session_bench")
	defer os.RemoveAll(tmpDir)
	
	fs, _ := NewFileStorage(&FileStorageConfig{Directory: tmpDir})
	ctx := context.Background()
	
	data := &SessionData{
		Passport: "bench_test",
		Player:   &PlayerInfo{ID: 1, Name: "Bench"},
	}
	fs.Save(ctx, "bench_session", data)
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		fs.Load(ctx, "bench_session")
	}
}

func BenchmarkMemoryStorage_Save(b *testing.B) {
	ms := NewMemoryStorage()
	ctx := context.Background()
	
	data := &SessionData{
		Passport: "bench_test",
		Player:   &PlayerInfo{ID: 1, Name: "Bench"},
	}
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		ms.Save(ctx, "bench_session", data)
	}
}