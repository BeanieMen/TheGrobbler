package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/BeanieMen/TheGrobbler/internal/controllers"
	"github.com/BeanieMen/TheGrobbler/internal/models"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// load caches
	models.LoadAccounts()
	models.LoadGrobbles()

	go models.TokenRefresher()
	go models.FetchGrobbles()

	http.HandleFunc("/login", controllers.LoginHandler)
	http.HandleFunc("/callback", controllers.CallbackHandler)

	log.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}
