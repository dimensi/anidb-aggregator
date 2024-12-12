package anime365

type Titles struct {
	En     string `json:"en,omitempty"`
	Ja     string `json:"ja,omitempty"`
	Romaji string `json:"romaji,omitempty"`
	Ru     string `json:"ru,omitempty"`
}

type ShowType string

const (
	Preview ShowType = "preview"
	Tv      ShowType = "tv"
)

type Genre struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Data struct {
	AllTitles          []string      `json:"allTitles"`
	AniDBID            int64         `json:"aniDbId"`
	AnimeNewsNetworkID int64         `json:"animeNewsNetworkId"`
	Descriptions       []Description `json:"descriptions"`
	Episodes           []Episode     `json:"episodes"`
	FansubsID          int64         `json:"fansubsId"`
	Genres             []Genre       `json:"genres"`
	ID                 int64         `json:"id"`
	ImdbID             int64         `json:"imdbId"`
	IsActive           int64         `json:"isActive"`
	IsAiring           int64         `json:"isAiring"`
	IsHentai           int64         `json:"isHentai"`
	MyAnimeListID      int64         `json:"myAnimeListId"`
	MyAnimeListScore   string        `json:"myAnimeListScore"`
	NumberOfEpisodes   int64         `json:"numberOfEpisodes"`
	PosterURL          string        `json:"posterUrl"`
	PosterURLSmall     string        `json:"posterUrlSmall"`
	Season             string        `json:"season"`
	Title              string        `json:"title"`
	TitleLines         []string      `json:"titleLines"`
	Titles             Titles        `json:"titles"`
	Type               ShowType      `json:"type"`
	TypeTitle          string        `json:"typeTitle"`
	URL                string        `json:"url"`
	WorldArtID         int64         `json:"worldArtId"`
	WorldArtScore      string        `json:"worldArtScore"`
	WorldArtTopPlace   interface{}   `json:"worldArtTopPlace"`
	Year               int64         `json:"year"`
}

type Episode struct {
	EpisodeFull           string   `json:"episodeFull"`
	EpisodeInt            string   `json:"episodeInt"`
	EpisodeTitle          string   `json:"episodeTitle"`
	EpisodeType           ShowType `json:"episodeType"`
	FirstUploadedDateTime string   `json:"firstUploadedDateTime"`
	ID                    int64    `json:"id"`
	IsActive              int64    `json:"isActive"`
	IsFirstUploaded       int64    `json:"isFirstUploaded"`
	SeriesID              int64    `json:"seriesId"`
}

type Description struct {
	Source          string `json:"source"`
	UpdatedDateTime string `json:"updatedDateTime"`
	Value           string `json:"value"`
}
