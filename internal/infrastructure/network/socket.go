package network

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fruitbot/internal/infrastructure/crypto"
	domainErrors "fruitbot/internal/domain/errors"

	"go.uber.org/zap"
)


const (
	// TCP connection settings
	defaultServerAddr = "iranchat.fruitcraft.ir:1337"
	defaultBufferSize = 4096
	reconnectDelay    = 5 * time.Second
	maxReconnectAttempts = 10

	// Protocol markers
	startMarker = "__JSON__START__"
	endMarker   = "__JSON__END__"
	subMarker   = "__SUBSCRIBE__"
	endSubMarker = "__ENDSUBSCRIBE__"
	unsubMarker = "__UNSUBSCRIBE__"
	endUnsubMarker = "__ENDUNSUBSCRIBE__"

	// Channel prefixes
	userChannelPrefix  = "user"
	tribeChannelPrefix = "tribe"
)

const (
	MsgTypeChat         = "chat"
	MsgTypeTribeJoin    = "tribe_join"
	MsgTypeTribeKick    = "tribe_kick"
	MsgTypeTribeLeave   = "tribe_leave"
)

type Message map[string]interface{}

type Handler func(msg Message)

// ============================================================
// Socket Configuration
// ============================================================

type SocketConfig struct {
	ServerAddr          string
	ReconnectDelay      time.Duration
	MaxReconnectAttempts int
	BufferSize          int
	Logger              *zap.Logger
}

func DefaultSocketConfig() *SocketConfig {
	return &SocketConfig{
		ServerAddr:          defaultServerAddr,
		ReconnectDelay:      reconnectDelay,
		MaxReconnectAttempts: maxReconnectAttempts,
		BufferSize:          defaultBufferSize,
		Logger:              zap.NewNop(),
	}
}

// ============================================================
// Socket Client
// ============================================================

type Socket struct {
	cfg *SocketConfig
	
	// User information (atomic for lock-free reads)
	userID    atomic.Value 
	tribeID   atomic.Value 
	avatarID  atomic.Value 
	userName  atomic.Value 
	
	conn      net.Conn
	connected atomic.Bool
	startTime time.Time
	
	// Cryptography
	crypto *crypto.Encryption
	
	// Handler registry
	handlers   map[string]Handler
	handlersMu sync.RWMutex
	
	// Reconnection control
	shouldReconnect atomic.Bool
	reconnectCh     chan struct{}
	
	// Lifecycle management
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	
	// Buffered I/O
	reader    *bufio.Reader
	writer    *bufio.Writer
	
	// Message parsing buffer
	buffer    strings.Builder
	bufferMu  sync.Mutex
	
	// Goroutine pool for message handlers
	handlerPool chan struct{}
	
	// Metrics
	messagesSent     uint64
	messagesReceived uint64
	reconnectCount   uint64
	parseErrors      uint64
	
	// Close notification
	closeCh   chan struct{}
	closeOnce sync.Once
}

