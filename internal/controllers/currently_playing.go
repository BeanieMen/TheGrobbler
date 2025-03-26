package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/BeanieMen/TheGrobbler/internal/models"
)

type PlaybackResponse struct {
	Item struct {
		Name    string `json:"name"`
		Artists []struct {
			Name string `json:"name"`
		} `json:"artists"`
		Album struct {
			Name string `json:"name"`
		} `json:"album"`
	} `json:"item"`
	ProgressMs int  `json:"progress_ms"`
	IsPlaying  bool `json:"is_playing"`
}

// unused function for logging the song being played currently for all users to stdout
func playbackPoller() {
	for {
		models.AccountsMu.RLock()
		for username, acc := range models.Accounts {
			getAndPrintPlayback(username, acc.AccessToken)
		}
		models.AccountsMu.RUnlock()
		time.Sleep(10 * time.Second)
	}
}

func getAndPrintPlayback(username, token string) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/currently-playing", nil)
	if err != nil {
		log.Printf("Error creating request for %s: %v", username, err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching playback data for %s: %v", username, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		log.Printf("%s is not playing a song.", username)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response for %s: %v", username, err)
		return
	}

	var playback PlaybackResponse
	if err := json.Unmarshal(body, &playback); err != nil {
		log.Printf("Error unmarshaling playback data for %s: %v", username, err)
		return
	}

	if !playback.IsPlaying || playback.Item.Name == "" {
		log.Printf("%s is not playing a song.", username)
		return
	}

	artist := "unknown artist"
	if len(playback.Item.Artists) > 0 {
		artist = playback.Item.Artists[0].Name
	}
	log.Printf("%s is playing %s by %s on album %s", username, playback.Item.Name, artist, playback.Item.Album.Name)
}
