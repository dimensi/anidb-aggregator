package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"dimensi/db-aggregator/pkg/anime365"
	"dimensi/db-aggregator/pkg/db"
	"dimensi/db-aggregator/pkg/jikan"
	"dimensi/db-aggregator/pkg/shikimori"
)

const (
	ScreenshotsLimit = 5
	RolesLimit       = 5
	SimilarLimit     = 5
)

func main() {
	// Открываем входные файлы
	anime365File, err := os.Open("anime365-db.jsonl")
	if err != nil {
		log.Fatalf("Failed to open anime365 file: %v", err)
	}
	defer anime365File.Close()

	shikimoriFile, err := os.Open("shikimori-db.jsonl")
	if err != nil {
		log.Fatalf("Failed to open shikimori file: %v", err)
	}
	defer shikimoriFile.Close()

	jikanFile, err := os.Open("jikan-db.jsonl")
	if err != nil {
		log.Fatalf("Failed to open jikan file: %v", err)
	}
	defer jikanFile.Close()

	// Создаем выходной файл
	outputFile, err := os.Create("db.jsonl")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Записываем метаданные в первую строку
	metadata := struct {
		Timestamp int64 `json:"timestamp"`
	}{
		Timestamp: time.Now().Unix(),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		log.Fatalf("Failed to marshal metadata: %v", err)
	}
	outputFile.WriteString(string(metadataJSON) + "\n")

	// Читаем данные из файлов в мапы
	anime365Data := make([]anime365.Data, 0)
	shikimoriData := make(map[int]shikimori.Data)
	jikanData := make(map[int]jikan.Data)

	// Читаем anime365 данные
	fmt.Println("Читаем anime365 данные")
	scanner := bufio.NewScanner(anime365File)
	buf := make([]byte, 10*1024*1024)
	scanner.Buffer(buf, 10*1024*1024)
	anime365Count := 0
	for scanner.Scan() {
		var data anime365.Data
		if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
			continue
		}
		anime365Data = append(anime365Data, data)
		anime365Count++
		fmt.Printf("\ranime365 Count: %d", anime365Count)
	}

	// Читаем shikimori данные
	fmt.Println()
	fmt.Println("Читаем shikimori данные")
	scanner = bufio.NewScanner(shikimoriFile)
	scanner.Buffer(buf, 10*1024*1024)
	shikimoriCount := 0
	for scanner.Scan() {
		var data shikimori.Data
		if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
			log.Printf("Error parsing shikimori data: %v", err)
			continue
		}
		shikimoriData[data.MyAnimeListID] = data
		shikimoriCount++
		fmt.Printf("\rshikimori Count: %d", shikimoriCount)
	}

	// Читаем jikan данные
	fmt.Println()
	fmt.Println("Читаем jikan данные")
	scanner = bufio.NewScanner(jikanFile)
	scanner.Buffer(buf, 10*1024*1024)
	jikanCount := 0
	for scanner.Scan() {
		var data jikan.Data
		if err := json.Unmarshal(scanner.Bytes(), &data); err != nil {
			log.Printf("Error parsing jikan data: %v", err)
			continue
		}
		jikanData[data.MyAnimeListID] = data
		jikanCount++
		fmt.Printf("\rjikan Count: %d", jikanCount)
	}

	// Добавим подсчет общего количества
	totalAnime := anime365Count
	fmt.Println()
	fmt.Printf("Начинаю обработку %d аниме...\n", totalAnime)

	processed := 0
	for _, a365 := range anime365Data {
		malID := int(a365.MyAnimeListID)
		shiki, hasShiki := shikimoriData[malID]
		jikan, hasJikan := jikanData[malID]

		resultAnime := mapToResultAnime(a365, shiki, hasShiki, jikan, hasJikan)

		// Записываем результат в файл
		jsonData, err := json.Marshal(resultAnime)
		if err != nil {
			log.Printf("��шибка маршалинга для MAL ID %d: %v", malID, err)
			continue
		}
		outputFile.WriteString(string(jsonData) + "\n")

		// Обновляем прогресс
		processed++
		if processed%100 == 0 || processed == totalAnime {
			progress := float64(processed) / float64(totalAnime) * 100
			fmt.Printf("\rПрогресс: [%-50s] %.1f%% (%d/%d)",
				strings.Repeat("=", int(progress/2))+">",
				progress,
				processed,
				totalAnime,
			)
		}
	}
	fmt.Println("\nОбработка завершена!")

	// Получаем информацию о размере файла
	fileInfo, err := outputFile.Stat()
	if err != nil {
		log.Printf("Ошибка при получении размера файла: %v", err)
	} else {
		size := float64(fileInfo.Size()) / (1024 * 1024) // Конвертируем в МБ
		fmt.Printf("Размер выходного файла: %.2f МБ\n", size)
	}
}

