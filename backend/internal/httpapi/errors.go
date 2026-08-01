package httpapi

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func writeError(
	w http.ResponseWriter,
	status int,
	errorCode string,
	message string,
	details any,
) {
	writeJSON(w, status, errorResponse{
		Error:   errorCode,
		Message: message,
		Details: details,
	})
}

func decodeJSONBody(
	w http.ResponseWriter,
	r *http.Request,
	destination any,
) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_json",
			"request body must be valid JSON with recognized fields",
			nil,
		)
		return false
	}

	if decoder.More() {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_json",
			"request body must contain a single JSON object",
			nil,
		)
		return false
	}

	return true
}
