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
var profileTmpl = template.Must(template.ParseFiles("internal/views/layout.html", "internal/views/profile.html"))

type FormData map[string]interface{}

type PageData struct {
	Flash string
	Form  FormData
	User  interface{}
}

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID")
	if userID == nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var user struct {
		ID        int
		Name      string
		Email     string
		AvatarURL string
		CreatedAt time.Time
	}

	defaultAvatar := getDefaultAvatarURL()

	var avatarURL *string
	err := database.DB.QueryRow(`
		SELECT id, name, email, created_at, avatar_url 
		FROM users 
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &avatarURL)

	if err != nil {
		http.Error(w, fmt.Sprintf("User not found: %v", err), http.StatusNotFound)
		return
	}

	if avatarURL != nil && *avatarURL != "" {
		user.AvatarURL = *avatarURL
	} else {
		user.AvatarURL = defaultAvatar
	}

	flash := r.URL.Query().Get("flash")
	data := PageData{
		User:  user,
		Flash: flash,
	}

	err = profileTmpl.Lookup("layout").Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
	}
}

func getDefaultAvatarURL() string {
	return "https://res.cloudinary.com/demo/image/upload/avatar_default.png"
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "auth_token",
		Value:   "",
		Path:    "/",
		Expires: time.Now().Add(-1 * time.Hour),
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{}
	registerTmpl.Lookup("layout").Execute(w, data)
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
		registerTmpl.Lookup("layout").Execute(w, PageData{
			Flash: "Passwords don't match",
			Form:  FormData{"Name": name, "Email": email},
		})
		return
	}
	if len(password) < 8 {
		registerTmpl.Lookup("layout").Execute(w, PageData{
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
		registerTmpl.Lookup("layout").Execute(w, PageData{
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
	loginTmpl.Lookup("layout").Execute(w, data)
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
		loginTmpl.Lookup("layout").Execute(w, PageData{
			Flash: "Invalid email or password",
			Form:  FormData{"Email": email},
		})
		return
	}

	if !utils.CheckPasswordHash(password, hash) {
		loginTmpl.Lookup("layout").Execute(w, PageData{
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
