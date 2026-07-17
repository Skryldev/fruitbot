package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DomainError
		expected string
	}{
		{
			name:     "without params",
			err:      ErrAccessDenied,
			expected: "[116] Access denied.",
		},
		{
			name:     "with params",
			err:      ErrNotEnoughGold.WithParams("100"),
			expected: "[183] You need 100 more gold.",
		},
		{
			name:     "with multiple params",
			err:      ErrNameTooLong.WithParams("50"),
			expected: "[180] Name is too long. You can specify 50 characters for your name.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDomainError_Is(t *testing.T) {
	err := ErrCardNotFound.WithParams("123")
	
	if !errors.Is(err, ErrCardNotFound) {
		t.Error("errors.Is should return true for same error code")
	}
	
	if errors.Is(err, ErrAccessDenied) {
		t.Error("errors.Is should return false for different error code")
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("original error")
	err := ErrInternal.Wrap(cause)
	
	if !errors.Is(err, cause) {
		t.Error("Unwrap should return the cause")
	}
}

func TestIsCode(t *testing.T) {
	err := ErrAuctionClosed
	
	if !IsCode(err, CodeAuctionClosed) {
		t.Error("IsCode should return true for matching code")
	}
	
	if IsCode(err, CodeAccessDenied) {
		t.Error("IsCode should return false for non-matching code")
	}
}

func TestGetCode(t *testing.T) {
	err := ErrTribeFull
	if code := GetCode(err); code != CodeTribeFull {
		t.Errorf("GetCode() = %d, want %d", code, CodeTribeFull)
	}
	
	// Test with standard error
	standardErr := errors.New("standard error")
	if code := GetCode(standardErr); code != CodeUnknown {
		t.Errorf("GetCode() for standard error = %d, want %d", code, CodeUnknown)
	}
}

func TestNew(t *testing.T) {
	// Test creating error with params
	err := New(CodeNotEnoughGold, "500")
	expected := "[183] You need 500 more gold."
	if err.Error() != expected {
		t.Errorf("New() = %v, want %v", err.Error(), expected)
	}
	
	// Test creating unknown error code
	unknownErr := New(9999)
	if unknownErr.Code != 9999 {
		t.Errorf("New() code = %d, want %d", unknownErr.Code, 9999)
	}
}

func TestWrap(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := Wrap(CodeOperationTimeout, cause)
	
	if !IsCode(err, CodeOperationTimeout) {
		t.Error("Wrap should preserve error code")
	}
	
	if !errors.Is(err, cause) {
		t.Error("Wrap should preserve cause")
	}
}

// Benchmark tests
func BenchmarkErrorCreation(b *testing.B) {
	b.Run("WithoutParams", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ErrAccessDenied
		}
	})
	
	b.Run("WithParams", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = ErrNotEnoughGold.WithParams("100")
		}
	})
}

func BenchmarkErrorCheck(b *testing.B) {
	err := ErrCardNotFound.WithParams("123")
	
	b.Run("errors.Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			errors.Is(err, ErrCardNotFound)
		}
	})
	
	b.Run("IsCode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			IsCode(err, CodeCardNotFound)
		}
	})
}

func BenchmarkErrorFormatting(b *testing.B) {
	err := ErrNotEnoughGold.WithParams("100")
	
	b.Run("Error", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = err.Error()
		}
	})
}