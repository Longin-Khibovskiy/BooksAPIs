package handlers

import (
	"html/template"
	"math"
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

	pageSize := 12
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if num, err := strconv.Atoi(p); err == nil && num > 0 {
			page = num
		}
	}

	var total int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM books").Scan(&total)
	if err != nil {
		http.Error(w, "Error database", http.StatusInternalServerError)
		return
	}

	offset := (page - 1) * pageSize

	rows, err := database.DB.Query(`
		SELECT id, title, author, image, publisher, rank
		FROM books
		ORDER BY $1 LIMIT $2 OFFSET $3`, sortBy, pageSize, offset)
	if err != nil {
		http.Error(w, "Error database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Book struct {
		ID        int
		Title     string
		Author    string
		Image     string
		Publisher string
		Rank      int
	}

	var books []Book
	for rows.Next() {
		var b Book
		rows.Scan(&b.ID, &b.Title, &b.Author, &b.Image, &b.Publisher, &b.Rank)
		books = append(books, b)
	}

	pages := int(math.Ceil(float64(total) / float64(pageSize)))

	userID := r.Context().Value("userID")

	data := struct {
		Books     interface{}
		Flash     string
		SortBy    string
		User      interface{}
		CSRFToken string
		PageCSS   string
		Page      int
		Pages     int
	}{
		Books:     books,
		Flash:     "",
		SortBy:    sortBy,
		User:      userID,
		CSRFToken: csrf.Token(r),
		PageCSS:   "books",
		Page:      page,
		Pages:     pages,
	}

	tmpl, err := template.New("layout").Funcs(template.FuncMap{
		"add":   func(a, b int) int { return a + b },
		"minus": func(a, b int) int { return a - b },
		"until": func(n int) []int {
			arr := make([]int, n)
			for i := range arr {
				arr[i] = i
			}
			return arr
		},
	}).ParseFiles("internal/views/layout.html", "internal/views/books.html")
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
		PageCSS     string
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
		PageCSS:     "book",
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

func RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}
