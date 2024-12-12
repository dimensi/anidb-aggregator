package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

const (
	ScreenshotsLimit = 5
	RolesLimit       = 5
	SimilarLimit     = 5
)

type Anime365Titles struct {
	En     string `json:"en"`
	Ja     string `json:"ja"`
	Romaji string `json:"romaji"`
	Ru     string `json:"ru"`
}

type Anime365ShowType string

const (
	Preview Anime365ShowType = "preview"
	Tv      Anime365ShowType = "tv"
)

type Anime365Genre struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Структуры для входных данных
type Anime365Data struct {
	AllTitles          []string              `json:"allTitles"`
	AniDBID            int64                 `json:"aniDbId"`
	AnimeNewsNetworkID int64                 `json:"animeNewsNetworkId"`
	Descriptions       []Anime365Description `json:"descriptions"`
	Episodes           []Anime365Episode     `json:"episodes"`
	FansubsID          int64                 `json:"fansubsId"`
	Genres             []Anime365Genre       `json:"genres"`
	ID                 int64                 `json:"id"`
	ImdbID             int64                 `json:"imdbId"`
	IsActive           int64                 `json:"isActive"`
	IsAiring           int64                 `json:"isAiring"`
	IsHentai           int64                 `json:"isHentai"`
	MyAnimeListID      int64                 `json:"myAnimeListId"`
	MyAnimeListScore   string                `json:"myAnimeListScore"`
	NumberOfEpisodes   int64                 `json:"numberOfEpisodes"`
	PosterURL          string                `json:"posterUrl"`
	PosterURLSmall     string                `json:"posterUrlSmall"`
	Season             string                `json:"season"`
	Title              string                `json:"title"`
	TitleLines         []string              `json:"titleLines"`
	Titles             Anime365Titles        `json:"titles"`
	Type               Anime365ShowType      `json:"type"`
	TypeTitle          string                `json:"typeTitle"`
	URL                string                `json:"url"`
	WorldArtID         int64                 `json:"worldArtId"`
	WorldArtScore      string                `json:"worldArtScore"`
	WorldArtTopPlace   interface{}           `json:"worldArtTopPlace"`
	Year               int64                 `json:"year"`
}

type Anime365Episode struct {
	EpisodeFull           string           `json:"episodeFull"`
	EpisodeInt            string           `json:"episodeInt"`
	EpisodeTitle          string           `json:"episodeTitle"`
	EpisodeType           Anime365ShowType `json:"episodeType"`
	FirstUploadedDateTime string           `json:"firstUploadedDateTime"`
	ID                    int64            `json:"id"`
	IsActive              int64            `json:"isActive"`
	IsFirstUploaded       int64            `json:"isFirstUploaded"`
	SeriesID              int64            `json:"seriesId"`
}

type AniDBData struct {
	ID            int            `json:"id"`
	MyAnimeListID int            `json:"myAnimeListId"`
	Episodes      []AniDBEpisode `json:"episodes"`
}

type AniDBEpisode struct {
	ID      string      `json:"id"`
	Number  string      `json:"number"`
	AirDate string      `json:"airDate"`
	Titles  AniDBTitles `json:"titles"`
	Summary string      `json:"summary"`
	Rating  string      `json:"rating"`
}

type AniDBTitles struct {
	En   string `json:"en"`
	Fr   string `json:"fr"`
	Ja   string `json:"ja"`
	XJat string `json:"x-jat"`
}

type ShikimoriData struct {
	MyAnimeListID int                      `json:"myAnimeListId"`
	ShikimoriData ShikimoriAnimeShow       `json:"shikimoriData"`
	Roles         []map[string]interface{} `json:"roles"`
	Screenshots   []map[string]interface{} `json:"screenshots"`
	Similar       []map[string]interface{} `json:"similar"`
}

type ShikimoriAnimeShow struct {
	AiredOn       string            `json:"aired_on"`
	Anons         bool              `json:"anons"`
	Duration      int64             `json:"duration"`
	Episodes      int64             `json:"episodes"`
	EpisodesAired int64             `json:"episodes_aired"`
	Fandubbers    []string          `json:"fandubbers"`
	Fansubbers    []string          `json:"fansubbers"`
	Favoured      bool              `json:"favoured"`
	Franchise     string            `json:"franchise"`
	ID            int64             `json:"id"`
	Image         ShikimoriImage    `json:"image"`
	Kind          string            `json:"kind"`
	Ongoing       bool              `json:"ongoing"`
	ReleasedOn    string            `json:"released_on"`
	Screenshots   []Screenshot      `json:"screenshots"`
	Status        string            `json:"status"`
	Studios       []ShikimoriStudio `json:"studios"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Videos        []ShikimoriVideo  `json:"videos"`
}

type ShikimoriVideo struct {
	Hosting   string `json:"hosting"`
	ID        int64  `json:"id"`
	ImageURL  string `json:"image_url"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	PlayerURL string `json:"player_url"`
	URL       string `json:"url"`
}

