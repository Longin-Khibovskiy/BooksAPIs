package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Book struct {
	ID          int
	Title       string
	Author      string
	Description string
	Publisher   string
	Image       string
	AmazonURL   string
	Rank        int
}

type NYTResponse struct {
	Results struct {
		Lists []struct {
			DisplayName string `json:"display_name"`
			Books       []struct {
				Title       string `json:"title"`
				Author      string `json:"author"`
				Description string `json:"description"`
				Publisher   string `json:"publisher"`
				Image       string `json:"image"`
				AmazonURL   string `json:"amazon_product_url"`
				Rank        int    `json:"rank"`
			} `json:"books"`
		} `json:"lists"`
	} `json:"results"`
}

var db *sql.DB

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env file not found")
	}
	db = initDB()
	defer db.Close()
	router := mux.NewRouter()

	fs := http.FileServer(http.Dir("stylesheets"))
	router.PathPrefix("/stylesheets/").Handler(http.StripPrefix("/stylesheets/", fs))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/books", http.StatusFound)
	})
	router.HandleFunc("/books", getBooks).Methods("GET")
	router.HandleFunc("/books/create", createBook).Methods("POST")
	router.HandleFunc("/books/delete/{id}", delBook).Methods("DELETE")
	router.HandleFunc("/book/{id}", getBookById).Methods("GET")

	fmt.Println("Server started on server http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func initDB() *sql.DB {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS books (
	    id serial PRIMARY KEY,
	    title TEXT,
	    author TEXT, 
	    description TEXT, 
	    publisher TEXT, 
	    image TEXT,
	    amazon_url TEXT,
	    rang INT,
	    created_at TIMESTAMP DEFAULT NOW()
	);
`)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func getBooks(w http.ResponseWriter, _ *http.Request) {
	tmplPath := filepath.Join("templates", "books.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Error template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, books)
}

func getBookById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	prms := mux.Vars(r)
	idx, err := strconv.Atoi(prms["id"])
	if err != nil {
		http.Error(w, "invalid book ID", http.StatusBadRequest)
		return
	}
	for _, book := range books {
		if book.ID == idx {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(book)
			return
		}
	}
	http.Error(w, "Book not found", http.StatusNotFound)
}
