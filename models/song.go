package models

import "time"

type Song struct {
	ID          string    `json:"id,omitempty"`
	GroupName   string    `json:"group_name"`
	SongTitle   string    `json:"song_title"`
	ReleaseDate time.Time `json:"release_date"`
	Text        string    `json:"text"`
	Link        string    `json:"link"`
	Verses      []string  `json:"verses"`
}

type AddSongRequest struct {
	GroupName   string    `json:"group_name"`
	SongTitle   string    `json:"song_title"`
	ReleaseDate time.Time `json:"release_date"`
	Text        string    `json:"text"`
	Link        string    `json:"link"`
}

type GetSong struct {
	GroupName   string    `json:"group_name"`
	SongTitle   string    `json:"song_title"`
	ReleaseDate time.Time `json:"release_date"`
	Text        string    `json:"text"`
	Link        string    `json:"link"`
}
