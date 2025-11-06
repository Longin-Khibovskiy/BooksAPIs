package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

type Book struct {
	ID     int     `json:"id"`
	Name   string  `json:"Name"`
	Author *Author `json:"Author"`
	Year   int     `json:"Year"`
}

type Author struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

var books []Book

func main() {
	router := mux.NewRouter()
	for i := 1; i <= 25; i++ {
		books = append(books, Book{
			ID:   i,
			Name: "Name of book",
			Author: &Author{
				FirstName: "Ivan",
				LastName:  "Ivanov",
			},
			Year: 2000 + i,
		})
	}

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

func createBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	book.ID = len(books) + 1
	books = append(books, book)
	json.NewEncoder(w).Encode(books)
}

func delBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	prms := mux.Vars(r)
	idx, err := strconv.Atoi(prms["id"])
	if err != nil {
		http.Error(w, "invalid book ID", http.StatusBadRequest)
		return
	}
	for i, book := range books {
		if book.ID == idx {
			books = append(books[:i], books[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Book not found", http.StatusNotFound)
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
