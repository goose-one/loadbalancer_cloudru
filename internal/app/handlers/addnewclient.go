package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"loadbalancer/internal/domain/models"
	httperror "loadbalancer/internal/pkg/http"
	"net/http"
)

type RateLimiter interface {
	UpdateClientConfig(ctx context.Context, client models.Client) (int64, error)
}

type AddClientHandler struct {
	rl  RateLimiter
	log Logger
}

func NewAddClientHandler(rl RateLimiter, logger Logger) *AddClientHandler {
	return &AddClientHandler{
		rl:  rl,
		log: logger,
	}
}

func (h *AddClientHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	addClientReq := &AddClientRequest{}
	if err := json.NewDecoder(r.Body).Decode(&addClientReq); err != nil {
		StatusCode := http.StatusBadRequest
		if err = httperror.SendError(w, StatusCode, "Error decode parameters"); err != nil {
			h.log.Errorf("httperror.SendError %v", err.Error())
		}
		return
	}

	newClient := &models.Client{
		IP:         addClientReq.IP,
		Capacity:   addClientReq.Capacity,
		RatePerSec: addClientReq.RatePerSec,
	}

	id, err := h.rl.UpdateClientConfig(r.Context(), *newClient)
	if err != nil {
		if err = httperror.SendError(w, http.StatusInternalServerError, "Error add new client"); err != nil {
			h.log.Errorf("httperror.SendError %v", err.Error())
		}
		return
	}

	msg := fmt.Sprintf("Add new client with id %d", id)
	res := &AddClientResponse{
		Code:    http.StatusOK,
		Message: msg,
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.log.Errorf("json.NewEncode error %v", err.Error())
	}

	return
}
