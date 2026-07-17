package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fruitbot/internal/infrastructure/crypto"
	domainErrors "fruitbot/internal/domain/errors"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

const (
	defaultBaseURL   = "https://iran.fruitcraft.ir"
	defaultTimeout   = 10 * time.Second
	defaultUserAgent = "Dalvik/2.1.0"
	maxRetries       = 5
	retryDelay       = 500 * time.Millisecond

	cookieFruitPassport = "FRUITPASSPORT"
)

type ClientConfig struct {
	BaseURL             string
	Timeout             time.Duration
	MaxRetries          int
	EncVersion          crypto.Version
	Passport            string
	Logger              *zap.Logger
	CBOptions           *gobreaker.Settings
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
	EnableHTTP2         bool
}

func DefaultClientConfig(passport string) *ClientConfig {
	return &ClientConfig{
		BaseURL:             defaultBaseURL,
		Timeout:             defaultTimeout,
		MaxRetries:          maxRetries,
		EncVersion:          crypto.Version2,
		Passport:            passport,
		Logger:              zap.NewNop(),
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		EnableHTTP2:         true,
	}
}

type HTTPClient struct {
	cfg            *ClientConfig
	http           *http.Client
	crypto         *crypto.Encryption
	cb             *gobreaker.CircuitBreaker
	logger         *zap.Logger
	requestCount   uint64
	successCount   uint64
	failureCount   uint64
	retryCount     uint64
	timeoutCount   uint64
	rateLimitCount uint64
	bufferPool     sync.Pool
	baseHeaders    http.Header
}

func NewHTTPClient(cfg *ClientConfig) (*HTTPClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	if cfg.Passport == "" {
		return nil, fmt.Errorf("passport is required")
	}

	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = maxRetries
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}

	transport := &http.Transport{
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		DisableCompression:    false,
		DisableKeepAlives:     false,
		ForceAttemptHTTP2:     cfg.EnableHTTP2,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
	}

	if cfg.EnableHTTP2 {
		if err := http2.ConfigureTransport(transport); err != nil {
			cfg.Logger.Warn("Failed to configure HTTP/2", zap.Error(err))
		}
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	var cb *gobreaker.CircuitBreaker
	if cfg.CBOptions != nil {
		cb = gobreaker.NewCircuitBreaker(*cfg.CBOptions)
	} else {
		cbSettings := gobreaker.Settings{
			Name:        "FruitCraftAPI",
			MaxRequests: 5,         
			Interval:    30 * time.Second,  
			Timeout:     15 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.Requests >= 20 && failureRatio >= 0.8  // ← سخت‌گیرانه‌تر
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				cfg.Logger.Warn("Circuit breaker state changed",
					zap.String("name", name),
					zap.String("from", from.String()),
					zap.String("to", to.String()),
				)
			},
		}
		cb = gobreaker.NewCircuitBreaker(cbSettings)
	}

	enc := crypto.NewEncryption(crypto.WithVersion(cfg.EncVersion))

	headers := make(http.Header)
	headers.Set("Accept-Encoding", "gzip")
	headers.Set("Connection", "Keep-Alive")
	headers.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	headers.Set("Host", "iran.fruitcraft.ir")
	headers.Set("User-Agent", defaultUserAgent)

	c := &HTTPClient{
		cfg:     cfg,
		http:    httpClient,
		crypto:  enc,
		cb:      cb,
		logger:  cfg.Logger,
		baseHeaders: headers,
		bufferPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}

	return c, nil
}

type APIResponse struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
}

type APIErrorResponse struct {
	Status bool `json:"status"`
	Data   struct {
		Code      int           `json:"code"`
		Arguments []interface{} `json:"arguments"`
	} `json:"data"`
}

type RequestOptions struct {
	Method      string
	Path        string
	Input       map[string]interface{}
	MaxAttempts int
	Headers     map[string]string
}

// ============================================================
// SMART RETRY LOGIC
// ============================================================

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	retryablePatterns := []string{
		"cannot decrypt empty string",    
		"illegal base64 data",            
		"circuit breaker is open",    
		"request timeout",            
		"connection refused",      
		"connection reset",       
		"EOF",                      
		"TLS handshake",              
		"no such host",                  
		"i/o timeout",                     
		"An unknown error occurred",  
		"You are currently online",      
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	nonRetryablePatterns := []string{
		"Account is blocked",
		"Access denied",
		"Invalid email",
		"Card not found",
	}

	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	if domainErrors.IsCode(err, domainErrors.CodeTooManyRequests) {
		return true
	}

	return true
}