type ShikimoriStudio struct {
	FilteredName string `json:"filtered_name"`
	ID           int    `json:"id"`
	Image        string `json:"image"`
	Name         string `json:"name"`
	Real         bool   `json:"real"`
}

type ShikimoriImage struct {
	Original string `json:"original,omitempty"`
	Preview  string `json:"preview,omitempty"`
	X48      string `json:"x48,omitempty"`
	X96      string `json:"x96,omitempty"`
}

// Структура для выходных данных
type ResultAnime struct {
	ID               int                   `json:"id"`
	MyAnimeListID    int                   `json:"myAnimeListId"`
	Score            string                `json:"score"`
	Titles           map[string]string     `json:"titles"`
	Type             string                `json:"type"`
	Year             int                   `json:"year"`
	Season           string                `json:"season"`
	NumberOfEpisodes int                   `json:"numberOfEpisodes"`
	Duration         int                   `json:"duration"`
	AiredOn          string                `json:"airedOn"`
	ReleasedOn       string                `json:"releasedOn"`
	Descriptions     []Anime365Description `json:"descriptions"`
	Studios          []ShikimoriStudio     `json:"studios"`
	Poster           Poster                `json:"poster"`
	Trailers         []ShikimoriVideo      `json:"trailers"`
	Genres           []Anime365Genre       `json:"genres"`
	Roles            []Role                `json:"roles"`
	Screenshots      []Screenshot          `json:"screenshots"`
	Episodes         []ResultEpisode       `json:"episodes"`
	Similar          []Similar             `json:"similar"`
}

type Anime365Description struct {
	Source          string `json:"source"`
	UpdatedDateTime string `json:"updatedDateTime"`
	Value           string `json:"value"`
}

type Poster struct {
	Anime365  ShikimoriImage `json:"anime365"`
	Shikimori ShikimoriImage `json:"shikimori"`
}

type Role struct {
	Character    Character   `json:"character"`
	Person       interface{} `json:"person"`
	Roles        []string    `json:"roles"`
	RolesRussian []string    `json:"roles_russian"`
}

type Character struct {
	ID      int            `json:"id"`
	Image   ShikimoriImage `json:"image"`
	Name    string         `json:"name"`
	Russian string         `json:"russian"`
}

type Screenshot struct {
	Original string `json:"original"`
	Preview  string `json:"preview"`
}

type ResultEpisode struct {
	Number                string            `json:"number"`
	Type                  string            `json:"type"`
	FirstUploadedDateTime string            `json:"firstUploadedDateTime"`
	ID                    int               `json:"id"`
	IsActive              int               `json:"isActive"`
	SeriesID              int               `json:"seriesId"`
	AirDate               string            `json:"airDate"`
	Titles                map[string]string `json:"titles"`
	Summary               string            `json:"summary"`
	Rating                string            `json:"rating"`
}

type Similar struct {
	AiredOn       string            `json:"aired_on"`
	Episodes      int               `json:"episodes"`
	EpisodesAired int               `json:"episodes_aired"`
	ID            int               `json:"id"`
	Image         ShikimoriImage    `json:"image"`
	Kind          string            `json:"kind"`
	Titles        map[string]string `json:"titles"`
	ReleasedOn    string            `json:"released_on"`
	Score         string            `json:"score"`
	Status        string            `json:"status"`
}

