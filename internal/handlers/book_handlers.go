package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"example.com/m/v2/internal/database"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

func GetBooks(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "rank"
	}

	books, err := database.GetAllBooks(sortBy)
	if err != nil {
		http.Error(w, "Error database", http.StatusInternalServerError)
		return
	}

	userID := r.Context().Value("userID")

	data := struct {
		Books     interface{}
		Flash     string
		SortBy    string
		User      interface{}
		CSRFToken string
	}{
		Books:     books,
		Flash:     "",
		SortBy:    sortBy,
		User:      userID,
		CSRFToken: csrf.Token(r),
	}

	tmpl, err := template.ParseFiles("internal/views/layout.html", "internal/views/books.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Lookup("layout").Execute(w, data)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetBookByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	book, err := database.GetBookByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	userID := r.Context().Value("userID")

	data := struct {
		Book        interface{}
		Flash       string
		Title       string
		Image       string
		Author      string
		Publisher   string
		Rank        int
		Description string
		Links       interface{}
		User        interface{}
		CSRFToken   string
	}{
		Book:        book,
		Flash:       "",
		Title:       book.Title,
		Image:       book.Image,
		Author:      book.Author,
		Publisher:   book.Publisher,
		Rank:        book.Rank,
		Description: book.Description,
		Links:       book.Links,
		User:        userID,
		CSRFToken:   csrf.Token(r),
	}

	tmpl, err := template.ParseFiles("internal/views/layout.html", "internal/views/book.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Lookup("layout").Execute(w, data)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func RedirectToBooks(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/books", http.StatusFound)
}

func RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}
