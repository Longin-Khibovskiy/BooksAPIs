package handlers

import (
	"net/http"
	"strconv"

	"example.com/m/v2/internal/database"
	"example.com/m/v2/internal/views"
	"github.com/gorilla/mux"
)

func GetBooks(w http.ResponseWriter, r *http.Request) {
	books, err := database.GetAllBooks()
	if err != nil {
		http.Error(w, "Error database", http.StatusInternalServerError)
		return
	}

	views.BooksList(books).Render(r.Context(), w)
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

	views.BookDetail(book).Render(r.Context(), w)
}

func RedirectToBooks(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/books", http.StatusFound)
}