func mapToResultAnime(a365 anime365.Data, shiki shikimori.Data, hasShiki bool,
	jikan jikan.Data, hasJikan bool) db.Anime {

	resultAnime := db.Anime{
		ID:               int(a365.ID),
		MyAnimeListID:    int(a365.MyAnimeListID),
		Type:             string(a365.Type),
		TypeTitle:        a365.TypeTitle,
		IsAiring:         int(a365.IsAiring),
		Year:             int(a365.Year),
		Season:           a365.Season,
		NumberOfEpisodes: int(a365.NumberOfEpisodes),
		Titles: map[string]string{
			"en":     strings.TrimSpace(a365.Titles.En),
			"ja":     strings.TrimSpace(a365.Titles.Ja),
			"romaji": strings.TrimSpace(a365.Titles.Romaji),
			"ru":     strings.TrimSpace(a365.Titles.Ru),
		},
		Descriptions: mapDescriptions(a365.Descriptions),
		Score:        a365.MyAnimeListScore,
		Trailers:     []db.Video{},
		Studios:      []db.Studio{},
		Genres:       mapGenres(a365.Genres),
		Roles:        []db.Role{},
		Screenshots:  []db.Screenshot{},
		Similar:      []db.Similar{},
	}

	// Маппинг данных из Shikimori
	if hasShiki {
		resultAnime.AiredOn = shiki.ShikimoriData.AiredOn
		resultAnime.ReleasedOn = shiki.ShikimoriData.ReleasedOn
		resultAnime.Duration = int(shiki.ShikimoriData.Duration)
		resultAnime.Roles = mapRoles(shiki.Roles)
		resultAnime.Screenshots = mapScreenshots(shiki.Screenshots, ScreenshotsLimit)
		resultAnime.Similar = mapSimilar(shiki.Similar, SimilarLimit)
		resultAnime.Studios = mapStudios(shiki.ShikimoriData.Studios)
		resultAnime.Trailers = mapTrailers(shiki.ShikimoriData.Videos)
	}

	// Маппинг данных из Jikan
	if hasJikan && len(jikan.Episodes) > 0 {
		resultAnime.Episodes = mapEpisodesFromJikan(a365.Episodes, jikan.Episodes)
	} else {
		resultAnime.Episodes = mapEpisodesWithoutJikan(a365.Episodes)
	}

	// Маппинг постера
	resultAnime.Poster = db.Poster{
		Anime365: db.Image{
			Original: a365.PosterURL,
			Preview:  a365.PosterURLSmall,
		},
		Shikimori: db.Image{
			Original: "https://shikimori.one" + shiki.ShikimoriData.Image.Original,
			Preview:  "https://shikimori.one" + shiki.ShikimoriData.Image.Preview,
			X48:      "https://shikimori.one" + shiki.ShikimoriData.Image.X48,
			X96:      "https://shikimori.one" + shiki.ShikimoriData.Image.X96,
		},
	}

	// Маппинг жанров и студий

	return resultAnime
}

func mapGenres(genres []anime365.Genre) []db.Genre {
	result := make([]db.Genre, 0, len(genres))

	for _, g := range genres {
		result = append(result, db.Genre{
			ID:    g.ID,
			Title: g.Title,
			URL:   g.URL,
		})
	}

	return result
}

func mapStudios(studios []shikimori.Studio) []db.Studio {
	result := make([]db.Studio, 0, len(studios))

	for _, s := range studios {
		result = append(result, db.Studio{
			ID:           s.ID,
			FilteredName: s.FilteredName,
			Image:        s.Image,
			Name:         s.Name,
			Real:         s.Real,
		})
	}

	return result
}

