package httperror

import (
	"encoding/json"
	"net/http"
)

type errorMsg struct {
	Code  int    `json:"code"`
	Error string `json:"message"`
}

func SendError(w http.ResponseWriter, StatusCode int, message string) error {
	errorMsg := errorMsg{
		Code:  StatusCode,
		Error: message,
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(StatusCode)
	if err := json.NewEncoder(w).Encode(&errorMsg); err != nil {
		return err
	}

	return nil
}
