package auth

import (
	"encoding/json"
	"net/http"
	"os"

	"example.com/m/v2/internal/database"
	"example.com/m/v2/internal/utils"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		http.Error(w, "Error hashing", http.StatusInternalServerError)
		return
	}

	_, err = database.DB.Exec(`
		INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)
	`, input.Email, hash, input.Name)
	if err != nil {
		http.Error(w, "Error saving user", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Successful registration"}`))
}
