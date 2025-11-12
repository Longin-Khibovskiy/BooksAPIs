package auth

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"example.com/m/v2/internal/database"
	"example.com/m/v2/internal/utils"
)

var registerTmpl = template.Must(template.ParseFiles("internal/views/layout.html", "internal/views/register.html"))
var loginTmpl = template.Must(template.ParseFiles("internal/views/layout.html", "internal/views/login.html"))

type FormData map[string]interface{}

type PageData struct {
	Flash string
	Form  FormData
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
	registerTmpl.ExecuteTemplate(w, "layout.html", data)
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
		registerTmpl.ExecuteTemplate(w, "layout.html", PageData{
			Flash: "Passwords don't match",
			Form:  FormData{"Name": name, "Email": email},
		})
		return
	}
	if len(password) < 8 {
		registerTmpl.ExecuteTemplate(w, "layout.html", PageData{
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
		registerTmpl.ExecuteTemplate(w, "layout.html", PageData{
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
	data := PageData{}
	if r.URL.Query().Get("registered") == "1" {
		data.Flash = "Successfully registered. Enter email and password"
	}
	loginTmpl.ExecuteTemplate(w, "layout.html", data)
}

func LoginSubmit(w http.ResponseWriter, r *http.Request) {
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
		loginTmpl.ExecuteTemplate(w, "layout.html", PageData{
			Flash: "Invalid email or password",
			Form:  FormData{"Email": email},
		})
		return
	}

	if !utils.CheckPasswordHash(password, hash) {
		loginTmpl.ExecuteTemplate(w, "layout.html", PageData{
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
