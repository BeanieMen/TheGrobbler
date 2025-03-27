package auth

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/BeanieMen/TheGrobbler/internal/models"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	scope := "user-read-playback-state user-read-private user-read-recently-played" // added user-read-private for profile

	authURL := fmt.Sprintf(
		"https://accounts.spotify.com/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=%s",
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(scope),
	)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in query parameters", http.StatusBadRequest)
		return
	}

	tokenRes, err := models.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Error exchanging code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	username, err := models.GetUserProfile(tokenRes.AccessToken)
	if err != nil {
		http.Error(w, "Error fetching user profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	models.AccountsMu.Lock()
	models.Accounts[username] = models.Account{
		AccessToken:  tokenRes.AccessToken,
		RefreshToken: tokenRes.RefreshToken,
	}
	models.AccountsMu.Unlock()
	models.SaveAccounts()

	fmt.Fprintf(w, "Authentication successful for user '%s'! You can close this window.", username)
	log.Printf("Added account for user: %s", username)
}
