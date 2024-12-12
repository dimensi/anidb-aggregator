package shikimori

import "time"

type Data struct {
	MyAnimeListID int                      `json:"myAnimeListId"`
	ShikimoriData AnimeShow                `json:"shikimoriData"`
	Roles         []map[string]interface{} `json:"roles"`
	Screenshots   []map[string]interface{} `json:"screenshots"`
	Similar       []map[string]interface{} `json:"similar"`
}

type AnimeShow struct {
	AiredOn       string       `json:"aired_on"`
	Anons         bool         `json:"anons"`
	Duration      int64        `json:"duration"`
	Episodes      int64        `json:"episodes"`
	EpisodesAired int64        `json:"episodes_aired"`
	Fandubbers    []string     `json:"fandubbers"`
	Fansubbers    []string     `json:"fansubbers"`
	Favoured      bool         `json:"favoured"`
	Franchise     string       `json:"franchise"`
	ID            int64        `json:"id"`
	Image         Image        `json:"image"`
	Kind          string       `json:"kind"`
	Ongoing       bool         `json:"ongoing"`
	ReleasedOn    string       `json:"released_on"`
	Screenshots   []Screenshot `json:"screenshots"`
	Status        string       `json:"status"`
	Studios       []Studio     `json:"studios"`
	UpdatedAt     time.Time    `json:"updated_at"`
	Videos        []Video      `json:"videos"`
}

type Video struct {
	Hosting   string `json:"hosting"`
	ID        int64  `json:"id"`
	ImageURL  string `json:"image_url"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	PlayerURL string `json:"player_url"`
	URL       string `json:"url"`
}

type Studio struct {
	FilteredName string `json:"filtered_name"`
	ID           int    `json:"id"`
	Image        string `json:"image"`
	Name         string `json:"name"`
	Real         bool   `json:"real"`
}

type Image struct {
	Original string `json:"original,omitempty"`
	Preview  string `json:"preview,omitempty"`
	X48      string `json:"x48,omitempty"`
	X96      string `json:"x96,omitempty"`
}

type Screenshot struct {
	Original string `json:"original"`
	Preview  string `json:"preview"`
}
