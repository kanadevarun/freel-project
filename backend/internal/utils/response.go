package utils

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code string `json:"code"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func Success(w http.ResponseWriter, status int, message string, data interface{}) {
	JSON(w, status, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, status int, message string, errorCode string) {
	var errDetail interface{}
	if errorCode != "" {
		errDetail = ErrorDetail{Code: errorCode}
	}
	JSON(w, status, ErrorResponse{
		Success: false,
		Message: message,
		Error:   errDetail,
	})
}
