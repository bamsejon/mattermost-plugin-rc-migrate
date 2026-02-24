package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mattermost/mattermost/server/public/plugin"
)

type MigrateRequest struct {
	ChannelID string `json:"channel_id"`
	URL       string `json:"url"`
}

type MigrateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	cfg := p.getConfiguration()

	// Verify API key
	if cfg.APISecret == "" {
		http.Error(w, "API secret not configured", http.StatusServiceUnavailable)
		return
	}

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" || apiKey != cfg.APISecret {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.URL.Path {
	case "/api/v1/migrate":
		p.handleAPIMigrate(w, r)
	case "/api/v1/unmigrate":
		p.handleAPIUnmigrate(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (p *Plugin) handleAPIMigrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MigrateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, MigrateResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
		return
	}

	if req.ChannelID == "" || req.URL == "" {
		writeJSON(w, http.StatusBadRequest, MigrateResponse{
			Status:  "error",
			Message: "channel_id and url are required",
		})
		return
	}

	// Store migration info
	info := MigrationInfo{
		URL:        req.URL,
		MigratedBy: "api",
		MigratedAt: time.Now().Unix(),
	}
	data, err := json.Marshal(info)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, MigrateResponse{
			Status:  "error",
			Message: "Failed to marshal migration info",
		})
		return
	}

	if appErr := p.API.KVSet(kvPrefix+req.ChannelID, data); appErr != nil {
		writeJSON(w, http.StatusInternalServerError, MigrateResponse{
			Status:  "error",
			Message: "Failed to save migration info: " + appErr.Error(),
		})
		return
	}

	// Update channel header
	cfg := p.getConfiguration()
	headerText := fmt.Sprintf(":no_entry: %s: %s", cfg.RedirectMessage, req.URL)

	channel, appErr := p.API.GetChannel(req.ChannelID)
	if appErr != nil {
		writeJSON(w, http.StatusInternalServerError, MigrateResponse{
			Status:  "error",
			Message: "Migration saved but failed to get channel: " + appErr.Error(),
		})
		return
	}

	channel.Header = headerText
	if _, appErr := p.API.UpdateChannel(channel); appErr != nil {
		writeJSON(w, http.StatusInternalServerError, MigrateResponse{
			Status:  "error",
			Message: "Migration saved but failed to update header: " + appErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, MigrateResponse{
		Status:  "ok",
		Message: fmt.Sprintf("Channel %s migrated to %s", req.ChannelID, req.URL),
	})
}

func (p *Plugin) handleAPIUnmigrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ChannelID string `json:"channel_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ChannelID == "" {
		writeJSON(w, http.StatusBadRequest, MigrateResponse{
			Status:  "error",
			Message: "channel_id is required",
		})
		return
	}

	if appErr := p.API.KVDelete(kvPrefix + req.ChannelID); appErr != nil {
		writeJSON(w, http.StatusInternalServerError, MigrateResponse{
			Status:  "error",
			Message: "Failed to remove migration: " + appErr.Error(),
		})
		return
	}

	channel, appErr := p.API.GetChannel(req.ChannelID)
	if appErr == nil {
		channel.Header = ""
		p.API.UpdateChannel(channel)
	}

	writeJSON(w, http.StatusOK, MigrateResponse{
		Status:  "ok",
		Message: "Migration removed",
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