func mapRoles(roles []map[string]interface{}) []db.Role {
	result := make([]db.Role, 0, len(roles))

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

		role := db.Role{}

		// Маппинг персонажа
		if char, ok := r["character"].(map[string]interface{}); ok {
			role.Character = db.Character{
				ID:      int(char["id"].(float64)),
				Name:    char["name"].(string),
				Russian: char["russian"].(string),
			}

			if img, ok := char["image"].(map[string]interface{}); ok {
				role.Character.Image = db.Image{
					Original: "https://shikimori.one" + getString(img, "original"),
					Preview:  "https://shikimori.one" + getString(img, "preview"),
					X48:      "https://shikimori.one" + getString(img, "x48"),
					X96:      "https://shikimori.one" + getString(img, "x96"),
				}
			}
		}

		// Маппинг ролей
		if roles, ok := r["roles"].([]interface{}); ok {
			role.RoleNames = make([]db.RoleName, 0, len(roles))
			for _, r := range roles {
				role.RoleNames = append(role.RoleNames, db.RoleName{
					Name:    r.(string),
					Russian: "", // Будет заполнено позже
				})
			}
		}

		// Маппинг русских названий ролей
		if rolesRu, ok := r["roles_russian"].([]interface{}); ok {
			// Добавляем русские названия к существующим ролям
			for i, r := range rolesRu {
				if i < len(role.RoleNames) {
					role.RoleNames[i].Russian = r.(string)
				}
			}
		}

		result = append(result, role)
	}

	return result
}

func mapDescriptions(descriptions []anime365.Description) []db.Description {
	result := make([]db.Description, 0, len(descriptions))

	for _, d := range descriptions {
		result = append(result, db.Description{
			Source:          d.Source,
			UpdatedDateTime: d.UpdatedDateTime,
			Value:           d.Value,
		})
	}

	return result
}

func mapTrailers(trailers []shikimori.Video) []db.Video {
	result := make([]db.Video, 0, len(trailers))

	for _, t := range trailers {
		result = append(result, db.Video{
			Hosting:   t.Hosting,
			ID:        t.ID,
			ImageURL:  t.ImageURL,
			Kind:      t.Kind,
			Name:      t.Name,
			PlayerURL: t.PlayerURL,
			URL:       t.URL,
		})
	}

	return result
}

func mapScreenshots(screenshots []map[string]interface{}, limit int) []db.Screenshot {
	result := make([]db.Screenshot, 0, len(screenshots))

	for _, s := range screenshots {
		screenshot := db.Screenshot{}

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

func mapSimilar(similar []map[string]interface{}, limit int) []db.Similar {
	result := make([]db.Similar, 0, limit)

	for _, s := range similar {
		sim := db.Similar{
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
			sim.Image = db.Image{
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
		return strings.TrimSpace(val)
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

func mapEpisodesFromJikan(a365Episodes []anime365.Episode, jikanEpisodes []jikan.Episode) []db.Episode {
	result := make([]db.Episode, 0, len(a365Episodes))

	for _, ep := range a365Episodes {
		// Конвертируем строку в число
		episodeNum, _ := strconv.Atoi(ep.EpisodeInt)

		// Ищем соответствующий эпизод в Jikan данных
		var jikanEp *jikan.Episode
		for _, jEp := range jikanEpisodes {
			if fmt.Sprint(jEp.MalID) == ep.EpisodeInt {
				jikanEp = &jEp
				break
			}
		}

		resultEp := db.Episode{
			Number:                episodeNum,
			Type:                  string(ep.EpisodeType),
			Title:                 ep.EpisodeTitle,
			FirstUploadedDateTime: ep.FirstUploadedDateTime,
			ID:                    int(ep.ID),
			IsActive:              int(ep.IsActive),
			SeriesID:              int(ep.SeriesID),
			IsFirstUploaded:       int(ep.IsFirstUploaded),
		}

		// Добавляем данные из Jikan если они есть
		if jikanEp != nil {
			resultEp.AirDate = jikanEp.Aired
			resultEp.Titles = map[string]string{
				"en":     strings.TrimSpace(jikanEp.Title),
				"ja":     strings.TrimSpace(jikanEp.TitleJapanese),
				"romaji": strings.TrimSpace(jikanEp.TitleRomanji),
			}
			resultEp.Rating = fmt.Sprintf("%.2f", jikanEp.Score)
		}

		result = append(result, resultEp)
	}

	return result
}

func mapEpisodesWithoutJikan(a365Episodes []anime365.Episode) []db.Episode {
	result := make([]db.Episode, 0, len(a365Episodes))

	for _, ep := range a365Episodes {
		episodeNum, _ := strconv.Atoi(ep.EpisodeInt)

		resultEp := db.Episode{
			Number:                episodeNum,
			Type:                  string(ep.EpisodeType),
			Title:                 ep.EpisodeTitle,
			FirstUploadedDateTime: ep.FirstUploadedDateTime,
			ID:                    int(ep.ID),
			IsActive:              int(ep.IsActive),
			SeriesID:              int(ep.SeriesID),
			IsFirstUploaded:       int(ep.IsFirstUploaded),
		}
		result = append(result, resultEp)
	}

	return result
}
