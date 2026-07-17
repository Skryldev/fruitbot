package responses

import (
	"encoding/json"
	"net/http"
	"time"
)

type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Retry   bool   `json:"retry"`
}

type ErrorResponse struct {
	HTTPStatus int
	APIError   APIError
}

// ============================================================
// Pre-defined Errors
// ============================================================

var (
	ErrBadRequest = ErrorResponse{
		HTTPStatus: http.StatusBadRequest,
		APIError: APIError{
			Code:    "BAD_REQUEST",
			Message: "Invalid request parameters",
			Retry:   false,
		},
	}

	ErrUnauthorized = ErrorResponse{
		HTTPStatus: http.StatusUnauthorized,
		APIError: APIError{
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
			Retry:   false,
		},
	}

	ErrForbidden = ErrorResponse{
		HTTPStatus: http.StatusForbidden,
		APIError: APIError{
			Code:    "FORBIDDEN",
			Message: "Access denied",
			Retry:   false,
		},
	}

	ErrNotFound = ErrorResponse{
		HTTPStatus: http.StatusNotFound,
		APIError: APIError{
			Code:    "NOT_FOUND",
			Message: "Resource not found",
			Retry:   false,
		},
	}

	ErrInternalServer = ErrorResponse{
		HTTPStatus: http.StatusInternalServerError,
		APIError: APIError{
			Code:    "INTERNAL_ERROR",
			Message: "Internal server error",
			Retry:   true,
		},
	}

	ErrServiceUnavailable = ErrorResponse{
		HTTPStatus: http.StatusServiceUnavailable,
		APIError: APIError{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Service temporarily unavailable",
			Retry:   true,
		},
	}

	ErrTooManyRequests = ErrorResponse{
		HTTPStatus: http.StatusTooManyRequests,
		APIError: APIError{
			Code:    "RATE_LIMIT",
			Message: "Too many requests, please slow down",
			Retry:   true,
		},
	}

	ErrGameServerError = ErrorResponse{
		HTTPStatus: http.StatusBadGateway,
		APIError: APIError{
			Code:    "GAME_SERVER_ERROR",
			Message: "Game server returned an error",
			Retry:   true,
		},
	}

	ErrAccountOnline = ErrorResponse{
		HTTPStatus: http.StatusConflict,
		APIError: APIError{
			Code:    "ACCOUNT_ONLINE",
			Message: "Account is currently online on another device",
			Retry:   true,
		},
	}

	ErrAccountBlocked = ErrorResponse{
		HTTPStatus: http.StatusForbidden,
		APIError: APIError{
			Code:    "ACCOUNT_BLOCKED",
			Message: "Your account is blocked",
			Retry:   false,
		},
	}

	ErrSessionExpired = ErrorResponse{
		HTTPStatus: http.StatusUnauthorized,
		APIError: APIError{
			Code:    "SESSION_EXPIRED",
			Message: "Session expired, please reload player",
			Retry:   false,
		},
	}

	ErrNotImplemented = ErrorResponse{
		HTTPStatus: http.StatusNotImplemented,
		APIError: APIError{
			Code:    "NOT_IMPLEMENTED",
			Message: "This feature is not yet implemented",
			Retry:   false,
		},
	}
)

// ============================================================
// Error Detection Helper
// ============================================================

func DetectGameError(err error) ErrorResponse {
	if err == nil {
		return ErrorResponse{}
	}

	errStr := err.Error()

	switch {
	case contains(errStr, "[184]") || contains(errStr, "online on another device"):
		return ErrAccountOnline

	case contains(errStr, "[101]") || contains(errStr, "account is blocked"):
		return ErrAccountBlocked

	case contains(errStr, "[116]") || contains(errStr, "Access denied"):
		return ErrForbidden

	case contains(errStr, "[429]") || contains(errStr, "too many requests"):
		return ErrTooManyRequests

	case contains(errStr, "[0]") || contains(errStr, "unknown error"):
		return ErrGameServerError

	case contains(errStr, "circuit breaker is open"):
		return ErrServiceUnavailable

	case contains(errStr, "empty response") || contains(errStr, "cannot decrypt"):
		return ErrGameServerError

	case contains(errStr, "request timeout") || contains(errStr, "i/o timeout"):
		return ErrServiceUnavailable

	default:
		return ErrInternalServer
	}
}

// ============================================================
// Response Helpers
// ============================================================

func Success(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

func Created(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusCreated, APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

func Error(w http.ResponseWriter, err error) {
	errResp := DetectGameError(err)
	writeJSON(w, errResp.HTTPStatus, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    errResp.APIError.Code,
			Message: errResp.APIError.Message,
			Details: err.Error(),
			Retry:   errResp.APIError.Retry,
		},
		Timestamp: time.Now().Unix(),
	})
}

func ErrorWithStatus(w http.ResponseWriter, status int, code, message string, retry bool) {
	writeJSON(w, status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Retry:   retry,
		},
		Timestamp: time.Now().Unix(),
	})
}

func BadRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "BAD_REQUEST",
			Message: message,
			Retry:   false,
		},
		Timestamp: time.Now().Unix(),
	})
}

func NotFound(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusNotFound, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "NOT_FOUND",
			Message: message,
			Retry:   false,
		},
		Timestamp: time.Now().Unix(),
	})
}

func Unauthorized(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, APIResponse{
		Success: false,
		Error:   &ErrUnauthorized.APIError,
		Timestamp: time.Now().Unix(),
	})
}

func NotImplemented(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotImplemented, APIResponse{
		Success: false,
		Error:   &ErrNotImplemented.APIError,
		Timestamp: time.Now().Unix(),
	})
}

// ============================================================
// Internal Helpers
// ============================================================

func JSON(w http.ResponseWriter, status int, data interface{}) {
	writeJSON(w, status, data)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============================================================
// HTTP Constants
// ============================================================

const (
	StatusOK                  = http.StatusOK
	StatusCreated             = http.StatusCreated
	StatusBadRequest          = http.StatusBadRequest
	StatusUnauthorized        = http.StatusUnauthorized
	StatusForbidden           = http.StatusForbidden
	StatusNotFound            = http.StatusNotFound
	StatusConflict            = http.StatusConflict
	StatusTooManyRequests     = http.StatusTooManyRequests
	StatusInternalServerError = http.StatusInternalServerError
	StatusBadGateway          = http.StatusBadGateway
	StatusServiceUnavailable  = http.StatusServiceUnavailable
)