package models

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

type Grobble struct {
	Track        string `json:"track"`
	Artist       string `json:"artist"`
	PlayedAt     string `json:"played_at"`
	SongDuration int    `json:"duration_ms"`
}

func LoadGrobbles() {
	file, err := os.Open(GrobblesFile)
	if err != nil {
		if os.IsNotExist(err) {
			Grobbles = make(map[string][]Grobble)
			return
		}
		log.Fatalf("Error opening Grobbles file: %v", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&Grobbles); err != nil {
		log.Fatalf("Error decoding Grobbles file: %v", err)
	}
}

func SaveGrobbles() {
	GrobblesMu.RLock()
	defer GrobblesMu.RUnlock()

	file, err := os.Create(GrobblesFile)
	if err != nil {
		log.Printf("Error saving Grobbles file: %v", err)
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(Grobbles); err != nil {
		log.Printf("Error encoding Grobbles file: %v", err)
	}
}

func FetchGrobbles() {
	for {
		AccountsMu.RLock()
		users := make([]string, 0, len(Accounts))
		for user := range Accounts {
			users = append(users, user)
		}
		AccountsMu.RUnlock()

		for _, user := range users {
			GrobblesMu.Lock()

			newGrobbles, err := GetUserGrobbles(user)
			if err != nil {
				log.Printf("Error fetching Grobbles for %s: %v", user, err)
				GrobblesMu.Unlock()
				continue
			}

			mergedGrobbles := MergeGrobbles(Grobbles[user], newGrobbles)
			sort.Slice(mergedGrobbles, func(i, j int) bool {
				return mergedGrobbles[i].PlayedAt > mergedGrobbles[j].PlayedAt
			})
			Grobbles[user] = mergedGrobbles
			GrobblesMu.Unlock()
		}

		SaveGrobbles()

		log.Println("Grobbles updated!")
		time.Sleep(5 * time.Minute)
	}
}

func GetUserGrobbles(username string) ([]Grobble, error) {
	AccountsMu.RLock()
	account, exists := Accounts[username]
	AccountsMu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("account not found")
	}

	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/recently-played", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+account.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var response struct {
		Items []struct {
			Track struct {
				Name    string `json:"name"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
				SongDuration int `json:"duration_ms"`
			} `json:"track"`
			PlayedAt string `json:"played_at"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	var Grobbles []Grobble
	for _, item := range response.Items {
		Grobbles = append(Grobbles, Grobble{
			Track:        item.Track.Name,
			Artist:       item.Track.Artists[0].Name,
			PlayedAt:     item.PlayedAt,
			SongDuration: item.Track.SongDuration,
		})
	}

	return Grobbles, nil
}
