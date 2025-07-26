// This is a complete example for your GopherCon UK workshop
// It demonstrates parsing comments and generating HTTP handler tests

package handler

import (
	"encoding/json"
	"net/http"
)

type (
	// User represents a user in our system
	User struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	// CreateUserRequest represents the request payload for creating a user
	CreateUserRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// CreateUserResponse represents the response when creating a user
	CreateUserResponse struct {
		User    User   `json:"user"`
		Message string `json:"message"`
	}

	// ErrorResponse represents an error response
	ErrorResponse struct {
		Error   string `json:"error"`
		Code    string `json:"code"`
		Details string `json:"details,omitempty"`
	}
)

// CreateUserHandler handles user creation requests
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Invalid request",
			Code:  "INVALID_INPUT",
		})
		return
	}

	// Simulate user creation logic
	user := User{
		ID:    1,
		Name:  req.Name,
		Email: req.Email,
	}

	response := CreateUserResponse{
		User:    user,
		Message: "User created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetUserHandler handles user retrieval requests
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simulate user retrieval
	user := User{
		ID:    123,
		Name:  "Jane Smith",
		Email: "jane@example.com",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// HealthCheckHandler handles health check requests
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","timestamp":"2024-01-01T00:00:00Z"}`))
}
