package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"example.com/m/v2/internal/models"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init() *sql.DB {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	database, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return database
}

func CountBooks() int {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
	if err != nil {
		log.Println("Error in counting books:", err)
	}
	return count
}

func GetAllBooks(sortBy string) ([]models.Book, error) {
	orderClause := "ORDER BY rank ASC"
	switch sortBy {
	case "rank":
		orderClause = "ORDER BY rank ASC"
	case "title":
		orderClause = "ORDER BY title ASC"
	case "author":
		orderClause = "ORDER BY author ASC"
	default:
		orderClause = "ORDER BY rank ASC"
	}

	query := fmt.Sprintf("SELECT id, title, author, description, publisher, image, amazon_url, rank FROM books %s", orderClause)
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Description, &b.Publisher, &b.Image, &b.AmazonURL, &b.Rank); err != nil {
			log.Println("Error scanning:", err)
			continue
		}
		books = append(books, b)
	}
	return books, nil
}

func GetBookByID(id int) (*models.Book, error) {
	var b models.Book
	err := DB.QueryRow(`
		SELECT id, title, author, description, publisher, image, amazon_url, rank
		FROM books
		WHERE id=$1
	`, id).Scan(&b.ID, &b.Title, &b.Author, &b.Description, &b.Publisher, &b.Image, &b.AmazonURL, &b.Rank)
	if err != nil {
		return nil, err
	}

	rows, err := DB.Query(`SELECT name, url FROM book_links WHERE book_id=$1`, id)
	if err != nil {
		log.Println("Error getting links:", err)
		return &b, nil
	}
	defer rows.Close()

	for rows.Next() {
		var link models.Link
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
