package config

import (
	"testing"
	"time"
	
	domainErrors "fruitbot/internal/domain/errors"
)

func TestErrorMapper_MapError(t *testing.T) {
	mapper := NewErrorMapper()
	
	tests := []struct {
		name     string
		code     domainErrors.ErrorCode
		expected *domainErrors.DomainError
	}{
		{
			name:     "known error",
			code:     domainErrors.CodeAccessDenied,
			expected: domainErrors.ErrAccessDenied,
		},
		{
			name:     "unknown error",
			code:     domainErrors.ErrorCode(9999),
			expected: nil,
		},
		{
			name:     "general error",
			code:     domainErrors.CodeGeneralError,
			expected: domainErrors.ErrGeneral,
		},
		{
			name:     "too many requests",
			code:     domainErrors.CodeTooManyRequests,
			expected: domainErrors.ErrTooManyRequests,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := mapper.MapError(tt.code)
			if tt.expected == nil {
				if ok {
					t.Error("expected no mapping, but got one")
				}
				return
			}
			if !ok {
				t.Error("expected mapping, but got none")
				return
			}
			if got != tt.expected {
				t.Errorf("MapError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorMapper_MustMapError(t *testing.T) {
	mapper := NewErrorMapper()
	
	// Known error
	if got := mapper.MustMapError(domainErrors.CodeAccessDenied); got != domainErrors.ErrAccessDenied {
		t.Errorf("MustMapError() = %v, want %v", got, domainErrors.ErrAccessDenied)
	}
	
	// Unknown error should return ErrUnknown
	if got := mapper.MustMapError(9999); got != domainErrors.ErrUnknown {
		t.Errorf("MustMapError() for unknown = %v, want %v", got, domainErrors.ErrUnknown)
	}
}

func TestDeviceFingerprinter_GetRandomModel(t *testing.T) {
	df := NewDeviceFingerprinter()
	
	// Get multiple models and verify they're from the list
	models := df.GetAllModels()
	modelSet := make(map[string]bool, len(models))
	for _, m := range models {
		modelSet[m] = true
	}
	
	// Test 100 random selections
	for i := 0; i < 100; i++ {
		model := df.GetRandomModel()
		if !modelSet[model] {
			t.Errorf("GetRandomModel() = %q, not in known models", model)
		}
	}
}

func TestDeviceFingerprinter_AddModel(t *testing.T) {
	df := NewDeviceFingerprinter()
	initialCount := len(df.GetAllModels())
	
	newModel := "iPhone 15 Pro Max"
	df.AddModel(newModel)
	
	models := df.GetAllModels()
	if len(models) != initialCount+1 {
		t.Errorf("expected %d models, got %d", initialCount+1, len(models))
	}
	
	found := false
	for _, m := range models {
		if m == newModel {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("added model %q not found in list", newModel)
	}
}

func TestConfig_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.ServerAddr == "" {
		t.Error("server address should not be empty")
	}
	if cfg.ConnectTimeout <= 0 {
		t.Error("connect timeout should be positive")
	}
	if cfg.MaxRetries < 0 {
		t.Error("max retries should not be negative")
	}
	if cfg.ErrorMapper == nil {
		t.Error("error mapper should not be nil")
	}
	if cfg.DeviceFingerprinter == nil {
		t.Error("device fingerprinter should not be nil")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name: "missing server address",
			cfg: &Config{
				ConnectTimeout: 10 * time.Second,
				MaxRetries:     3,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			cfg: &Config{
				ServerAddr:     "localhost:8080",
				ConnectTimeout: -1 * time.Second,
				MaxRetries:     3,
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_FunctionalOptions(t *testing.T) {
	cfg := NewConfig(
		WithServerAddr("test.example.com:8080"),
		WithTimeouts(5*time.Second, 15*time.Second, 15*time.Second),
		WithRetryConfig(5, 2*time.Second),
	)
	
	if cfg.ServerAddr != "test.example.com:8080" {
		t.Errorf("ServerAddr = %v, want test.example.com:8080", cfg.ServerAddr)
	}
	if cfg.ConnectTimeout != 5*time.Second {
		t.Errorf("ConnectTimeout = %v, want 5s", cfg.ConnectTimeout)
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("MaxRetries = %v, want 5", cfg.MaxRetries)
	}
}

// Benchmarks
func BenchmarkErrorMapper_MapError(b *testing.B) {
	mapper := NewErrorMapper()
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		mapper.MapError(domainErrors.CodeAccessDenied)
	}
}

func BenchmarkErrorMapper_MustMapError(b *testing.B) {
	mapper := NewErrorMapper()
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		mapper.MustMapError(domainErrors.CodeAccessDenied)
	}
}

func BenchmarkDeviceFingerprinter_GetRandomModel(b *testing.B) {
	df := NewDeviceFingerprinter()
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		df.GetRandomModel()
	}
}