// NewSocket creates a new Socket client
func NewSocket(cfg *SocketConfig) *Socket {
	if cfg == nil {
		cfg = DefaultSocketConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	s := &Socket{
		cfg:          cfg,
		crypto:       crypto.NewEncryption(crypto.WithSocketMode()),
		handlers:     make(map[string]Handler, 10),
		reconnectCh:  make(chan struct{}, 1),
		handlerPool:  make(chan struct{}, 50), 
		closeCh:      make(chan struct{}),
		startTime:    time.Now(),
		ctx:          ctx,
		cancel:       cancel,
	}
	
	s.userID.Store(int64(0))
	s.tribeID.Store(int64(0))
	s.avatarID.Store(int64(1)) 
	s.userName.Store("")
	
	return s
}

func (s *Socket) Connect() error {
	s.closeExistingConnection()
	
	conn, err := net.DialTimeout("tcp", s.cfg.ServerAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", s.cfg.ServerAddr, err)
	}
	
	s.conn = conn
	s.reader = bufio.NewReaderSize(conn, s.cfg.BufferSize)
	s.writer = bufio.NewWriterSize(conn, s.cfg.BufferSize)
	s.connected.Store(true)
	
	if userID := s.UserID(); userID != 0 {
		s.subscribe(fmt.Sprintf("%s%d", userChannelPrefix, userID))
	}
	if tribeID := s.TribeID(); tribeID != 0 {
		s.subscribe(fmt.Sprintf("%s%d", tribeChannelPrefix, tribeID))
	}
	
	s.wg.Add(1)
	go s.receiveMessages()
	
	if s.shouldReconnect.Load() {
		s.wg.Add(1)
		go s.reconnectionLoop()
	}
	
	s.cfg.Logger.Info("Connected to server",
		zap.String("addr", s.cfg.ServerAddr),
		zap.Int64("user_id", s.UserID()),
	)
	
	return nil
}

func (s *Socket) Close(reconnect bool) {
	s.shouldReconnect.Store(reconnect)
	
	s.closeOnce.Do(func() {
		close(s.closeCh)
	})
	
	if !reconnect {
		s.connected.Store(false)
		s.cancel()
	}
	
	s.closeExistingConnection()
	s.wg.Wait()
}

func (s *Socket) closeExistingConnection() {
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	s.connected.Store(false)
}

// ============================================================
// Channel Subscription
// ============================================================

func (s *Socket) subscribe(channel string) error {
	if !s.IsConnected() {
		return fmt.Errorf("not connected")
	}
	
	encrypted, err := s.crypto.Encrypt(channel)
	if err != nil {
		return fmt.Errorf("failed to encrypt channel: %w", err)
	}
	
	message := fmt.Sprintf("%s%s%s", subMarker, encrypted, endSubMarker)
	return s.sendRaw(message)
}

func (s *Socket) unsubscribe(channel string) error {
	if !s.IsConnected() {
		return fmt.Errorf("not connected")
	}
	
	encrypted, err := s.crypto.Encrypt(channel)
	if err != nil {
		return fmt.Errorf("failed to encrypt channel: %w", err)
	}
	
	message := fmt.Sprintf("%s%s%s", unsubMarker, encrypted, endUnsubMarker)
	return s.sendRaw(message)
}

// ============================================================
// Data Sending
// ============================================================

func (s *Socket) sendRaw(data string) error {
	if !s.IsConnected() {
		return fmt.Errorf("not connected")
	}
	
	_, err := s.writer.WriteString(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}
	
	return s.writer.Flush()
}

func (s *Socket) sendJSON(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	encrypted, err := s.crypto.Encrypt(string(jsonData))
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}
	
	message := fmt.Sprintf("%s%s%s", startMarker, encrypted, endMarker)
	return s.sendRaw(message)
}

// ============================================================
// User Information
// ============================================================

func (s *Socket) SetInfo(userID, tribeID, avatarID int64, userName string) {
	if tribeID != 0 {
		s.tribeID.Store(tribeID)
		s.subscribe(fmt.Sprintf("%s%d", tribeChannelPrefix, tribeID))
	}
	
	if userID != 0 {
		s.userID.Store(userID)
		s.subscribe(fmt.Sprintf("%s%d", userChannelPrefix, userID))
	}
	
	if avatarID != 0 {
		s.avatarID.Store(avatarID)
	}
	
	if userName != "" {
		s.userName.Store(userName)
	}
}

func (s *Socket) UnsetInfo(tribe bool) {
	if tribe {
		s.tribeID.Store(int64(0))
		// Reconnect to clear subscriptions
		go func() {
			if err := s.Connect(); err != nil {
				s.cfg.Logger.Error("Failed to reconnect after unsetting tribe info", zap.Error(err))
			}
		}()
	}
}

func (s *Socket) UserID() int64 {
	if v := s.userID.Load(); v != nil {
		return v.(int64)
	}
	return 0
}

