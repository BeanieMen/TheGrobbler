package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/BeanieMen/TheGrobbler/internal/models"
	"github.com/BeanieMen/TheGrobbler/internal/auth"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// load caches
	models.LoadAccounts()

	go models.TokenRefresher()
	go models.FetchGrobbles()

	http.HandleFunc("/login", auth.LoginHandler)
	http.HandleFunc("/callback", auth.CallbackHandler)

	log.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}
