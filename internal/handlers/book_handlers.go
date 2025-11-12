package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"example.com/m/v2/internal/database"
	"github.com/gorilla/mux"
)

func GetBooks(w http.ResponseWriter, _ *http.Request) {
	books, err := database.GetAllBooks()
	if err != nil {
		http.Error(w, "Error database", http.StatusInternalServerError)
		return
	}

	data := struct {
		Books interface{}
		Flash string
	}{
		Books: books,
		Flash: "",
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

	data := struct {
		Book  interface{}
		Flash string
		Title       string
		Image       string
		Author      string
		Publisher   string
		Rank        int
		Description string
		Links       interface{}
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
