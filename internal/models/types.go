package models

type Grobble struct {
	Track        string `json:"track"`
	Artist       string `json:"artist"`
	PlayedAt     string `json:"played_at"`
	SongDuration int    `json:"duration_ms"`
}

type ProfileResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type Account struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
}

type SpotifyResponse struct {
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
