package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dimensi/db-aggregator/pkg/anime365"
	"dimensi/db-aggregator/pkg/db"
	"dimensi/db-aggregator/pkg/jikan"
	jikanapi "dimensi/db-aggregator/pkg/jikan/api"
	"dimensi/db-aggregator/pkg/ratelimiter"
	"dimensi/db-aggregator/pkg/shikimori"
	shikiapi "dimensi/db-aggregator/pkg/shikimori/api"
)

const (
	ScreenshotsLimit = 5
	RolesLimit       = 5
	SimilarLimit     = 5
)

func main() {
	// Определяем флаги командной строки
	inputDir := flag.String("input", ".", "Директория с входными файлами")
	outputDir := flag.String("output", ".", "Директория для выходного файла")
	flag.Parse()

	// Открываем входной файл anime365
	anime365File, err := os.Open(filepath.Join(*inputDir, "anime365-db.jsonl"))
	if err != nil {
		log.Fatalf("Failed to open anime365 file: %v", err)
	}
	defer anime365File.Close()

	// Создаем клиенты API
	httpClient := &http.Client{}
	shikimoriClient := shikiapi.NewClient(httpClient, ratelimiter.New(3, 70))
	jikanClient := jikanapi.NewClient(httpClient, ratelimiter.New(3, 60))

	// Создаем выходной файл с timestamp в названии
	timestamp := time.Now().Unix()
	outputFileName := fmt.Sprintf("db_%d.jsonl", timestamp)
	outputPath := filepath.Join(*outputDir, outputFileName)

	// Создаем директорию для выходного файла, если её нет
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Читаем данные из anime365
	anime365Data := make([]anime365.Data, 0)
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

	// Обработка каждого аниме
	totalAnime := anime365Count
	fmt.Printf("\nНачинаю обработку %d аниме...\n", totalAnime)

	processed := 0
	for _, a365 := range anime365Data {

		// Получаем данные Shikimori
		shikiData, hasShiki := shikimoriClient.FetchAnimeData(int(a365.MyAnimeListID))

		// Получаем данные Jikan
		jikanData, hasJikan := jikanClient.FetchAnimeData(int(a365.MyAnimeListID))

		resultAnime := mapToResultAnime(a365, shikiData, hasShiki, jikanData, hasJikan)

		// Записываем результат в файл
		jsonData, err := json.Marshal(resultAnime)
		if err != nil {
			log.Printf("Ошибка маршалинга для MAL ID %d: %v", int(a365.MyAnimeListID), err)
			continue
		}
		outputFile.WriteString(string(jsonData) + "\n")

		// Обновляем прогресс
		processed++
		if processed%10 == 0 || processed == totalAnime {
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
		resultAnime.Screenshots = mapScreenshotsBase(shiki.ShikimoriData.Screenshots)
		// resultAnime.Screenshots = mapScreenshots(shiki.Screenshots, ScreenshotsLimit)
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

func mapScreenshotsBase(screenshots []shikimori.Screenshot) []db.Screenshot {
	result := make([]db.Screenshot, 0, len(screenshots))

	for _, s := range screenshots {
		result = append(result, db.Screenshot{
			Original: s.Original,
			Preview:  s.Preview,
		})
	}

	return result
}

// func mapScreenshots(screenshots []map[string]interface{}, limit int) []db.Screenshot {
// 	result := make([]db.Screenshot, 0, len(screenshots))

// 	for _, s := range screenshots {
// 		screenshot := db.Screenshot{}

// 		if original, ok := s["original"].(string); ok {
// 			screenshot.Original = original
// 		}
// 		if preview, ok := s["preview"].(string); ok {
// 			screenshot.Preview = preview
// 		}

// 		result = append(result, screenshot)

// 		if len(result) >= limit {
// 			break
// 		}
// 	}

// 	return result
// }

func mapSimilar(similar []map[string]interface{}, limit int) []db.Similar {
	result := make([]db.Similar, 0, limit)

	for _, s := range similar {
		sim := db.Similar{
			MyAnimeListID: getInt(s, "id"),
			Titles: map[string]string{
				"en": getString(s, "name"),
				"ru": getString(s, "russian"),
			},
			Score: getString(s, "score"),
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
