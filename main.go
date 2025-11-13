package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"example.com/m/v2/internal/api"
	"example.com/m/v2/internal/auth"
	"example.com/m/v2/internal/database"
	"example.com/m/v2/internal/handlers"
	"example.com/m/v2/internal/middleware"
	"example.com/m/v2/internal/services"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	database.DB = database.Init()
	defer database.DB.Close()

	if err := services.InitCloudinary(); err != nil {
		log.Println("Cloudinary not configured, avatar uploads will be disabled:", err)
	}

	log.Println("Tables are managed via migrations in migrations/ folder")

	count := database.CountBooks()
	if count == 0 {
		fmt.Println("Table is empty - downloading books from NYT API")
		if err := api.UpdateBooksFromNYT(); err != nil {
			log.Fatal("Error loading books:", err)
		}
		fmt.Println("Books successfully added")
	} else {
		fmt.Printf("Database contains %d books — skipping update\n", count)
	}

	router := mux.NewRouter()

	generalLimiter := middleware.NewRateLimiter(rate.Limit(10), 20)
	generalLimiter.CleanupVisitors()

	fs := http.FileServer(http.Dir("internal/views/static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	js := http.FileServer(http.Dir("internal/views/javascript"))
	router.PathPrefix("/javascript/").Handler(http.StripPrefix("/javascript/", js))

	stylesheets := http.FileServer(http.Dir("stylesheets"))
	router.PathPrefix("/stylesheets/").Handler(http.StripPrefix("/stylesheets/", stylesheets))

	router.HandleFunc("/", handlers.RedirectToBooks)
	router.HandleFunc("/books", handlers.GetBooks).Methods("GET")
	router.HandleFunc("/book/{id}", handlers.GetBookByID).Methods("GET")

	authRouter := router.PathPrefix("").Subrouter()
	authRouter.Use(middleware.StrictRateLimit)
	authRouter.HandleFunc("/register", auth.RegisterPage).Methods("GET")
	authRouter.HandleFunc("/register", auth.RegisterSubmit).Methods("POST")
	authRouter.HandleFunc("/login", auth.LoginPage).Methods("GET")
	authRouter.HandleFunc("/login", auth.LoginSubmit).Methods("POST")
	authRouter.HandleFunc("/logout", auth.LogoutHandler).Methods("POST")

	protected := router.PathPrefix("").Subrouter()
	protected.Use(auth.AuthMiddleware)
	protected.HandleFunc("/profile", auth.ProfilePage).Methods("GET")
	protected.HandleFunc("/profile/upload-avatar", auth.UploadAvatarHandler).Methods("POST")

	csrfKey := []byte(os.Getenv("CSRF_KEY"))
	if len(csrfKey) == 0 {
		csrfKey = []byte("32-byte-long-auth-key-change-me!")
		log.Println("WARNING: Using default CSRF key. Set CSRF_KEY in production!")
	}

	csrfMiddleware := csrf.Protect(
		csrfKey,
		csrf.Secure(false),
		csrf.Path("/"),
		csrf.SameSite(csrf.SameSiteStrictMode),
	)

	handler := middleware.Logger(
		middleware.SecurityHeaders(
			generalLimiter.RateLimit(
				csrfMiddleware(router),
			),
		),
	)

	fmt.Println("Server started on http://localhost:8000")
	fmt.Println("✓ Security Headers enabled")
	fmt.Println("✓ Rate Limiting enabled (10 req/s)")
	fmt.Println("✓ CSRF Protection enabled")
	fmt.Println("✓ Request Logging enabled")
	log.Fatal(http.ListenAndServe(":8000", handler))
}
