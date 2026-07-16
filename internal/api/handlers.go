package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"otc-predictor/internal/storage"
	"otc-predictor/internal/streamer"
)

type Handlers struct {
	store     *storage.Store
	cfg       *Config
	liveStore *streamer.LiveTickStore
}

func NewHandlers(store *storage.Store, cfg *Config, liveStore *streamer.LiveTickStore) *Handlers {
	return &Handlers{store: store, cfg: cfg, liveStore: liveStore}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(v)
}

// GET /api/signals?tf=1h
func (h *Handlers) GetSignals(w http.ResponseWriter, r *http.Request) {
	tf := r.URL.Query().Get("tf")
	if tf == "" {
		tf = "1h"
	}
	signals := h.store.GetAllSignals(tf)
	writeJSON(w, map[string]interface{}{
		"timeframe": tf,
		"count":     len(signals),
		"signals":   signals,
	})
}

// GET /api/signal?symbol=EURUSD&tf=1h
func (h *Handlers) GetSignal(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	tf := r.URL.Query().Get("tf")
	if tf == "" {
		tf = "1h"
	}
	sig, ok := h.store.GetSignal(symbol, tf)
	if !ok {
		http.Error(w, "signal not found", http.StatusNotFound)
		return
	}
	writeJSON(w, sig)
}

// GET /api/history?limit=50
func (h *Handlers) GetHistory(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = v
		}
	}
	history := h.store.GetHistory(limit)
	writeJSON(w, map[string]interface{}{
		"count":   len(history),
		"history": history,
	})
}

// GET /api/performance
func (h *Handlers) GetPerformance(w http.ResponseWriter, r *http.Request) {
	stats := h.store.GetPerformanceStats()
	writeJSON(w, map[string]interface{}{
		"stats": stats,
	})
}

// GET /api/outcomes
func (h *Handlers) GetOutcomes(w http.ResponseWriter, r *http.Request) {
	outcomes := h.store.GetAllOutcomes()
	writeJSON(w, map[string]interface{}{
		"count":    len(outcomes),
		"outcomes": outcomes,
	})
}

// GET /api/markets
func (h *Handlers) GetMarkets(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"forex":     h.cfg.Markets.Forex,
		"synthetic": h.cfg.Markets.Synthetic,
	})
}

// GET /api/reliability
func (h *Handlers) GetReliability(w http.ResponseWriter, r *http.Request) {
	flags := h.store.GetReliabilityFlags()
	writeJSON(w, map[string]interface{}{
		"count": len(flags),
		"reliability": flags,
	})
}

// GET /api/live-prices
func (h *Handlers) GetLivePrices(w http.ResponseWriter, r *http.Request) {
	prices := h.liveStore.All()
	writeJSON(w, map[string]interface{}{
		"count":  len(prices),
		"prices": prices,
	})
}

// GET /health
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}
