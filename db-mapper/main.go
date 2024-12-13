package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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

	// Читаем данные из файлов в мапы
	anime365Data := make(map[int]anime365.Data)
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
		anime365Data[int(data.MyAnimeListID)] = data
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
	totalAnime := len(anime365Data)
	fmt.Println()
	fmt.Printf("Начинаю обработку %d аниме...\n", totalAnime)

	processed := 0
	for malID, a365 := range anime365Data {
		shiki, hasShiki := shikimoriData[malID]
		jikan, hasJikan := jikanData[malID]

		resultAnime := mapToResultAnime(a365, shiki, hasShiki, jikan, hasJikan)

		// Записываем результат в файл
		jsonData, err := json.Marshal(resultAnime)
		if err != nil {
			log.Printf("Ошибка маршалинга для MAL ID %d: %v", malID, err)
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
		log.Printf("Ошибка при пол��чении размера файла: %v", err)
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
		Year:             int(a365.Year),
		Season:           a365.Season,
		NumberOfEpisodes: int(a365.NumberOfEpisodes),
		Titles: map[string]string{
			"en":     strings.TrimSpace(a365.Titles.En),
			"ja":     strings.TrimSpace(a365.Titles.Ja),
			"romaji": strings.TrimSpace(a365.Titles.Romaji),
			"ru":     strings.TrimSpace(a365.Titles.Ru),
		},
		Descriptions: a365.Descriptions,
		Score:        a365.MyAnimeListScore,
	}

	// Маппинг данных из Shikimori
	if hasShiki {
		resultAnime.AiredOn = shiki.ShikimoriData.AiredOn
		resultAnime.ReleasedOn = shiki.ShikimoriData.ReleasedOn
		resultAnime.Duration = int(shiki.ShikimoriData.Duration)
		resultAnime.Roles = mapRoles(shiki.Roles)
		resultAnime.Screenshots = mapScreenshots(shiki.Screenshots, ScreenshotsLimit)
		resultAnime.Similar = mapSimilar(shiki.Similar, SimilarLimit)
		resultAnime.Trailers = shiki.ShikimoriData.Videos
	}

	// Маппинг данных из Jikan
	if hasJikan && len(jikan.Episodes) > 0 {
		resultAnime.Episodes = mapEpisodesFromJikan(a365.Episodes, jikan.Episodes)
	} else {
		resultAnime.Episodes = mapEpisodesWithoutJikan(a365.Episodes)
	}

	// Маппинг постера
	resultAnime.Poster = db.Poster{
		Anime365: shikimori.Image{
			Original: a365.PosterURL,
			Preview:  a365.PosterURLSmall,
		},
		Shikimori: shikimori.Image{
			Original: "https://shikimori.one" + shiki.ShikimoriData.Image.Original,
			Preview:  "https://shikimori.one" + shiki.ShikimoriData.Image.Preview,
			X48:      "https://shikimori.one" + shiki.ShikimoriData.Image.X48,
			X96:      "https://shikimori.one" + shiki.ShikimoriData.Image.X96,
		},
	}

	// Маппинг жанров и студий
	resultAnime.Genres = a365.Genres
	resultAnime.Studios = shiki.ShikimoriData.Studios

	return resultAnime
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
				role.Character.Image = shikimori.Image{
					Original: "https://shikimori.one" + getString(img, "original"),
					Preview:  "https://shikimori.one" + getString(img, "preview"),
					X48:      "https://shikimori.one" + getString(img, "x48"),
					X96:      "https://shikimori.one" + getString(img, "x96"),
				}
			}
		}

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
			sim.Image = shikimori.Image{
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
		// Ищем соответствующий эпизод в Jikan данных
		var jikanEp *jikan.Episode
		for _, jEp := range jikanEpisodes {
			if fmt.Sprint(jEp.MalID) == ep.EpisodeInt {
				jikanEp = &jEp
				break
			}
		}

		resultEp := db.Episode{
			Number:                ep.EpisodeInt,
			Type:                  string(ep.EpisodeType),
			FirstUploadedDateTime: ep.FirstUploadedDateTime,
			ID:                    int(ep.ID),
			IsActive:              int(ep.IsActive),
			SeriesID:              int(ep.SeriesID),
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
		resultEp := db.Episode{
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
