// internal/infrastructure/crypto/crypto_test.go
package crypto

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestNewEncryption(t *testing.T) {
	tests := []struct {
		name       string
		opts       []Option
		wantKey    []byte
		wantDouble bool
		wantVer    Version
	}{
		{
			name:    "default",
			opts:    nil,
			wantKey: defaultKey,
		},
		{
			name:    "socket mode",
			opts:    []Option{WithSocketMode()},
			wantKey: socketKey,
			wantDouble: true,
		},
		{
			name:    "version 2",
			opts:    []Option{WithVersion(Version2)},
			wantKey: version2Key,
			wantVer: Version2,
		},
		{
			name:    "custom key",
			opts:    []Option{WithKey([]byte("custom_key_here"))},
			wantKey: []byte("custom_key_here"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEncryption(tt.opts...)
			if !bytes.Equal(e.key, tt.wantKey) {
				t.Errorf("key = %s, want %s", e.key, tt.wantKey)
			}
			if e.requiresDoubleEnc != tt.wantDouble {
				t.Errorf("requiresDoubleEnc = %v, want %v", e.requiresDoubleEnc, tt.wantDouble)
			}
			if tt.wantVer != 0 && e.version != tt.wantVer {
				t.Errorf("version = %d, want %d", e.version, tt.wantVer)
			}
		})
	}
}

func TestEncryption_EncryptDecrypt(t *testing.T) {
	tests := []struct {
		name    string
		enc     *Encryption
		message string
	}{
		{
			name:    "default encrypt/decrypt",
			enc:     DefaultEncryption,
			message: "Hello, World!",
		},
		{
			name:    "socket mode encrypt/decrypt",
			enc:     SocketEncryption,
			message: "Socket message test",
		},
		{
			name:    "version 2 encrypt/decrypt",
			enc:     Version2Encryption,
			message: "Version 2 test",
		},
		{
			name:    "empty message",
			enc:     DefaultEncryption,
			message: "",
		},
		{
			name:    "unicode message",
			enc:     DefaultEncryption,
			message: "سلام دنیا! 🌍",
		},
		{
			name:    "long message",
			enc:     DefaultEncryption,
			message: strings.Repeat("Long message test. ", 100),
		},
		{
			name:    "JSON-like message",
			enc:     DefaultEncryption,
			message: `{"key": "value", "number": 42}`,
		},
		{
			name:    "special characters",
			enc:     DefaultEncryption,
			message: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.message == "" {
				_, err := tt.enc.Encrypt(tt.message)
				if err == nil {
					t.Error("expected error for empty message")
				}
				return
			}

			encrypted, err := tt.enc.Encrypt(tt.message)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			if encrypted == tt.message {
				t.Error("encrypted message should not equal original")
			}

			decrypted, err := tt.enc.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tt.message {
				t.Errorf("Decrypt() = %q, want %q", decrypted, tt.message)
			}
		})
	}
}

func TestEncryption_DecryptSpecialCases(t *testing.T) {
	tests := []struct {
		name      string
		encrypted string
		want      string
		wantErr   bool
	}{
		{
			name:      "JSON pass-through",
			encrypted: `{"result": "ok"}`,
			want:      `{"result": "ok"}`,
			wantErr:   false,
		},
		{
			name:      "empty string",
			encrypted: "",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "invalid base64",
			encrypted: "!!!invalid!!!",
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultEncryption.Decrypt(tt.encrypted)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Decrypt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFastXOR(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		key  []byte
	}{
		{
			name: "simple XOR",
			data: []byte("Hello"),
			key:  []byte("key"),
		},
		{
			name: "single byte key",
			data: []byte("Test data"),
			key:  []byte{0x42},
		},
		{
			name: "empty key",
			data: []byte("Test"),
			key:  []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := make([]byte, len(tt.data))
			copy(original, tt.data)

			// XOR twice should return original
			FastXOR(tt.data, tt.key)
			FastXOR(tt.data, tt.key)

			if !bytes.Equal(tt.data, original) {
				t.Errorf("double XOR should return original: got %v, want %v", tt.data, original)
			}
		})
	}
}

func TestEncryption_Stats(t *testing.T) {
	e := NewEncryption()
	
	// Encrypt some messages
	for i := 0; i < 10; i++ {
		e.Encrypt(fmt.Sprintf("message %d", i))
	}
	
	// Decrypt with some errors
	for i := 0; i < 5; i++ {
		e.Decrypt("invalid base64!!!")
	}
	
	encCount, decCount, decErrors := e.Stats()
	
	if encCount != 10 {
		t.Errorf("encrypt count = %d, want 10", encCount)
	}
	if decCount != 5 {
		t.Errorf("decrypt count = %d, want 5", decCount)
	}
	if decErrors != 5 {
		t.Errorf("decrypt errors = %d, want 5", decErrors)
	}
	
	e.ResetStats()
	encCount, decCount, decErrors = e.Stats()
	if encCount != 0 || decCount != 0 || decErrors != 0 {
		t.Error("stats should be reset to zero")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	message := "test convenience functions"
	
	encrypted, err := Encrypt(message)
	if err != nil {
		t.Fatal(err)
	}
	
	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	
	if decrypted != message {
		t.Errorf("Decrypt() = %q, want %q", decrypted, message)
	}
}

// Benchmarks
func BenchmarkDefaultEncrypt(b *testing.B) {
	e := DefaultEncryption
	message := "Hello, World! This is a test message for benchmarking."
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		e.Encrypt(message)
	}
}

func BenchmarkDefaultDecrypt(b *testing.B) {
	e := DefaultEncryption
	encrypted, _ := e.Encrypt("Hello, World! This is a test message for benchmarking.")
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		e.Decrypt(encrypted)
	}
}

func BenchmarkSocketEncrypt(b *testing.B) {
	e := SocketEncryption
	message := "Socket encryption test message"
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		e.Encrypt(message)
	}
}

func BenchmarkFastXOR(b *testing.B) {
	data := []byte(strings.Repeat("Benchmark test data ", 100))
	key := []byte("test_key")
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		FastXOR(data, key)
	}
}

func BenchmarkFastXOR_SingleByteKey(b *testing.B) {
	data := []byte(strings.Repeat("Benchmark test data ", 100))
	key := []byte{0x42}
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		FastXOR(data, key)
	}
}