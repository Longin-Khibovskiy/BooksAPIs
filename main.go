package main

import (
	"fmt"
	"log"
	"net/http"

	"example.com/m/v2/internal/api"
	"example.com/m/v2/internal/database"
	"example.com/m/v2/internal/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	database.DB = database.Init()
	defer database.DB.Close()

	log.Println("Tables are managed via migrations in migrations/ folder")

	count := database.CountBooks()
	if count == 0 {
		fmt.Println("Table is empty - downloading books from NYT API")
		if err := api.UpdateBooksFromNYT(); err != nil {
			log.Fatal("Error loading books:", err)
		}
		fmt.Println("Books successfully added")
	} else {
		fmt.Printf("Database contains %d books — skipping update\n", count)
	}

	router := mux.NewRouter()

	fs := http.FileServer(http.Dir("stylesheets"))
	router.PathPrefix("/stylesheets/").Handler(http.StripPrefix("/stylesheets/", fs))

	router.HandleFunc("/", handlers.RedirectToBooks)
	router.HandleFunc("/books", handlers.GetBooks).Methods("GET")
	router.HandleFunc("/book/{id}", handlers.GetBookByID).Methods("GET")

	fmt.Println("Server started on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
