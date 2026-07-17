package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"fruitbot/api/responses"
	"fruitbot/internal/client"

	"go.uber.org/zap"
)

type MultiAccountHandlers struct {
	multiClient *client.MultiClient
	logger      *zap.Logger
}

func NewMultiAccountHandlers(mc *client.MultiClient, l *zap.Logger) *MultiAccountHandlers {
	return &MultiAccountHandlers{
		multiClient: mc,
		logger:      l,
	}
}

func (h *MultiAccountHandlers) AddAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		RestoreKey  string `json:"restore_key"`
		MobileModel string `json:"mobile_model"`
		DeviceName  string `json:"device_name"`
		StoreType   string `json:"store_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.BadRequest(w, "Invalid JSON body: "+err.Error())
		return
	}

	if req.Name == "" {
		responses.BadRequest(w, "name is required")
		return
	}
	if req.RestoreKey == "" {
		responses.BadRequest(w, "restore_key is required")
		return
	}

	h.logger.Info("Adding account",
		zap.String("name", req.Name),
		zap.String("restore_key", maskKey(req.RestoreKey)),
	)

	account, err := h.multiClient.AddAccount(r.Context(), req.Name, &client.AccountOptions{
		RestoreKey:  req.RestoreKey,
		MobileModel: req.MobileModel,
		DeviceName:  req.DeviceName,
		StoreType:   req.StoreType,
	})
	if err != nil {
		h.logger.Error("Failed to add account",
			zap.String("name", req.Name),
			zap.Error(err),
		)
		responses.Error(w, err)
		return
	}

	responses.Created(w, map[string]interface{}{
		"name":         account.Name,
		"restore_key":  maskKey(req.RestoreKey),
		"message":      "Account created successfully",
		"total_accounts": h.multiClient.Stats().TotalAccounts,
	})
}

func (h *MultiAccountHandlers) ListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts := h.multiClient.ListAccounts()
	stats := h.multiClient.Stats()

	responses.Success(w, map[string]interface{}{
		"accounts":        accounts,
		"count":           len(accounts),
		"active_accounts": stats.ActiveAccounts,
		"total_accounts":  stats.TotalAccounts,
	})
}

func (h *MultiAccountHandlers) RemoveAccount(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		responses.BadRequest(w, "Account name is required in URL path")
		return
	}

	h.logger.Info("Removing account", zap.String("name", name))

	if err := h.multiClient.RemoveAccount(r.Context(), name); err != nil {
		h.logger.Error("Failed to remove account",
			zap.String("name", name),
			zap.Error(err),
		)
		responses.Error(w, err)
		return
	}

	responses.Success(w, map[string]interface{}{
		"message":        fmt.Sprintf("Account '%s' removed successfully", name),
		"total_accounts": h.multiClient.Stats().TotalAccounts,
	})
}

func (h *MultiAccountHandlers) LoadAllPlayers(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Loading all players...")

	results := h.multiClient.LoadAllPlayers(r.Context())

	successCount := 0
	failCount := 0
	response := make(map[string]interface{})

	for name, err := range results {
		if err != nil {
			failCount++
			errResp := responses.DetectGameError(err)
			response[name] = map[string]interface{}{
				"status": "error",
				"error": map[string]interface{}{
					"code":    errResp.APIError.Code,
					"message": errResp.APIError.Message,
					"details": err.Error(),
					"retry":   errResp.APIError.Retry,
				},
			}
			h.logger.Error("Failed to load player",
				zap.String("account", name),
				zap.Error(err),
			)
		} else {
			successCount++
			response[name] = map[string]interface{}{
				"status": "success",
			}
		}
	}

	responses.Success(w, map[string]interface{}{
		"summary": map[string]interface{}{
			"total":   len(results),
			"success": successCount,
			"failed":  failCount,
		},
		"results": response,
	})
}

func (h *MultiAccountHandlers) CollectGoldOnAll(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Collecting gold on all accounts...")

	results := h.multiClient.CollectGoldOnAll(r.Context())

	successCount := 0
	failCount := 0
	totalGold := 0
	response := make(map[string]interface{})

	for name, err := range results {
		if err != nil {
			failCount++
			errResp := responses.DetectGameError(err)
			response[name] = map[string]interface{}{
				"status": "error",
				"error": map[string]interface{}{
					"code":    errResp.APIError.Code,
					"message": errResp.APIError.Message,
					"details": err.Error(),
					"retry":   errResp.APIError.Retry,
				},
			}
			h.logger.Error("Failed to collect gold",
				zap.String("account", name),
				zap.Error(err),
			)
		} else {
			successCount++
			response[name] = map[string]interface{}{
				"status": "success",
			}
		}
	}

	responses.Success(w, map[string]interface{}{
		"summary": map[string]interface{}{
			"total":      len(results),
			"success":    successCount,
			"failed":     failCount,
			"total_gold": totalGold,
		},
		"results": response,
	})
}

func (h *MultiAccountHandlers) SendMessageToAll(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.BadRequest(w, "Invalid JSON body: "+err.Error())
		return
	}

	if req.Text == "" {
		responses.BadRequest(w, "text is required")
		return
	}

	h.logger.Info("Sending message to all tribes",
		zap.String("text", req.Text[:min(50, len(req.Text))]),
	)

	results := h.multiClient.SendMessageToAllTribes(r.Context(), req.Text)

	successCount := 0
	failCount := 0
	response := make(map[string]interface{})

	for name, err := range results {
		if err != nil {
			failCount++
			errResp := responses.DetectGameError(err)
			response[name] = map[string]interface{}{
				"status": "error",
				"error": map[string]interface{}{
					"code":    errResp.APIError.Code,
					"message": errResp.APIError.Message,
					"details": err.Error(),
					"retry":   errResp.APIError.Retry,
				},
			}
		} else {
			successCount++
			response[name] = map[string]interface{}{
				"status": "success",
			}
		}
	}

	responses.Success(w, map[string]interface{}{
		"summary": map[string]interface{}{
			"total":   len(results),
			"success": successCount,
			"failed":  failCount,
		},
		"results": response,
	})
}

func (h *MultiAccountHandlers) GetMultiStats(w http.ResponseWriter, r *http.Request) {
	stats := h.multiClient.Stats()
	
	responses.Success(w, map[string]interface{}{
		"total_accounts":  stats.TotalAccounts,
		"active_accounts": stats.ActiveAccounts,
		"accounts":        stats.Accounts,
		"total_requests":  stats.TotalRequests,
	})
}

func maskKey(key string) string {
	if len(key) > 8 {
		return key[:8] + "..."
	}
	if len(key) > 0 {
		return key + "..."
	}
	return "***"
}