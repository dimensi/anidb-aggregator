package db

type Anime struct {
	ID               int               `json:"id"`
	MyAnimeListID    int               `json:"myAnimeListId"`
	Score            string            `json:"score"`
	Titles           map[string]string `json:"titles"`
	Type             string            `json:"type"`
	Year             int               `json:"year"`
	Season           string            `json:"season"`
	NumberOfEpisodes int               `json:"numberOfEpisodes"`
	Duration         int               `json:"duration"`
	AiredOn          string            `json:"airedOn"`
	ReleasedOn       string            `json:"releasedOn"`
	Descriptions     []Description     `json:"descriptions"`
	Studios          []Studio          `json:"studios"`
	Poster           Poster            `json:"poster"`
	Trailers         []Video           `json:"trailers"`
	Genres           []Genre           `json:"genres"`
	Roles            []Role            `json:"roles"`
	Screenshots      []Screenshot      `json:"screenshots"`
	Episodes         []Episode         `json:"episodes"`
	Similar          []Similar         `json:"similar"`
}

type Image struct {
	Original string `json:"original,omitempty"`
	Preview  string `json:"preview,omitempty"`
	X48      string `json:"x48,omitempty"`
	X96      string `json:"x96,omitempty"`
}

type Video struct {
	Hosting   string `json:"hosting"`
	ID        int64  `json:"id"`
	ImageURL  string `json:"imageUrl"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	PlayerURL string `json:"playerUrl"`
	URL       string `json:"url"`
}

type Genre struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Studio struct {
	FilteredName string `json:"filteredName"`
	ID           int    `json:"id"`
	Image        string `json:"image"`
	Name         string `json:"name"`
	Real         bool   `json:"real"`
}

type Description struct {
	Source          string `json:"source"`
	UpdatedDateTime string `json:"updatedDateTime"`
	Value           string `json:"value"`
}

type Poster struct {
	Anime365  Image `json:"anime365"`
	Shikimori Image `json:"shikimori"`
}

type Role struct {
	Character    Character `json:"character"`
	Roles        []string  `json:"roles"`
	RolesRussian []string  `json:"rolesRussian"`
}

type Character struct {
	ID      int    `json:"id"`
	Image   Image  `json:"image"`
	Name    string `json:"name"`
	Russian string `json:"russian"`
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
	AiredOn       string            `json:"airedOn"`
	Episodes      int               `json:"episodes"`
	EpisodesAired int               `json:"episodesAired"`
	ID            int               `json:"id"`
	Image         Image             `json:"image"`
	Kind          string            `json:"kind"`
	Titles        map[string]string `json:"titles"`
	ReleasedOn    string            `json:"releasedOn"`
	Score         string            `json:"score"`
	Status        string            `json:"status"`
}
