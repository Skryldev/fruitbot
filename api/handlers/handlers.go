package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fruitbot/api/responses"
	"fruitbot/internal/client"
	"fruitbot/internal/domain/enums"

	"go.uber.org/zap"
)

type Handlers struct {
	multi   *client.MultiClient
	logger  *zap.Logger
	version string
}

func NewHandlers(mc *client.MultiClient, l *zap.Logger, v string) *Handlers {
	return &Handlers{
		multi:   mc,
		logger:  l,
		version: v,
	}
}

func (h *Handlers) getOrCreateSession(ctx context.Context, restoreKey string) (*client.AccountClient, error) {
	if restoreKey == "" {
		return nil, fmt.Errorf("restore_key is required")
	}

	name := restoreKey
	if len(name) > 16 {
		name = name[:16]
	}

	return h.multi.AddAccount(ctx, name, &client.AccountOptions{
		RestoreKey: restoreKey,
	})
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	stats := h.multi.Stats()

	responses.JSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"version":   h.version,
		"accounts":  stats.TotalAccounts,
		"uptime":    "running",
		"timestamp": time.Now().Unix(),
	})
}

func (h *Handlers) GetStatus(w http.ResponseWriter, r *http.Request) {
	stats := h.multi.Stats()
	responses.Success(w, stats)
}

func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	responses.Success(w, h.multi.Stats())
}

func (h *Handlers) LoadPlayer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RestoreKey  string `json:"restore_key"`
		SaveSession bool   `json:"save_session"`
		InviteCode  string `json:"invite_code"`
		MobileModel string `json:"mobile_model"`
		DeviceName  string `json:"device_name"`
		StoreType   string `json:"store_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.BadRequest(w, "Invalid JSON body: "+err.Error())
		return
	}

	if req.RestoreKey == "" {
		responses.BadRequest(w, "restore_key is required in request body")
		return
	}

	h.logger.Info("Loading player",
		zap.String("restore_key", req.RestoreKey[:min(8, len(req.RestoreKey))]+"..."),
	)

	account, err := h.getOrCreateSession(r.Context(), req.RestoreKey)
	if err != nil {
		h.logger.Error("Failed to create session", zap.Error(err))
		responses.Error(w, err)
		return
	}

	if req.MobileModel != "" {
		account.Opts.MobileModel = req.MobileModel
	}
	if req.DeviceName != "" {
		account.Opts.DeviceName = req.DeviceName
	}
	if req.StoreType != "" {
		account.Opts.StoreType = req.StoreType
	}

	player, err := account.Client.LoadPlayer(r.Context(), &client.LoadPlayerParams{
		SaveSession: req.SaveSession,
		InviteCode:  req.InviteCode,
		MobileModel: req.MobileModel,
		DeviceName:  req.DeviceName,
		StoreType:   req.StoreType,
	})
	if err != nil {
		h.logger.Error("Failed to load player", zap.Error(err))
		responses.Error(w, err)
		return
	}

	h.logger.Info("Player loaded successfully",
		zap.String("name", fmt.Sprintf("%v", player.Data["name"])),
	)

	responses.Success(w, player.Data)
}

func (h *Handlers) GetPlayerInfo(w http.ResponseWriter, r *http.Request) {
	restoreKey := r.URL.Query().Get("restore_key")

	if restoreKey == "" {
		responses.BadRequest(w, "restore_key query parameter is required")
		return
	}

	account, err := h.getOrCreateSession(r.Context(), restoreKey)
	if err != nil {
		responses.Error(w, err)
		return
	}

	responses.Success(w, map[string]interface{}{
		"player_id":   account.Client.PlayerID(),
		"player_name": account.Client.PlayerName(),
		"restore_key": account.Client.RestoreKey(),
	})
}

func (h *Handlers) CollectMinedGold(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RestoreKey string `json:"restore_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.BadRequest(w, "Invalid JSON body")
		return
	}

	if req.RestoreKey == "" {
		responses.BadRequest(w, "restore_key is required")
		return
	}

	account, err := h.getOrCreateSession(r.Context(), req.RestoreKey)
	if err != nil {
		responses.Error(w, err)
		return
	}

	result, err := account.Client.CollectMinedGold(r.Context())
	if err != nil {
		h.logger.Error("Failed to collect gold", zap.Error(err))
		responses.Error(w, err)
		return
	}

	h.logger.Info("Gold collected",
		zap.String("restore_key", req.RestoreKey[:min(8, len(req.RestoreKey))]+"..."),
	)

	responses.Success(w, result)
}

func (h *Handlers) SendTribeMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RestoreKey string `json:"restore_key"`
		Text       string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.BadRequest(w, "Invalid JSON body")
		return
	}

	if req.RestoreKey == "" {
		responses.BadRequest(w, "restore_key is required")
		return
	}
	if req.Text == "" {
		responses.BadRequest(w, "text is required")
		return
	}

	account, err := h.getOrCreateSession(r.Context(), req.RestoreKey)
	if err != nil {
		responses.Error(w, err)
		return
	}

	msg, err := account.Client.SendTribeMessage(r.Context(), req.Text)
	if err != nil {
		responses.Error(w, err)
		return
	}

	responses.Success(w, msg)
}

func (h *Handlers) BuyCardPack(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RestoreKey string `json:"restore_key"`
		PackType   int    `json:"pack_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.BadRequest(w, "Invalid JSON body")
		return
	}

	if req.RestoreKey == "" {
		responses.BadRequest(w, "restore_key is required")
		return
	}
	if req.PackType < 1 || req.PackType > 32 {
		responses.BadRequest(w, "pack_type must be between 1 and 32")
		return
	}

	account, err := h.getOrCreateSession(r.Context(), req.RestoreKey)
	if err != nil {
		responses.Error(w, err)
		return
	}

	result, err := account.Client.BuyCardPack(r.Context(), enums.CardPackType(req.PackType))
	if err != nil {
		responses.Error(w, err)
		return
	}

	responses.Success(w, result)
}