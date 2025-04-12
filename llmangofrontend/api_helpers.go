package llmangofrontend

import (
	"log"
	"math/rand"
	"net/http"
)

// BadRequest sends a 400 Bad Request response with the given message
func BadRequest(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

// ServerError sends a 500 Internal Server Error response with a generic message
// and logs the actual error
func ServerError(w http.ResponseWriter, err error) {
	log.Printf("Server error: %v", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

// generateUID generates a unique ID
func generateUID() string {
	return "uid_" + randomString(8)
}

// randomString generates a random string of the specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[randomInt(len(charset))]
	}
	return string(b)
}

// randomInt returns a random int in range [0, max)
func randomInt(max int) int {
	// Simple implementation for now
	return rand.Intn(max)
}
