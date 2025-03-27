package models

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// token fs
func LoadAccounts() {
	file, err := os.Open(TokensFile)
	if err != nil {
		if os.IsNotExist(err) {
			Accounts = make(map[string]Account)
			return
		}
		log.Fatalf("Error opening tokens file: %v", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&Accounts); err != nil {
		log.Fatalf("Error decoding tokens file: %v", err)
	}
}

func SaveAccounts() {
	AccountsMu.RLock()
	defer AccountsMu.RUnlock()

	file, err := os.Create(TokensFile)
	if err != nil {
		log.Printf("Error saving tokens file: %v", err)
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(Accounts); err != nil {
		log.Printf("Error encoding tokens file: %v", err)
	}
}


// token logic
func ExchangeCodeForToken(code string) (*TokenResponse, error) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenRes TokenResponse
	if err := json.Unmarshal(body, &tokenRes); err != nil {
		return nil, err
	}
	return &tokenRes, nil
}

func RefreshAccessToken(acc Account) (Account, error) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", acc.RefreshToken)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return acc, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return acc, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenRes TokenResponse
	if err := json.Unmarshal(body, &tokenRes); err != nil {
		return acc, err
	}

	acc.AccessToken = tokenRes.AccessToken
	return acc, nil
}

func TokenRefresher() {
	for {
		AccountsMu.Lock()
		for username, acc := range Accounts {
			newAcc, err := RefreshAccessToken(acc)
			if err != nil {
				log.Printf("Error refreshing token for %s: %v", username, err)
				continue
			}
			Accounts[username] = newAcc
			log.Printf("Refreshed token for user: %s", username)
		}
		AccountsMu.Unlock()
		SaveAccounts()
		time.Sleep(30 * time.Minute)
	}
}

func GetUserProfile(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var profile ProfileResponse
	if err := json.Unmarshal(body, &profile); err != nil {
		return "", err
	}
	if profile.DisplayName != "" {
		return profile.DisplayName, nil
	}
	return profile.ID, nil
}