func main() {
	// Открываем входные файлы
	anime365File, err := os.Open("anime365-db.jsonl")
	if err != nil {
		log.Fatalf("Failed to open anime365 file: %v", err)
	}
	defer anime365File.Close()

	shikimoriFile, err := os.Open("shimori-db.jsonl")
	if err != nil {
		log.Fatalf("Failed to open shikimori file: %v", err)
	}
	defer shikimoriFile.Close()

	anidbFile, err := os.Open("anidb-db.jsonl")
	if err != nil {
		log.Fatalf("Failed to open anidb file: %v", err)
	}
	defer anidbFile.Close()

	// Создаем выходной файл
	outputFile, err := os.Create("db.jsonl")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Читаем данные из файлов в мапы
	anime365Data := make(map[int]Anime365Data)
	shikimoriData := make(map[int]ShikimoriData)
	anidbData := make(map[int]AniDBData)

	// Читаем anime365 данные
	scanner := bufio.NewScanner(anime365File)
	buf := make([]byte, 10*1024*1024)
	scanner.Buffer(buf, 10*1024*1024)
	for scanner.Scan() {
		var data Anime365Data
		if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
			log.Printf("Error parsing anime365 data: %v", err)
			continue
		}
		anime365Data[int(data.MyAnimeListID)] = data
	}

	// Читаем shikimori данные
	scanner = bufio.NewScanner(shikimoriFile)
	scanner.Buffer(buf, 10*1024*1024)
	for scanner.Scan() {
		var data ShikimoriData
		if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
			log.Printf("Error parsing shikimori data: %v", err)
			continue
		}
		shikimoriData[data.MyAnimeListID] = data
	}

	// Читаем anidb данные
	scanner = bufio.NewScanner(anidbFile)
	scanner.Buffer(buf, 10*1024*1024)
	for scanner.Scan() {
		var data AniDBData
		if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
			log.Printf("Error parsing anidb data: %v", err)
			continue
		}
		anidbData[data.MyAnimeListID] = data
	}

	// Обрабатываем каждое аниме
	for malID, a365 := range anime365Data {
		shiki, hasShiki := shikimoriData[malID]
		anidb, hasAnidb := anidbData[malID]

		result := mapToResultAnime(a365, shiki, hasShiki, anidb, hasAnidb)

		// Записываем результат в файл
		jsonData, err := json.Marshal(result)
		if err != nil {
			log.Printf("Error marshaling result data: %v", err)
			continue
		}
		outputFile.WriteString(string(jsonData) + "\n")
	}
}

func mapToResultAnime(a365 Anime365Data, shiki ShikimoriData, hasShiki bool,
	anidb AniDBData, hasAnidb bool) ResultAnime {

	result := ResultAnime{
		ID:               int(a365.ID),
		MyAnimeListID:    int(a365.MyAnimeListID),
		Type:             string(a365.Type),
		Year:             int(a365.Year),
		Season:           a365.Season,
		NumberOfEpisodes: int(a365.NumberOfEpisodes),
		Titles: map[string]string{
			"en":     a365.Titles.En,
			"ja":     a365.Titles.Ja,
			"romaji": a365.Titles.Romaji,
			"ru":     a365.Titles.Ru,
		},
		Descriptions: a365.Descriptions,
		Score:        a365.MyAnimeListScore,
	}

	// Маппинг данных из Shikimori
	if hasShiki {
		result.AiredOn = shiki.ShikimoriData.AiredOn
		result.ReleasedOn = shiki.ShikimoriData.ReleasedOn
		result.Duration = int(shiki.ShikimoriData.Duration)
		result.Roles = mapRoles(shiki.Roles)
		result.Screenshots = mapScreenshots(shiki.Screenshots, ScreenshotsLimit)
		result.Similar = mapSimilar(shiki.Similar, SimilarLimit)
		result.Trailers = shiki.ShikimoriData.Videos
	}

	// Маппинг данных из AniDB
	if hasAnidb {
		result.Episodes = mapEpisodes(a365.Episodes, anidb.Episodes)
	} else {
		result.Episodes = mapEpisodesWithoutAnidb(a365.Episodes)
	}

	// Маппинг постера
	result.Poster = Poster{
		Anime365: ShikimoriImage{
			Original: a365.PosterURL,
			Preview:  a365.PosterURLSmall,
		},
		Shikimori: ShikimoriImage{
			Original: "https://shikimori.one" + shiki.ShikimoriData.Image.Original,
			Preview:  "https://shikimori.one" + shiki.ShikimoriData.Image.Preview,
			X48:      "https://shikimori.one" + shiki.ShikimoriData.Image.X48,
			X96:      "https://shikimori.one" + shiki.ShikimoriData.Image.X96,
		},
	}

	// Маппинг жанров и студий
	result.Genres = a365.Genres
	result.Studios = shiki.ShikimoriData.Studios

	return result
}

