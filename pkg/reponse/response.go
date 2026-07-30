package response

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

type APIResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // omit if nil
}

func SendJSONResponse(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := APIResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

func SendXMLResponse(w http.ResponseWriter, data interface{}, statusCode int) {

	w.Header().Set("Content-Type", "application/xml")

	w.WriteHeader(statusCode)

	if err := xml.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode XML response", http.StatusInternalServerError)
	}
}
