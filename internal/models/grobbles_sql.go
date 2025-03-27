package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// spotify fetching grobble bs
func SpotifyRecentlyPlayed(username string) (*http.Response, error) {
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
	return client.Do(req)
}

func ParseSpotifyResponse(body io.Reader) ([]Grobble, error) {
	var sr SpotifyResponse
	if err := json.NewDecoder(body).Decode(&sr); err != nil {
		return nil, err
	}

	var grobbles []Grobble
	for _, item := range sr.Items {
		if len(item.Track.Artists) == 0 {
			continue
		}
		grobbles = append(grobbles, Grobble{
			Track:        item.Track.Name,
			Artist:       item.Track.Artists[0].Name,
			PlayedAt:     item.PlayedAt,
			SongDuration: item.Track.SongDuration,
		})
	}
	return grobbles, nil
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
			resp, err := SpotifyRecentlyPlayed(user)
			if err != nil {
				log.Printf("Error fetching Grobbles for %s: %v", user, err)
				continue
			}

			newGrobbles, err := ParseSpotifyResponse(resp.Body)
			resp.Body.Close()

			if err != nil {
				log.Printf("Error fetching Grobbles for %s: %v", user, err)
				continue
			}

			if err := SaveGrobbles(user, newGrobbles); err != nil {
				log.Printf("Error saving Grobbles for %s: %v", user, err)
			} else {
				log.Printf("Grobbles updated for %s!", user)
			}
		}

		time.Sleep(5 * time.Minute)
	}
}


// grobble db bs
func ensureUserDBPath(username string) (string, error) {
	userDir := filepath.Join("data", username)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", err
	}
	dbPath := filepath.Join(userDir, GrobblesFile)
	return dbPath, nil
}

func initUserDB(username string) (*sql.DB, error) {
	dbPath, err := ensureUserDBPath(username)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	createTableQuery := `CREATE TABLE IF NOT EXISTS grobbles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		track TEXT,
		artist TEXT,
		played_at TEXT UNIQUE,
		duration_ms INTEGER
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func SaveGrobbles(username string, grobbles []Grobble) error {
	db, err := initUserDB(username)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT OR REPLACE INTO grobbles (track, artist, played_at, duration_ms) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, g := range grobbles {
		var exists int
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM grobbles WHERE played_at = ?)", g.PlayedAt).Scan(&exists)
		if err != nil {
			tx.Rollback()
			return err
		}

		if exists != 1  {
			_, err = stmt.Exec(g.Track, g.Artist, g.PlayedAt, g.SongDuration)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

// helper that might be used in the future, unused for now
func LoadGrobbles(username string) ([]Grobble, error) {
	db, err := initUserDB(username)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT track, artist, played_at, duration_ms FROM grobbles ORDER BY played_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grobbles []Grobble
	for rows.Next() {
		var g Grobble
		if err := rows.Scan(&g.Track, &g.Artist, &g.PlayedAt, &g.SongDuration); err != nil {
			return nil, err
		}
		grobbles = append(grobbles, g)
	}

	return grobbles, nil
}

