// internal/interfaces/client/client.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"fruitbot/internal/domain/enums"
	"fruitbot/internal/domain/models"
	"fruitbot/internal/domain/utils"
	"fruitbot/internal/infrastructure/config"
	"fruitbot/internal/infrastructure/crypto"
	"fruitbot/internal/infrastructure/data"
	"fruitbot/internal/infrastructure/network"
	"fruitbot/internal/infrastructure/session"

	"go.uber.org/zap"
)

// ============================================================
// Client Configuration
// ============================================================

type ClientConfig struct {
	SessionName string
	RestoreKey  string
	Passport    string
	UDID        string
	BaseURL     string
	EncVersion  crypto.Version
	Timeout     time.Duration
	Logger      *zap.Logger

	// Device info
	MobileModel string
	DeviceName  string
	StoreType   string

	// Game version
	GameVersion     string
	OSVersion       string
	ConstantVersion string
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		BaseURL:         "https://iran.fruitcraft.ir",
		EncVersion:      crypto.Version2,
		Timeout:         30 * time.Second,
		Logger:          zap.NewNop(),
		StoreType:       "bazar",
		GameVersion:     "1.10.10744",
		OSVersion:       "9",
		ConstantVersion: "142",
	}
}

type Option func(*ClientConfig)

func WithRestoreKey(restoreKey string) Option {
	return func(c *ClientConfig) {
		c.RestoreKey = restoreKey
	}
}

func WithSessionName(name string) Option {
	return func(c *ClientConfig) {
		c.SessionName = name
	}
}

