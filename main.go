package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Link struct {
	Name string
	Url  string
}

type Book struct {
	ID          int
	Title       string
	Author      string
	Description string
	Publisher   string
	Image       string
	AmazonURL   string
	Rank        int
	Links       []Link
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
				Image       string `json:"book_image"`
				AmazonURL   string `json:"amazon_product_url"`
				Rank        int    `json:"rank"`
				BuyLinks    []struct {
					Name string `json:"name"`
					Url  string `json:"url"`
				} `json:"buy_links"`
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

	createTable()

	count := countBooks()
	if count == 0 {
		fmt.Println("Table is empty - download books from NYT API")
		err := updateBooksFromNYT()
		if err != nil {
			log.Fatal("Error loading books:", err)
		}
		fmt.Println("Books successfully added")
	} else {
		fmt.Printf("In table %d books — skip update table\n", count)
	}

	router := mux.NewRouter()

	fs := http.FileServer(http.Dir("stylesheets"))
	router.PathPrefix("/stylesheets/").Handler(http.StripPrefix("/stylesheets/", fs))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/books", http.StatusFound)
	})
	router.HandleFunc("/books", getBooks).Methods("GET")
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
	return db
}

func createTable() {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS books (
	    id SERIAL PRIMARY KEY,
	    title TEXT,
	    author TEXT,
	    description TEXT,
	    publisher TEXT,
	    image TEXT,
	    amazon_url TEXT,
	    rank INT,
	    created_at TIMESTAMP DEFAULT NOW()
	);
`)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS book_links (
	    id SERIAL PRIMARY KEY,
	    book_id INT REFERENCES books(id) ON DELETE CASCADE,
	    name TEXT,
	    url TEXT
	);
`)
	if err != nil {
		log.Fatal(err)
	}
}

func countBooks() int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
	if err != nil {
		log.Println("Error in counting books:", err)
	}
	return count
}

func updateBooksFromNYT() error {
	apikey := os.Getenv("API_KEY")
	if apikey == "" {
		return fmt.Errorf("API_KEY not find in .env")
	}

	url := fmt.Sprintf("https://api.nytimes.com/svc/books/v3/lists/overview.json?api-key=%s", apikey)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect to the API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	var nytResp NYTResponse
	if err := json.Unmarshal(body, &nytResp); err != nil {
		return fmt.Errorf("error JSON: %v", err)
	}

	_, err = db.Exec("DELETE FROM books")
	if err != nil {
		return fmt.Errorf("error clear table: %v", err)
	}
	_, err = db.Exec("DELETE FROM book_links")
	if err != nil {
		return fmt.Errorf("error clear table: %v", err)
	}

	for _, list := range nytResp.Results.Lists {
		for _, b := range list.Books {
			var bookID int
			err := db.QueryRow(`
				INSERT INTO books (title, author, description, publisher, image, amazon_url, rank)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id
			`, b.Title, b.Author, b.Description, b.Publisher, b.Image, b.AmazonURL, b.Rank).Scan(&bookID)
			if err != nil {
				log.Println("Error add book:", err)
				continue
			}
			for _, link := range b.BuyLinks {
				_, err := db.Exec(`
				INSERT INTO book_links (book_id, name, url)
				VALUES ($1, $2, $3)
			`, bookID, link.Name, link.Url)
				if err != nil {
					log.Println("Failed to insert link:", err)
				}
			}
		}
	}
	return nil
}

func getAllBooks() ([]Book, error) {
	rows, err := db.Query("SELECT id, title, author, description, publisher, image, amazon_url, rank FROM books ORDER BY ID")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Description, &b.Publisher, &b.Image, &b.AmazonURL, &b.Rank); err != nil {
			log.Println("Error scanning:", err)
			continue
		}
		books = append(books, b)
	}
	return books, nil
}

func getBookId(id int) (*Book, error) {
	var b Book
	err := db.QueryRow(`
		SELECT id, title, author, description, publisher, image, amazon_url, rank
		FROM books
		WHERE id=$1
	`, id).Scan(&b.ID, &b.Title, &b.Author, &b.Description, &b.Publisher, &b.Image, &b.AmazonURL, &b.Rank)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`SELECT name, url FROM book_links WHERE book_id=$1`, id)
	if err != nil {
		log.Println("Error getting links:", err)
		return &b, nil
	}
	defer rows.Close()

	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.Name, &link.Url); err != nil {
			log.Println("Error scanning link:", err)
			continue
		}
		b.Links = append(b.Links, link)
	}

	if err = rows.Err(); err != nil {
		log.Println("Rows error:", err)
	}

	return &b, nil
}

func getBooks(w http.ResponseWriter, _ *http.Request) {
	tmpl, _ := template.ParseFiles("templates/books.html")
	books, err := getAllBooks()
	if err != nil {
		http.Error(w, "Error database", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, books)
}

func getBookById(w http.ResponseWriter, r *http.Request) {
	prms := mux.Vars(r)
	id, _ := strconv.Atoi(prms["id"])
	book, err := getBookId(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	tmpl, _ := template.ParseFiles("templates/book.html")
	tmpl.Execute(w, book)
}
