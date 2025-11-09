package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"example.com/m/v2/internal/database"
	"example.com/m/v2/internal/models"
	"example.com/m/v2/internal/utils"
	"github.com/golang-jwt/jwt/v5"
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

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	var user models.User
	err := database.DB.QueryRow(`
		SELECT id, password_hash FROM users WHERE email = $1	
	`, input.Email).Scan(&user.ID, &user.PasswordHash)

	if err != nil {
		http.Error(w, "User is not found", http.StatusUnauthorized)
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		http.Error(w, "Incorrect password", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, _ := token.SignedString(jwtSecret)

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
	})
	w.Write([]byte(`{"message": "Successfully entrance"}`))
}