func (s *Socket) TribeID() int64 {
	if v := s.tribeID.Load(); v != nil {
		return v.(int64)
	}
	return 0
}

func (s *Socket) AvatarID() int64 {
	if v := s.avatarID.Load(); v != nil {
		return v.(int64)
	}
	return 1
}

func (s *Socket) UserName() string {
	if v := s.userName.Load(); v != nil {
		return v.(string)
	}
	return ""
}

func (s *Socket) IsConnected() bool {
	return s.connected.Load()
}

// ============================================================
// Tribe Messaging
// ============================================================

type TribeMessageData struct {
	PushMessageType string  `json:"push_message_type"`
	ID              int64   `json:"id"`
	Text            string  `json:"text"`
	AvatarID        int64   `json:"avatar_id"`
	CreationDate    int64   `json:"creationDate"`
	Channel         string  `json:"channel"`
	Timestamp       float64 `json:"timestamp"`
	Sender          string  `json:"sender"`
	MessageType     int     `json:"messageType"`
}

func (s *Socket) SendTribeMessage(text string) (*TribeMessageData, error) {
	tribeID := s.TribeID()
	if tribeID == 0 {
		return nil, domainErrors.ErrNotInTribe
	}
	
	now := time.Now()
	elapsed := now.Sub(s.startTime).Milliseconds()
	
	data := &TribeMessageData{
		PushMessageType: MsgTypeChat,
		ID:              now.Unix(),
		Text:            text,
		AvatarID:        s.AvatarID(),
		CreationDate:    now.Unix(),
		Channel:         fmt.Sprintf("%s%d", tribeChannelPrefix, tribeID),
		Timestamp:       float64(elapsed),
		Sender:          s.UserName(),
		MessageType:     1,
	}
	
	if err := s.sendJSON(data); err != nil {
		return nil, fmt.Errorf("failed to send tribe message: %w", err)
	}
	
	atomic.AddUint64(&s.messagesSent, 1)
	return data, nil
}

// ============================================================
// Message Handlers
// ============================================================

func (s *Socket) AddHandler(messageType string, handler Handler) {
	s.handlersMu.Lock()
	s.handlers[messageType] = handler
	s.handlersMu.Unlock()
}

func (s *Socket) RemoveHandler(messageType string) {
	s.handlersMu.Lock()
	delete(s.handlers, messageType)
	s.handlersMu.Unlock()
}

// ============================================================
// Message Receiving
// ============================================================

func (s *Socket) receiveMessages() {
	defer s.wg.Done()
	
	buffer := make([]byte, s.cfg.BufferSize)
	var partialMsg strings.Builder
	
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}
		
		n, err := s.reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				s.cfg.Logger.Info("Connection closed by server")
			} else {
				s.cfg.Logger.Error("Error reading from socket", zap.Error(err))
			}
			
			s.handleDisconnection()
			return
		}
		
		if n > 0 {
			partialMsg.Write(buffer[:n])
			s.processBuffer(&partialMsg)
		}
	}
}

func (s *Socket) processBuffer(buffer *strings.Builder) {
	rawData := buffer.String()
	buffer.Reset()
	
	for {
		endIdx := strings.Index(rawData, endMarker)
		if endIdx == -1 {
			buffer.WriteString(rawData)
			return
		}
		
		msgPart := rawData[:endIdx]
		rawData = rawData[endIdx+len(endMarker):]
		
		startIdx := strings.Index(msgPart, startMarker)
		if startIdx == -1 {
			continue
		}
		
		encrypted := msgPart[startIdx+len(startMarker):]
		
		decrypted, err := s.crypto.Decrypt(encrypted)
		if err != nil {
			atomic.AddUint64(&s.parseErrors, 1)
			s.cfg.Logger.Error("Failed to decrypt message", zap.Error(err))
			continue
		}
		
		var msg Message
		if err := json.Unmarshal([]byte(decrypted), &msg); err != nil {
			atomic.AddUint64(&s.parseErrors, 1)
			s.cfg.Logger.Error("Failed to parse message JSON", zap.Error(err))
			continue
		}
		
		atomic.AddUint64(&s.messagesReceived, 1)
		
		s.handleSpecialMessages(msg)
		
		s.dispatchMessage(msg)
	}
}

