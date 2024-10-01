package server

import (
	"EffectiveMobileTask/internal/database"
	"time"
)

var DB *database.DB

type Song struct {
	ID          int       `json:"id,omitempty"`
	ReleaseDate time.Time `json:"release_date,omitempty"`
	Text        string    `json:"text,omitempty"`
	Link        string    `json:"link,omitempty"`
	Title       string    `json:"song,omitempty"`
	GroupName   string    `json:"group,omitempty"`
}

type SongToSend struct {
	ReleaseDate time.Time `json:"release_date"`
	Text        string    `json:"text"`
	Link        string    `json:"link"`
}

type Verses struct {
	Text []*Verse `json:"text"`
}

type Verse struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type ErrResponse struct {
	Error string `json:"error"`
}
