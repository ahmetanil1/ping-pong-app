package dto

import "time"

type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

func NewSuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func NewErrorResponse(err string) APIResponse {
	return APIResponse{
		Success:   false,
		Error:     err,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}