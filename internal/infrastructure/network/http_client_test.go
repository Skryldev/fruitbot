package network

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainErrors "fruitbot/internal/domain/errors"
)

func TestNewHTTPClient(t *testing.T) {
	cfg := DefaultClientConfig("test_passport")
	
	client, err := NewHTTPClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	
	if client == nil {
		t.Fatal("client should not be nil")
	}
	
	client.Close()
}

func TestHTTPClient_SendRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != defaultUserAgent {
			t.Errorf("User-Agent = %s, want %s", r.Header.Get("User-Agent"), defaultUserAgent)
		}
		
		resp := APIResponse{
			Status: true,
			Data:   map[string]string{"result": "ok"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	cfg := DefaultClientConfig("test_passport")
	cfg.BaseURL = server.URL
	
	client, err := NewHTTPClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	
	ctx := context.Background()
	result, err := client.Post(ctx, "test", map[string]interface{}{
		"action": "test",
	})
	
	if err != nil {
		t.Fatalf("SendRequest failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("result should not be nil")
	}
}

func TestHTTPClient_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()
	
	cfg := DefaultClientConfig("test_passport")
	cfg.BaseURL = server.URL
	
	client, _ := NewHTTPClient(cfg)
	defer client.Close()
	
	ctx := context.Background()
	_, err := client.Post(ctx, "test", nil)
	
	if !domainErrors.Is(err, domainErrors.ErrTooManyRequests) {
		t.Errorf("expected ErrTooManyRequests, got %v", err)
	}
	
	stats := client.Stats()
	if stats.RateLimitCount != 1 {
		t.Errorf("RateLimitCount = %d, want 1", stats.RateLimitCount)
	}
}

func TestHTTPClient_RetryLogic(t *testing.T) {
	attempts := 0
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			http.Error(w, "temporary error", http.StatusInternalServerError)
			return
		}
		
		resp := APIResponse{
			Status: true,
			Data:   map[string]string{"result": "ok"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	cfg := DefaultClientConfig("test_passport")
	cfg.BaseURL = server.URL
	cfg.MaxRetries = 5
	
	client, _ := NewHTTPClient(cfg)
	defer client.Close()
	
	ctx := context.Background()
	_, err := client.Post(ctx, "test", nil)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
	
	stats := client.Stats()
	if stats.RetryCount != 2 {
		t.Errorf("RetryCount = %d, want 2", stats.RetryCount)
	}
}

func TestHTTPClient_ImageResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte{0x89, 0x50, 0x4E, 0x47})
	}))
	defer server.Close()
	
	cfg := DefaultClientConfig("test_passport")
	cfg.BaseURL = server.URL
	
	client, _ := NewHTTPClient(cfg)
	defer client.Close()
	
	ctx := context.Background()
	result, err := client.Get(ctx, "image.png", nil)
	
	if err != nil {
		t.Fatal(err)
	}
	
	imageData, ok := result.([]byte)
	if !ok {
		t.Fatal("result should be []byte")
	}
	
	if len(imageData) != 4 {
		t.Errorf("image data length = %d, want 4", len(imageData))
	}
}

func TestHTTPClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()
	
	cfg := DefaultClientConfig("test_passport")
	cfg.BaseURL = server.URL
	cfg.Timeout = 100 * time.Millisecond
	
	client, _ := NewHTTPClient(cfg)
	defer client.Close()
	
	ctx := context.Background()
	_, err := client.Post(ctx, "test", nil)
	
	if err == nil {
		t.Error("expected timeout error")
	}
	
	stats := client.Stats()
	if stats.TimeoutCount != 1 {
		t.Errorf("TimeoutCount = %d, want 1", stats.TimeoutCount)
	}
}

func BenchmarkHTTPClient_SendRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			Status: true,
			Data:   map[string]string{"result": "ok"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	
	cfg := DefaultClientConfig("test_passport")
	cfg.BaseURL = server.URL
	
	client, _ := NewHTTPClient(cfg)
	defer client.Close()
	
	ctx := context.Background()
	input := map[string]interface{}{"action": "benchmark"}
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		client.Post(ctx, "test", input)
	}
}