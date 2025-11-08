package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"example.com/m/v2/internal/database"
	"example.com/m/v2/internal/models"
)

func UpdateBooksFromNYT() error {
	apikey := os.Getenv("API_KEY")
	if apikey == "" {
		return fmt.Errorf("API_KEY not found in environment")
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

	var nytResp models.NYTResponse
	if err := json.Unmarshal(body, &nytResp); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	_, err = database.DB.Exec("DELETE FROM books")
	if err != nil {
		return fmt.Errorf("error clearing books table: %v", err)
	}
	_, err = database.DB.Exec("DELETE FROM book_links")
	if err != nil {
		return fmt.Errorf("error clearing book_links table: %v", err)
	}

	for _, list := range nytResp.Results.Lists {
		for _, b := range list.Books {
			var bookID int
			err := database.DB.QueryRow(`
				INSERT INTO books (title, author, description, publisher, image, amazon_url, rank)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id
			`, b.Title, b.Author, b.Description, b.Publisher, b.Image, b.AmazonURL, b.Rank).Scan(&bookID)
			if err != nil {
				log.Println("Error adding book:", err)
				continue
			}

			for _, link := range b.BuyLinks {
				_, err := database.DB.Exec(`
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
