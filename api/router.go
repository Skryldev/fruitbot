// api/router.go
package api

import (
	"net/http"

	"fruitbot/api/handlers"
)

func registerRoutes(mux *http.ServeMux, h *handlers.Handlers, mh *handlers.MultiAccountHandlers) {
	// Health & Info
	mux.HandleFunc("GET /health", h.HealthCheck)
	mux.HandleFunc("GET /api/status", h.GetStatus)
	mux.HandleFunc("GET /api/stats", h.GetStats)

	// Player Management - restore_key in request body
	mux.HandleFunc("POST /api/player/load", h.LoadPlayer)
	mux.HandleFunc("GET /api/player/info", h.GetPlayerInfo)

	// Cards
	mux.HandleFunc("POST /api/cards/collect-gold", h.CollectMinedGold)

	// Tribe
	mux.HandleFunc("POST /api/tribe/message", h.SendTribeMessage)

	// Store
	mux.HandleFunc("POST /api/store/buy-pack", h.BuyCardPack)

	// Multi-Account Management
	mux.HandleFunc("GET /api/accounts", mh.ListAccounts)
	mux.HandleFunc("POST /api/accounts", mh.AddAccount)
	mux.HandleFunc("DELETE /api/accounts/{name}", mh.RemoveAccount)
	mux.HandleFunc("POST /api/accounts/load-all", mh.LoadAllPlayers)
	mux.HandleFunc("POST /api/accounts/collect-gold-all", mh.CollectGoldOnAll)
	mux.HandleFunc("POST /api/accounts/message-all", mh.SendMessageToAll)
	mux.HandleFunc("GET /api/accounts/stats", mh.GetMultiStats)
}