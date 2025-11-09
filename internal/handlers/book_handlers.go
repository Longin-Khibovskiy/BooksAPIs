package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"example.com/m/v2/internal/database"
	"github.com/gorilla/mux"
)

func GetBooks(w http.ResponseWriter, _ *http.Request) {
	tmpl, err := template.ParseFiles("internal/views/books.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	books, err := database.GetAllBooks()
	if err != nil {
		http.Error(w, "Error database", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, books)
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

	tmpl, err := template.ParseFiles("internal/views/book.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, book)
}

func RedirectToBooks(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/books", http.StatusFound)
}