func (s *Socket) handleSpecialMessages(msg Message) {
	msgType, _ := msg["push_message_type"].(string)
	
	switch msgType {
	case MsgTypeTribeJoin:
		if tribe, ok := msg["tribe"].(map[string]interface{}); ok {
			if id, ok := tribe["id"].(float64); ok {
				tribeID := int64(id)
				s.tribeID.Store(tribeID)
				s.subscribe(fmt.Sprintf("%s%d", tribeChannelPrefix, tribeID))
			}
		}
		
	case MsgTypeTribeKick:
		s.UnsetInfo(true)
	}
}

func (s *Socket) dispatchMessage(msg Message) {
	msgType, _ := msg["push_message_type"].(string)
	if msgType == "" {
		return
	}
	
	s.handlersMu.RLock()
	handler, exists := s.handlers[msgType]
	s.handlersMu.RUnlock()
	
	if !exists {
		return
	}
	
	select {
	case s.handlerPool <- struct{}{}:
		s.wg.Add(1)
		go func() {
			defer func() {
				<-s.handlerPool
				s.wg.Done()
				
				if r := recover(); r != nil {
					s.cfg.Logger.Error("Handler panic recovered",
						zap.String("type", msgType),
						zap.Any("panic", r),
					)
				}
			}()
			handler(msg)
		}()
	default:
		s.cfg.Logger.Warn("Handler pool full, dropping message",
			zap.String("type", msgType),
		)
	}
}

// ============================================================
// Reconnection Logic
// ============================================================

func (s *Socket) handleDisconnection() {
	if s.shouldReconnect.Load() {
		select {
		case s.reconnectCh <- struct{}{}:
		default:
		}
	} else {
		s.connected.Store(false)
	}
}

func (s *Socket) reconnectionLoop() {
	defer s.wg.Done()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.reconnectCh:
			s.attemptReconnection()
		}
	}
}

func (s *Socket) attemptReconnection() {
	for attempt := 1; attempt <= s.cfg.MaxReconnectAttempts; attempt++ {
		if !s.shouldReconnect.Load() {
			return
		}
		
		s.cfg.Logger.Info("Attempting to reconnect",
			zap.Int("attempt", attempt),
			zap.Int("max", s.cfg.MaxReconnectAttempts),
		)
		
		if err := s.Connect(); err == nil {
			atomic.AddUint64(&s.reconnectCount, 1)
			s.cfg.Logger.Info("Successfully reconnected")
			return
		}
		
		delay := s.cfg.ReconnectDelay * time.Duration(attempt)
		s.cfg.Logger.Warn("Reconnection failed, retrying",
			zap.Duration("delay", delay),
			zap.Error(fmt.Errorf("attempt %d failed", attempt)),
		)
		
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(delay):
		}
	}
	
	s.cfg.Logger.Error("Max reconnection attempts reached")
	s.connected.Store(false)
}

// ============================================================
// Metrics
// ============================================================

type SocketStats struct {
	Connected        bool
	MessagesSent     uint64
	MessagesReceived uint64
	ParseErrors      uint64
	ReconnectCount   uint64
	HandlerPoolSize  int
}

func (s *Socket) Stats() SocketStats {
	return SocketStats{
		Connected:        s.IsConnected(),
		MessagesSent:     atomic.LoadUint64(&s.messagesSent),
		MessagesReceived: atomic.LoadUint64(&s.messagesReceived),
		ParseErrors:      atomic.LoadUint64(&s.parseErrors),
		ReconnectCount:   atomic.LoadUint64(&s.reconnectCount),
		HandlerPoolSize:  len(s.handlerPool),
	}
}

func (s *Socket) Uptime() time.Duration {
	return time.Since(s.startTime)
}