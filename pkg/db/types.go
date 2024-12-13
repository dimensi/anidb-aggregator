package db

import (
	"dimensi/db-aggregator/pkg/anime365"
	"dimensi/db-aggregator/pkg/shikimori"
)

type Anime struct {
	ID               int                    `json:"id"`
	MyAnimeListID    int                    `json:"myAnimeListId"`
	Score            string                 `json:"score"`
	Titles           map[string]string      `json:"titles"`
	Type             string                 `json:"type"`
	Year             int                    `json:"year"`
	Season           string                 `json:"season"`
	NumberOfEpisodes int                    `json:"numberOfEpisodes"`
	Duration         int                    `json:"duration"`
	AiredOn          string                 `json:"airedOn"`
	ReleasedOn       string                 `json:"releasedOn"`
	Descriptions     []anime365.Description `json:"descriptions"`
	Studios          []shikimori.Studio     `json:"studios"`
	Poster           Poster                 `json:"poster"`
	Trailers         []shikimori.Video      `json:"trailers"`
	Genres           []anime365.Genre       `json:"genres"`
	Roles            []Role                 `json:"roles"`
	Screenshots      []Screenshot           `json:"screenshots"`
	Episodes         []Episode              `json:"episodes"`
	Similar          []Similar              `json:"similar"`
}

type Poster struct {
	Anime365  shikimori.Image `json:"anime365"`
	Shikimori shikimori.Image `json:"shikimori"`
}

type Role struct {
	Character    Character `json:"character"`
	Roles        []string  `json:"roles"`
	RolesRussian []string  `json:"roles_russian"`
}

type Character struct {
	ID      int             `json:"id"`
	Image   shikimori.Image `json:"image"`
	Name    string          `json:"name"`
	Russian string          `json:"russian"`
}

type Screenshot struct {
	Original string `json:"original"`
	Preview  string `json:"preview"`
}

type Episode struct {
	Number                string            `json:"number"`
	Type                  string            `json:"type"`
	FirstUploadedDateTime string            `json:"firstUploadedDateTime"`
	ID                    int               `json:"id"`
	IsActive              int               `json:"isActive"`
	SeriesID              int               `json:"seriesId"`
	AirDate               string            `json:"airDate"`
	Titles                map[string]string `json:"titles"`
	Rating                string            `json:"rating"`
}

type Similar struct {
	AiredOn       string            `json:"aired_on"`
	Episodes      int               `json:"episodes"`
	EpisodesAired int               `json:"episodes_aired"`
	ID            int               `json:"id"`
	Image         shikimori.Image   `json:"image"`
	Kind          string            `json:"kind"`
	Titles        map[string]string `json:"titles"`
	ReleasedOn    string            `json:"released_on"`
	Score         string            `json:"score"`
	Status        string            `json:"status"`
}