func mapRoles(roles []map[string]interface{}) []Role {
	result := make([]Role, 0, len(roles))

	for _, r := range roles {
		// Проверяем, есть ли роль "Main" среди ролей персонажа
		isMain := false
		if roles, ok := r["roles"].([]interface{}); ok {
			for _, r := range roles {
				if strings.ToLower(r.(string)) == "main" {
					isMain = true
					break
				}
			}
		}

		// Пропускаем персонажа, если у него нет роли "Main"
		if !isMain {
			continue
		}

		role := Role{}

		// Маппинг персонажа
		if char, ok := r["character"].(map[string]interface{}); ok {
			role.Character = Character{
				ID:      int(char["id"].(float64)),
				Name:    char["name"].(string),
				Russian: char["russian"].(string),
			}

			if img, ok := char["image"].(map[string]interface{}); ok {
				role.Character.Image = ShikimoriImage{
					Original: "https://shikimori.one" + getString(img, "original"),
					Preview:  "https://shikimori.one" + getString(img, "preview"),
					X48:      "https://shikimori.one" + getString(img, "x48"),
					X96:      "https://shikimori.one" + getString(img, "x96"),
				}
			}
		}

		// Маппинг person (может быть null)
		role.Person = r["person"]

		// Маппинг ролей
		if roles, ok := r["roles"].([]interface{}); ok {
			role.Roles = make([]string, 0, len(roles))
			for _, r := range roles {
				role.Roles = append(role.Roles, r.(string))
			}
		}

		// Маппинг русских названий ролей
		if rolesRu, ok := r["roles_russian"].([]interface{}); ok {
			role.RolesRussian = make([]string, 0, len(rolesRu))
			for _, r := range rolesRu {
				role.RolesRussian = append(role.RolesRussian, r.(string))
			}
		}

		result = append(result, role)
	}

	return result
}

func mapScreenshots(screenshots []map[string]interface{}, limit int) []Screenshot {
	result := make([]Screenshot, 0, len(screenshots))

	for _, s := range screenshots {
		screenshot := Screenshot{}

		if original, ok := s["original"].(string); ok {
			screenshot.Original = original
		}
		if preview, ok := s["preview"].(string); ok {
			screenshot.Preview = preview
		}

		result = append(result, screenshot)

		if len(result) >= limit {
			break
		}
	}

	return result
}

func mapSimilar(similar []map[string]interface{}, limit int) []Similar {
	result := make([]Similar, 0, limit)

	for _, s := range similar {
		sim := Similar{
			AiredOn:       getString(s, "aired_on"),
			Episodes:      getInt(s, "episodes"),
			EpisodesAired: getInt(s, "episodes_aired"),
			ID:            getInt(s, "id"),
			Kind:          getString(s, "kind"),
			ReleasedOn:    getString(s, "released_on"),
			Titles: map[string]string{
				"en": getString(s, "name"),
				"ru": getString(s, "russian"),
			},
			Score:  getString(s, "score"),
			Status: getString(s, "status"),
		}

		// Маппинг изображения
		if img, ok := s["image"].(map[string]interface{}); ok {
			sim.Image = ShikimoriImage{
				Original: "https://shikimori.one" + getString(img, "original"),
				Preview:  "https://shikimori.one" + getString(img, "preview"),
				X48:      "https://shikimori.one" + getString(img, "x48"),
				X96:      "https://shikimori.one" + getString(img, "x96"),
			}
		}

		result = append(result, sim)

		if len(result) >= limit {
			break
		}
	}

	return result
}

// Вспомогательные функции для безопасного получения значений
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}

func mapEpisodes(a365Episodes []Anime365Episode, anidbEpisodes []AniDBEpisode) []ResultEpisode {
	result := make([]ResultEpisode, 0, len(a365Episodes))

	for _, ep := range a365Episodes {
		// Ищем соответствующий эпизод в AniDB данных
		var anidbEp *AniDBEpisode
		for _, aEp := range anidbEpisodes {
			if aEp.Number == ep.EpisodeInt {
				anidbEp = &aEp
				break
			}
		}

		resultEp := ResultEpisode{
			Number:                ep.EpisodeInt,
			Type:                  string(ep.EpisodeType),
			FirstUploadedDateTime: ep.FirstUploadedDateTime,
			ID:                    int(ep.ID),
			IsActive:              int(ep.IsActive),
			SeriesID:              int(ep.SeriesID),
		}

		// Добавляем данные из AniDB если они есть
		if anidbEp != nil {
			resultEp.AirDate = anidbEp.AirDate
			resultEp.Titles = map[string]string{
				"en":   anidbEp.Titles.En,
				"xjat": anidbEp.Titles.XJat,
			}
			resultEp.Summary = anidbEp.Summary
			resultEp.Rating = anidbEp.Rating
		}

		result = append(result, resultEp)
	}

	return result
}

func mapEpisodesWithoutAnidb(a365Episodes []Anime365Episode) []ResultEpisode {
	result := make([]ResultEpisode, 0, len(a365Episodes))

	for _, ep := range a365Episodes {
		resultEp := ResultEpisode{
			Number:                ep.EpisodeInt,
			Type:                  string(ep.EpisodeType),
			FirstUploadedDateTime: ep.FirstUploadedDateTime,
			ID:                    int(ep.ID),
			IsActive:              int(ep.IsActive),
			SeriesID:              int(ep.SeriesID),
		}
		result = append(result, resultEp)
	}

	return result
}
