package handlers

type AddClientRequest struct {
	Capacity   int    `json:"capacity"`
	IP         string `json:"ip"`
	RatePerSec int    `json:"rate_per_sec"`
}