func (c *HTTPClient) SendRequest(ctx context.Context, opts *RequestOptions) (interface{}, error) {
	if opts == nil {
		return nil, fmt.Errorf("request options are required")
	}

	atomic.AddUint64(&c.requestCount, 1)

	if opts.Method == "" {
		opts.Method = http.MethodPost
	}
	if opts.MaxAttempts == 0 {
		opts.MaxAttempts = c.cfg.MaxRetries
	}
	if opts.Input == nil {
		opts.Input = map[string]interface{}{"client": "iOS"}
	}

	url := fmt.Sprintf("%s/%s", c.cfg.BaseURL, opts.Path)

	var body io.Reader
	var req *http.Request
	var err error

	if opts.Method == http.MethodPost {
		jsonData, err := json.Marshal(opts.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		encrypted, err := c.crypto.Encrypt(string(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt request: %w", err)
		}

		bodyStr := fmt.Sprintf("edata=%s&version=%d", encrypted, c.cfg.EncVersion)
		body = strings.NewReader(bodyStr)
	}

	req, err = http.NewRequestWithContext(ctx, opts.Method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req, opts)

	if c.cb.State() == gobreaker.StateOpen {
		c.logger.Warn("Circuit breaker is open, waiting and retrying...")
		time.Sleep(retryDelay)
	}

	return c.executeWithSmartRetry(ctx, req, opts.MaxAttempts)
}

func (c *HTTPClient) executeWithSmartRetry(ctx context.Context, req *http.Request, maxAttempts int) (interface{}, error) {
	var lastErr error

	if maxAttempts < 1 {
		maxAttempts = 5
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result, err := c.cb.Execute(func() (interface{}, error) {
			return c.executeSingleRequest(req)
		})

		if err == nil {
			atomic.AddUint64(&c.successCount, 1)
			return result, nil
		}

		lastErr = err

		if !isRetryableError(err) {
			atomic.AddUint64(&c.failureCount, 1)
			return nil, err
		}

		if attempt == maxAttempts {
			atomic.AddUint64(&c.failureCount, 1)
			break
		}

		atomic.AddUint64(&c.retryCount, 1)

		baseBackoff := time.Duration(attempt*attempt) * 200 * time.Millisecond
		if baseBackoff > 5*time.Second {
			baseBackoff = 5 * time.Second
		}

		c.logger.Warn("Request failed, retrying...",
			zap.Int("attempt", attempt),
			zap.Int("max", maxAttempts),
			zap.Duration("backoff", baseBackoff),
			zap.String("error", err.Error()[:min(100, len(err.Error()))]),
		)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(baseBackoff):
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxAttempts, lastErr)
}

func (c *HTTPClient) executeSingleRequest(req *http.Request) (interface{}, error) {
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		req.ContentLength = int64(len(bodyBytes))
	}

	resp, err := c.http.Do(req)
	if err != nil {
		if isTimeout(err) {
			atomic.AddUint64(&c.timeoutCount, 1)
			return nil, fmt.Errorf("request timeout: %w", err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		atomic.AddUint64(&c.rateLimitCount, 1)
		return nil, domainErrors.ErrTooManyRequests
	}

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("server error: HTTP %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "image/") {
		imageData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read image response: %w", err)
		}
		return imageData, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	bodyStr := strings.TrimLeft(string(body), "\ufeff")

	if len(strings.TrimSpace(bodyStr)) == 0 {
		return nil, fmt.Errorf("empty response from server")
	}

	if strings.HasPrefix(bodyStr, "<!DOCTYPE") {
		return nil, domainErrors.ErrUnknown
	}

	decrypted, err := c.crypto.Decrypt(bodyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt response: %w", err)
	}

	if decrypted == "" {
		return nil, fmt.Errorf("cannot decrypt empty string")
	}

	var successResp APIResponse
	if err := json.Unmarshal([]byte(decrypted), &successResp); err == nil {
		if successResp.Status {
			return successResp.Data, nil
		}
	}

	var errorResp APIErrorResponse
	if err := json.Unmarshal([]byte(decrypted), &errorResp); err != nil {
		return decrypted, nil
	}

	errorCode := domainErrors.ErrorCode(errorResp.Data.Code)
	err = domainErrors.New(errorCode, errorResp.Data.Arguments...)

	if errorCode == domainErrors.CodeUnknown || errorCode == 0 {
		return nil, fmt.Errorf("[%d] %s", errorCode, err.Error())
	}

	return nil, err
}

func (c *HTTPClient) setHeaders(req *http.Request, opts *RequestOptions) {
	for key, values := range c.baseHeaders {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}

	req.Header.Set("Cookie", fmt.Sprintf("%s=%s;", cookieFruitPassport, c.cfg.Passport))

	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}
}

func (c *HTTPClient) Get(ctx context.Context, path string, input map[string]interface{}) (interface{}, error) {
	return c.SendRequest(ctx, &RequestOptions{
		Method: http.MethodGet,
		Path:   path,
		Input:  input,
	})
}

func (c *HTTPClient) Post(ctx context.Context, path string, input map[string]interface{}) (interface{}, error) {
	return c.SendRequest(ctx, &RequestOptions{
		Method: http.MethodPost,
		Path:   path,
		Input:  input,
	})
}

type HTTPStats struct {
	RequestCount   uint64
	SuccessCount   uint64
	FailureCount   uint64
	RetryCount     uint64
	TimeoutCount   uint64
	RateLimitCount uint64
	CBState        string
}

func (c *HTTPClient) Stats() HTTPStats {
	return HTTPStats{
		RequestCount:   atomic.LoadUint64(&c.requestCount),
		SuccessCount:   atomic.LoadUint64(&c.successCount),
		FailureCount:   atomic.LoadUint64(&c.failureCount),
		RetryCount:     atomic.LoadUint64(&c.retryCount),
		TimeoutCount:   atomic.LoadUint64(&c.timeoutCount),
		RateLimitCount: atomic.LoadUint64(&c.rateLimitCount),
		CBState:        c.cb.State().String(),
	}
}

func (c *HTTPClient) ResetStats() {
	atomic.StoreUint64(&c.requestCount, 0)
	atomic.StoreUint64(&c.successCount, 0)
	atomic.StoreUint64(&c.failureCount, 0)
	atomic.StoreUint64(&c.retryCount, 0)
	atomic.StoreUint64(&c.timeoutCount, 0)
	atomic.StoreUint64(&c.rateLimitCount, 0)
}

func (c *HTTPClient) Close() {
	c.http.CloseIdleConnections()
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	type timeout interface {
		Timeout() bool
	}
	t, ok := err.(timeout)
	return ok && t.Timeout()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}