func WithDeviceInfo(model, name, store string) Option {
	return func(c *ClientConfig) {
		if model != "" {
			c.MobileModel = model
		}
		if name != "" {
			c.DeviceName = name
		}
		if store != "" {
			c.StoreType = store
		}
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(c *ClientConfig) {
		c.Logger = logger
	}
}

func WithBaseURL(url string) Option {
	return func(c *ClientConfig) {
		c.BaseURL = url
	}
}

// ============================================================
// Main Client
// ============================================================

// Client is the main game client
type Client struct {
	cfg    *ClientConfig
	logger *zap.Logger

	// Core services (Dependency Injection)
	httpClient *network.HTTPClient
	socket     *network.Socket
	sessionMgr *session.SessionManager
	dataStore  *data.Store

	// State
	queueNumber     int64 // atomic
	constantVersion string
	mu              sync.RWMutex

	// Device fingerprint
	mobileModel string
	deviceName  string

	// Player info (loaded after LoadPlayer)
	playerID   int64
	playerName string
	avatarID   int64
	tribeID    int64

	// Metrics
	requestCount uint64

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

func NewClient(opts ...Option) (*Client, error) {
	cfg := DefaultClientConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &Client{
		cfg:             cfg,
		logger:          cfg.Logger,
		constantVersion: cfg.ConstantVersion,
		ctx:             ctx,
		cancel:          cancel,
	}

	if err := c.initSession(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize session: %w", err)
	}

	httpClient, err := network.NewHTTPClient(&network.ClientConfig{
		BaseURL:    cfg.BaseURL,
		Timeout:    cfg.Timeout,
		EncVersion: cfg.EncVersion,
		Passport:   cfg.Passport,
		Logger:     cfg.Logger,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}
	c.httpClient = httpClient

	socketCfg := network.DefaultSocketConfig()
	socketCfg.Logger = cfg.Logger
	c.socket = network.NewSocket(socketCfg)

	// Initialize data store
	dataStore, err := data.NewStore(nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create data store: %w", err)
	}
	c.dataStore = dataStore

	return c, nil
}

func (c *Client) initSession() error {
	// Create session storage
	storage, err := session.NewFileStorage(&session.FileStorageConfig{
		Directory: "sessions",
	})
	if err != nil {
		return err
	}

	sessionName := c.cfg.SessionName
	if sessionName == "" && c.cfg.RestoreKey != "" {
		// Use restore_key as session name (or first 8 chars)
		if len(c.cfg.RestoreKey) > 8 {
			sessionName = c.cfg.RestoreKey[:8]
		} else {
			sessionName = c.cfg.RestoreKey
		}
		c.cfg.SessionName = sessionName
	}

	c.sessionMgr, err = session.NewSessionManager(&session.SessionConfig{
		Storage:     storage,
		SessionName: sessionName,
		Logger:      c.logger,
	})
	if err != nil {
		return err
	}

	ctx := context.Background()

	if exists, _ := c.sessionMgr.DoesSessionExist(ctx); exists {
		sessionData, err := c.sessionMgr.LoadSessionData(ctx)
		if err != nil {
			c.logger.Warn("Failed to load session, creating new", zap.Error(err))
		} else {
			// Restore from saved session
			if c.cfg.RestoreKey == "" {
				c.cfg.RestoreKey = sessionData.RestoreKey
			}
			if c.cfg.Passport == "" {
				c.cfg.Passport = sessionData.Passport
			}
			if c.cfg.UDID == "" {
				c.cfg.UDID = sessionData.UDID
			}
			if c.cfg.MobileModel == "" {
				c.cfg.MobileModel = sessionData.MobileModel
			}

			c.logger.Info("Session loaded from disk",
				zap.String("session", sessionName),
				zap.String("restore_key", c.cfg.RestoreKey[:min(8, len(c.cfg.RestoreKey))]+"..."),
			)
		}
	}

	if c.cfg.Passport == "" {
		c.cfg.Passport = utils.MustGenerateRandomPassport()
	}
	if c.cfg.UDID == "" {
		c.cfg.UDID = utils.MustGenerateRandomUDID()
	}
	if c.cfg.MobileModel == "" {
		c.cfg.MobileModel = utils.GetRandomMobileModel(
			config.NewDeviceFingerprinter(),
		)
	}

	c.mobileModel = c.cfg.MobileModel
	c.deviceName = c.cfg.DeviceName
	if c.deviceName == "" {
		c.deviceName = "unknown"
	}

	c.logger.Info("Client initialized",
		zap.String("session", sessionName),
		zap.String("restore_key", func() string {
			if len(c.cfg.RestoreKey) > 8 {
				return c.cfg.RestoreKey[:8] + "..."
			}
			return c.cfg.RestoreKey
		}()),
		zap.String("mobile_model", c.mobileModel),
	)

	return nil
}

func (c *Client) RestoreKey() string {
	return c.cfg.RestoreKey
}

// ============================================================
// Request Helper
// ============================================================

func (c *Client) sendRequest(ctx context.Context, path string, input map[string]interface{}) (interface{}, error) {
	atomic.AddUint64(&c.requestCount, 1)

	return c.httpClient.Post(ctx, path, input)
}

func (c *Client) sendRequestWithMethod(ctx context.Context, method, path string, input map[string]interface{}) (interface{}, error) {
	atomic.AddUint64(&c.requestCount, 1)

	return c.httpClient.SendRequest(ctx, &network.RequestOptions{
		Method: method,
		Path:   path,
		Input:  input,
	})
}

// ============================================================
// Player Loading
// ============================================================

type LoadPlayerParams struct {
	SaveSession  bool
	KochavaUID   string
	AppsflyerUID string
	StoreType    string
	MetrixUID    string
	MobileModel  string
	DeviceName   string
	InviteCode   string
}

func (c *Client) LoadPlayer(ctx context.Context, params *LoadPlayerParams) (*PlayerLoadResponse, error) {
	if params == nil {
		params = &LoadPlayerParams{}
	}

	if params.MobileModel != "" {
		c.mobileModel = params.MobileModel
	}

	input := map[string]interface{}{
		"game_version": c.cfg.GameVersion,
		"udid":         c.cfg.UDID,
		"os_type":      2,
		"os_version":   c.cfg.OSVersion,
		"model":        c.mobileModel,
		"device_name":  params.DeviceName,
		"store_type":   params.StoreType,
	}

	if input["device_name"] == nil || input["device_name"] == "" {
		input["device_name"] = c.deviceName
	}
	if input["store_type"] == nil || input["store_type"] == "" {
		input["store_type"] = c.cfg.StoreType
	}

	if c.cfg.RestoreKey != "" {
		input["restore_key"] = c.cfg.RestoreKey
	}

	// Optional tracking IDs
	if params.MetrixUID != "" {
		input["metrix_uid"] = params.MetrixUID
	}
	if params.KochavaUID != "" {
		input["kochava_uid"] = params.KochavaUID
	}
	if params.AppsflyerUID != "" {
		input["appsflyer_uid"] = params.AppsflyerUID
	}
	if params.InviteCode != "" {
		input["invitation_ticket"] = params.InviteCode
	}

	c.logger.Info("Loading player...",
		zap.String("restore_key", func() string {
			if len(c.cfg.RestoreKey) > 8 {
				return c.cfg.RestoreKey[:8] + "..."
			}
			return c.cfg.RestoreKey
		}()),
	)

	response, err := c.sendRequest(ctx, "player/load", input)
	if err != nil {
		return nil, fmt.Errorf("failed to load player: %w", err)
	}

	respMap, ok := response.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid player load response")
	}

	if restoreKey := getString(respMap, "restore_key"); restoreKey != "" {
		c.cfg.RestoreKey = restoreKey
	}
	c.constantVersion = fmt.Sprintf("%v", respMap["latest_constants_version"])

	if q, ok := respMap["q"].(float64); ok {
		atomic.StoreInt64(&c.queueNumber, int64(q))
	}

	// Store player info
	c.playerID = getInt64(respMap, "id")
	c.playerName = getString(respMap, "name")
	c.avatarID = getInt64(respMap, "avatar_id")

	if tribe, ok := respMap["tribe"].(map[string]interface{}); ok {
		if id, ok := tribe["id"].(float64); ok {
			c.tribeID = int64(id)
		}
	}

	// Set socket info
	c.socket.SetInfo(c.playerID, c.tribeID, c.avatarID, c.playerName)

	// Save session if requested
	if params.SaveSession {
		playerInfo := &session.PlayerInfo{
			ID:        int(c.playerID),
			Name:      c.playerName,
			InviteKey: getString(respMap, "invite_key"),
		}
		if err := c.sessionMgr.SaveSession(ctx, playerInfo); err != nil {
			c.logger.Warn("Failed to save session", zap.Error(err))
		} else {
			// Also update restore_key and passport in session
			c.sessionMgr.UpdateRestoreKey(ctx, c.cfg.RestoreKey)
			c.sessionMgr.UpdatePassport(ctx, c.cfg.Passport)
		}
	}

	c.logger.Info("Player loaded successfully",
		zap.Int64("player_id", c.playerID),
		zap.String("player_name", c.playerName),
	)

	return &PlayerLoadResponse{Data: respMap}, nil
}

type PlayerLoadResponse struct {
	Data map[string]interface{}
}

// ============================================================
// Player Info Getters
// ============================================================

func (c *Client) PlayerID() int64 {
	return c.playerID
}

func (c *Client) PlayerName() string {
	return c.playerName
}

// ============================================================
// Socket Event Handlers
// ============================================================

func (c *Client) OnMessageUpdate(handler func(msg network.Message)) {
	c.socket.AddHandler("chat", handler)
}

func (c *Client) OnPlayerStatusUpdate(handler func(msg network.Message)) {
	c.socket.AddHandler("player_status", handler)
	c.socket.AddHandler("tribe_player_status", handler)
}

func (c *Client) OnBattleAlert(handler func(msg network.Message)) {
	c.socket.AddHandler("battle_request", handler)
	c.socket.AddHandler("battle_help", handler)
}

func (c *Client) OnBattleUpdate(handler func(msg network.Message)) {
	c.socket.AddHandler("battle_hero_ability", handler)
	c.socket.AddHandler("battle_update", handler)
	c.socket.AddHandler("battle_join", handler)
	c.socket.AddHandler("battle_finished", handler)
}

func (c *Client) OnAuctionUpdate(handler func(msg network.Message)) {
	c.socket.AddHandler("auction_sold", handler)
	c.socket.AddHandler("auction_bid", handler)
}

func (c *Client) OnTribeMembershipUpdate(handler func(msg network.Message)) {
	c.socket.AddHandler("tribe_join", handler)
	c.socket.AddHandler("tribe_kick", handler)
}

func (c *Client) OnSpecialEvent(pushMessageType string, handler func(msg network.Message)) {
	c.socket.AddHandler(pushMessageType, handler)
}

// ============================================================
// Game Actions
// ============================================================

func (c *Client) GetOpponents(ctx context.Context) (interface{}, error) {
	return c.sendRequest(ctx, "battle/getopponents", nil)
}

func (c *Client) GetPlayerInfo(ctx context.Context, playerID int) (interface{}, error) {
	return c.sendRequest(ctx, "player/getplayerinfo", map[string]interface{}{
		"player_id": playerID,
	})
}

func (c *Client) ChangePlayerName(ctx context.Context, name string) error {
	response, err := c.setPlayerInfo(ctx, map[string]interface{}{
		"name": name,
		"lang": "fa",
	})
	if err != nil {
		return err
	}

	if respMap, ok := response.(map[string]interface{}); ok {
		if respMap["name_changed"] != nil {
			c.playerName = name
			c.socket.SetInfo(c.playerID, c.tribeID, c.avatarID, name)
		}
	}

	return nil
}

func (c *Client) ChangePlayerAvatar(ctx context.Context, avatarID int) error {
	_, err := c.setPlayerInfo(ctx, map[string]interface{}{
		"avatar_id": avatarID,
	})
	if err != nil {
		return err
	}

	c.avatarID = int64(avatarID)
	c.socket.SetInfo(c.playerID, c.tribeID, int64(avatarID), c.playerName)
	return nil
}

func (c *Client) ChangePlayerMood(ctx context.Context, moodID enums.Mood) error {
	_, err := c.setPlayerInfo(ctx, map[string]interface{}{
		"mood_id": int(moodID),
	})
	return err
}

func (c *Client) ChangePlayerGender(ctx context.Context, gender enums.Gender) error {
	_, err := c.setPlayerInfo(ctx, map[string]interface{}{
		"gender": int(gender),
	})
	return err
}

func (c *Client) setPlayerInfo(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return c.sendRequest(ctx, "player/setplayerinfo", params)
}

// ============================================================
// Tribe Operations
// ============================================================

func (c *Client) CreateTribe(ctx context.Context, name, description string, status enums.TribeStatus) (interface{}, error) {
	response, err := c.sendRequest(ctx, "tribe/create", map[string]interface{}{
		"name":        name,
		"description": description,
		"status":      int(status),
	})
	if err != nil {
		return nil, err
	}

	if respMap, ok := response.(map[string]interface{}); ok {
		if tribe, ok := respMap["tribe"].(map[string]interface{}); ok {
			if id, ok := tribe["id"].(float64); ok {
				c.tribeID = int64(id)
				c.socket.SetInfo(c.playerID, c.tribeID, c.avatarID, c.playerName)
			}
		}
	}

	return response, nil
}

func (c *Client) JoinTribe(ctx context.Context, tribeID int) (interface{}, error) {
	response, err := c.sendRequest(ctx, "tribe/joinrequest", map[string]interface{}{
		"tribe_id": tribeID,
	})
	if err != nil {
		return nil, err
	}

	if respMap, ok := response.(map[string]interface{}); ok {
		if tribe, ok := respMap["tribe"].(map[string]interface{}); ok {
			if id, ok := tribe["id"].(float64); ok {
				c.tribeID = int64(id)
				c.socket.SetInfo(c.playerID, c.tribeID, c.avatarID, c.playerName)
			}
		}
	}

	return response, nil
}

func (c *Client) LeaveTribe(ctx context.Context) (interface{}, error) {
	response, err := c.sendRequest(ctx, "tribe/leave", nil)
	if err != nil {
		return nil, err
	}

	c.tribeID = 0
	c.socket.UnsetInfo(true)
	return response, nil
}

func (c *Client) SendTribeMessage(ctx context.Context, text string) (*network.TribeMessageData, error) {
	if !c.socket.IsConnected() {
		if _, err := c.ComebackToGame(ctx, true); err != nil {
			return nil, fmt.Errorf("failed to reconnect: %w", err)
		}
	}

	return c.socket.SendTribeMessage(text)
}

func (c *Client) GetTribeMembers(ctx context.Context, coachTribe bool) (interface{}, error) {
	return c.sendRequest(ctx, "tribe/members", map[string]interface{}{
		"coach_tribe": coachTribe,
	})
}

// ============================================================
// Battle Operations
// ============================================================

func (c *Client) AttackOpponent(ctx context.Context, params *AttackParams) (interface{}, error) {
	if params == nil {
		return nil, fmt.Errorf("attack params are required")
	}

	input := map[string]interface{}{
		"opponent_id":      params.OpponentID,
		"check":            utils.HashQueueNumber(int(atomic.LoadInt64(&c.queueNumber))),
		"attacks_in_today": params.AttacksToday,
	}

	cardIDs := params.CardIDs
	if params.HeroID != 0 && !contains(cardIDs, params.HeroID) {
		cardIDs = append(cardIDs, params.HeroID)
		input["hero_id"] = params.HeroID
	}

	input["cards"] = formatIDList(cardIDs)

	response, err := c.sendRequest(ctx, "battle/battle", input)
	if err != nil {
		return nil, err
	}

	if respMap, ok := response.(map[string]interface{}); ok {
		if q, ok := respMap["q"].(float64); ok {
			atomic.StoreInt64(&c.queueNumber, int64(q))
		}
	}

	return response, nil
}

type AttackParams struct {
	OpponentID   int
	CardIDs      []int
	HeroID       int
	AttacksToday int
}

// ============================================================
// Card Operations
// ============================================================

func (c *Client) EquipHeroItems(ctx context.Context, heroes []*models.HeroWithItems, defaultHeroID int) (interface{}, error) {
	heroDetails := make([]map[string]interface{}, len(heroes))

	for i, hero := range heroes {
		items := make([]map[string]interface{}, 0)

		for _, itemID := range hero.LeftItemIDs() {
			items = append(items, map[string]interface{}{
				"position":         1,
				"base_heroitem_id": itemID,
			})
		}

		for _, itemID := range hero.RightItemIDs() {
			items = append(items, map[string]interface{}{
				"position":         -1,
				"base_heroitem_id": itemID,
			})
		}

		heroDetails[i] = map[string]interface{}{
			"items":   items,
			"hero_id": hero.BaseHeroID(),
		}
	}

	input := map[string]interface{}{
		"hero_details": mustMarshalToString(heroDetails),
	}

	if defaultHeroID != 0 {
		input["default_hero_id"] = defaultHeroID
	}

	return c.sendRequest(ctx, "cards/equipheroitems", input)
}

func (c *Client) EnhanceCard(ctx context.Context, cardID int, sacrificeCardIDs []int) (interface{}, error) {
	return c.sendRequest(ctx, "cards/enhance", map[string]interface{}{
		"card_id":    cardID,
		"sacrifices": fmt.Sprintf("%v", sacrificeCardIDs),
	})
}

// ============================================================
// Auction Operations
// ============================================================

func (c *Client) BidUpCardInAuction(ctx context.Context, auctionID int) (interface{}, error) {
	return c.sendRequest(ctx, "auction/bid", map[string]interface{}{
		"auction_id": auctionID,
	})
}

func (c *Client) SubmitCardForAuction(ctx context.Context, cardID int) (interface{}, error) {
	return c.sendRequest(ctx, "auction/setcardforauction", map[string]interface{}{
		"card_id": cardID,
	})
}

// ============================================================
// Store Operations
// ============================================================

func (c *Client) BuyCardPack(ctx context.Context, packType enums.CardPackType) (interface{}, error) {
	return c.sendRequest(ctx, "store/buycardpack", map[string]interface{}{
		"type": int(packType),
	})
}

func (c *Client) BuyHeroCardPack(ctx context.Context, heroType enums.HeroCardPackType) (interface{}, error) {
	return c.BuyCardPack(ctx, enums.CardPackHero)
}

// ============================================================
// Connection Management
// ============================================================

func (c *Client) ComebackToGame(ctx context.Context, openSocket bool) (interface{}, error) {
	if openSocket {
		if c.playerID == 0 {
			if _, err := c.LoadPlayer(ctx, nil); err != nil {
				return nil, fmt.Errorf("failed to load player: %w", err)
			}
		}

		if err := c.socket.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect socket: %w", err)
		}
	}

	return c.sendRequest(ctx, "player/comeback", nil)
}

func (c *Client) StopUpdates() {
	c.socket.Close(false)
}

func (c *Client) Close() error {
	c.cancel()
	c.socket.Close(false)
	c.httpClient.Close()
	return c.sessionMgr.Close()
}

// ============================================================
// Utility Methods
// ============================================================

func (c *Client) GetConstants(ctx context.Context) (interface{}, error) {
	response, err := c.sendRequest(ctx, "device/constants", map[string]interface{}{
		"game_version":     c.cfg.GameVersion,
		"os_version":       c.cfg.OSVersion,
		"model":            c.mobileModel,
		"constant_version": c.constantVersion,
		"store_type":       c.cfg.StoreType,
	})
	if err != nil {
		return nil, err
	}

	if respMap, ok := response.(map[string]interface{}); ok {
		if v, ok := respMap["LATEST_CONSTANTS_VERSION"]; ok {
			c.constantVersion = fmt.Sprintf("%v", v)
		}
	}

	return response, nil
}

func (c *Client) GetAllCardsInfo(ctx context.Context) error {
	response, err := c.sendRequestWithMethod(ctx, "GET", "cards/cardsjsonexport", nil)
	if err != nil {
		return err
	}

	if data, ok := response.(map[string]interface{}); ok {
		return c.dataStore.SaveJSON("cards.json", data)
	}

	return fmt.Errorf("invalid response format")
}

func (c *Client) GetCaptcha(ctx context.Context) ([]byte, error) {
	response, err := c.sendRequestWithMethod(ctx, "GET", "bot/getcaptcha", nil)
	if err != nil {
		return nil, err
	}

	if data, ok := response.([]byte); ok {
		return data, nil
	}

	return nil, fmt.Errorf("invalid captcha response")
}

func (c *Client) SolveCaptcha(ctx context.Context, resp int) (interface{}, error) {
	return c.sendRequest(ctx, "bot/challengeresponse", map[string]interface{}{
		"resp": resp,
	})
}

func (c *Client) CollectMinedGold(ctx context.Context) (interface{}, error) {
	return c.sendRequest(ctx, "cards/collectgold", nil)
}

func (c *Client) RedeemGiftCode(ctx context.Context, code string) (interface{}, error) {
	return c.sendRequest(ctx, "player/redeemgift", map[string]interface{}{
		"code": code,
	})
}

func (c *Client) GetConfig(ctx context.Context) error {
	response, err := c.sendRequestWithMethod(ctx, "GET", "config.json", nil)
	if err != nil {
		return err
	}

	if data, ok := response.(map[string]interface{}); ok {
		return c.dataStore.SaveJSON("config.json", data)
	}

	return fmt.Errorf("invalid config response")
}

func (c *Client) DepositToBank(ctx context.Context, amount int) (interface{}, error) {
	return c.sendRequest(ctx, "player/deposittobank", map[string]interface{}{
		"amount": amount,
	})
}

func (c *Client) WithdrawFromBank(ctx context.Context, amount int) (interface{}, error) {
	return c.sendRequest(ctx, "player/withdrawfrombank", map[string]interface{}{
		"amount": amount,
	})
}

// ============================================================
// Stats
// ============================================================

// Stats returns client statistics
type Stats struct {
	RequestCount    uint64              `json:"request_count"`
	HTTPStats       network.HTTPStats   `json:"http_stats"`
	SocketStats     network.SocketStats `json:"socket_stats"`
	QueueNumber     int64               `json:"queue_number"`
	ConstantVersion string              `json:"constant_version"`
	Connected       bool                `json:"connected"`
	PlayerID        int64               `json:"player_id"`
	PlayerName      string              `json:"player_name"`
	RestoreKey      string              `json:"restore_key"`
}

func (c *Client) Stats() Stats {
	stats := Stats{
		RequestCount:    atomic.LoadUint64(&c.requestCount),
		HTTPStats:       c.httpClient.Stats(),
		SocketStats:     c.socket.Stats(),
		QueueNumber:     atomic.LoadInt64(&c.queueNumber),
		ConstantVersion: c.constantVersion,
		Connected:       c.socket.IsConnected(),
		PlayerID:        c.playerID,
		PlayerName:      c.playerName,
	}

	if len(c.cfg.RestoreKey) > 8 {
		stats.RestoreKey = c.cfg.RestoreKey[:8] + "..."
	} else {
		stats.RestoreKey = c.cfg.RestoreKey
	}

	return stats
}

// ============================================================
// Helper Functions
// ============================================================

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	return 0
}

func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func formatIDList(ids []int) string {
	return fmt.Sprintf("%v", ids)
}

func mustMarshalToString(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}