package response

import (
	"encoding/json"
	"net/http"
)

type envelope struct {
	Data interface{} `json:"data"`
}

type errorBody struct {
	Error string `json:"error"`
}

type validationErrorBody struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Data: data})
}

func JSONList(w http.ResponseWriter, status int, data interface{}, pagination interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":       data,
		"pagination": pagination,
	})
}

func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorBody{Error: message})
}

func ValidationError(w http.ResponseWriter, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(validationErrorBody{
		Error:  "validation failed",
		Fields: fields,
	})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
