package auth

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"example.com/m/v2/internal/database"
	"example.com/m/v2/internal/models"
	"example.com/m/v2/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

var templates = template.Must(template.ParseFiles(
	"internal/views/layout.html",
	"internal/views/register.html",
	"internal/views/login.html",
))

type FormData map[string]interface{}

type PageData struct {
	Flash string
	Form  FormData
}

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

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "auth_token",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
	})
	w.Write([]byte(`{"message": "You have been logged out"}`))
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{}
	templates.ExecuteTemplate(w, "register.html", data)
}

func RegisterSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")

	if password != passwordConfirm {
		templates.ExecuteTemplate(w, "register.html", PageData{
			Flash: "Passwords don't match",
			Form:  FormData{"Name": name, "Email": email},
		})
		return
	}
	if len(password) < 8 {
		templates.ExecuteTemplate(w, "register.html", PageData{
			Flash: "The password must contain at least 8 characters",
			Form:  FormData{"Name": name, "Email": email},
		})
		return
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	var UserID int
	err = database.DB.QueryRow(`
		INSERT INTO users (email, password_hash, name, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, email, hash, name, time.Now()).Scan(&UserID)

	if err != nil {
		templates.ExecuteTemplate(w, "register.html", PageData{
			Flash: fmt.Sprintf("Failed to create user: %v", err),
			Form:  FormData{"Name": name, "Email": email},
		})
		return
	}

	tokenString, err := utils.CreateJWT(UserID)
	if err != nil {
		http.Error(w, "Error generation token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
	})

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	remember := r.FormValue("remember") == "on"

	var id int
	var hash string
	err := database.DB.QueryRow("SELECT id, password_hash FROM users WHERE email=$1", email).Scan(&id, &hash)
	if err != nil {
		templates.ExecuteTemplate(w, "login.html", PageData{
			Flash: "Invalid email or password",
			Form:  FormData{"Email": email},
		})
		return
	}

	if !utils.CheckPasswordHash(password, hash) {
		templates.ExecuteTemplate(w, "login.html", PageData{
			Flash: "Invalid email or password",
			Form:  FormData{"Email": email},
		})
		return
	}

	tokenString, err := utils.CreateJWT(id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
	}
	if remember {
		cookie.Expires = time.Now().Add(30 * 24 * time.Hour)
	} else {
		cookie.Expires = time.Now().Add(24 * time.Hour)
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
