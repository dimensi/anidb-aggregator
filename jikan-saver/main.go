package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Episode struct {
	MalID         int     `json:"mal_id"`
	URL           string  `json:"url"`
	Title         string  `json:"title"`
	TitleJapanese string  `json:"title_japanese"`
	TitleRomanji  string  `json:"title_romanji"`
	Aired         string  `json:"aired"`
	Score         float64 `json:"score"`
	Filler        bool    `json:"filler"`
	Recap         bool    `json:"recap"`
	ForumURL      string  `json:"forum_url"`
}

type JikanResponse struct {
	Episodes   []Episode `json:"data"`
	Pagination struct {
		LastVisiblePage int  `json:"last_visible_page"`
		HasNextPage     bool `json:"has_next_page"`
	} `json:"pagination"`
}

type AnimeOutput struct {
	ID            int       `json:"id"`
	MyAnimeListID int       `json:"myAnimeListId"`
	Episodes      []Episode `json:"episodes"`
}

type AnimeInput struct {
	ID            int `json:"id"`
	MyAnimeListID int `json:"myAnimeListId"`
}

type Config struct {
	skipExisting bool
	inputFile    string
	outputFile   string
}

type RateLimiter struct {
	requests chan time.Time
	ticker   *time.Ticker
}

func NewRateLimiter(requestsPerSecond, requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		requests: make(chan time.Time, requestsPerMinute),
		ticker:   time.NewTicker(time.Second / time.Duration(requestsPerSecond)),
	}

	// Очистка минутного буфера
	go func() {
		minuteTicker := time.NewTicker(time.Minute)
		for range minuteTicker.C {
			// Очищаем канал
			for len(rl.requests) > 0 {
				<-rl.requests
			}
		}
	}()

	return rl
}

func (rl *RateLimiter) Wait() {
	<-rl.ticker.C
	rl.requests <- time.Now()

	// Если достигли лимита в минуту
	if len(rl.requests) >= 60 {
		// Ждем, пока не освободится место
		<-rl.requests
	}
}

func fetchWithRetry(client *http.Client, url string, rateLimiter *RateLimiter) ([]byte, error) {
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Fetching URL: %s (attempt %d/%d)", url, attempt, maxRetries)

		rateLimiter.Wait()

		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch URL %s: %v", url, err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}

		if resp.StatusCode == 429 {
			if attempt < maxRetries {
				log.Printf("Rate limit exceeded, waiting before retry...")
				time.Sleep(time.Second * 5)
				continue
			}
			return nil, fmt.Errorf("rate limit exceeded after %d attempts", maxRetries)
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		log.Printf("Successfully fetched URL: %s", url)
		return body, nil
	}

	return nil, fmt.Errorf("max retries reached")
}

func fetchAllEpisodes(client *http.Client, malID int, rateLimiter *RateLimiter) ([]Episode, error) {
	var allEpisodes []Episode
	page := 1
	baseURL := "https://api.jikan.moe/v4/anime/%d/episodes"

	for {
		url := fmt.Sprintf(baseURL+"?page=%d", malID, page)
		body, err := fetchWithRetry(client, url, rateLimiter)
		if err != nil {
			return nil, err
		}

		var response JikanResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %v", err)
		}

		allEpisodes = append(allEpisodes, response.Episodes...)

		if !response.Pagination.HasNextPage {
			break
		}
		page++
	}

	return allEpisodes, nil
}

func readExistingEntries(filename string) (map[int]bool, error) {
	existing := make(map[int]bool)

	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return existing, nil
		}
		return nil, fmt.Errorf("failed to open existing file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var anime AnimeOutput
		if err := json.Unmarshal(scanner.Bytes(), &anime); err != nil {
			log.Printf("Warning: Failed to parse existing entry: %v", err)
			continue
		}
		existing[anime.ID] = true
	}

	return existing, scanner.Err()
}

func main() {
	config := Config{
		skipExisting: false,
		inputFile:    "anime365-db.jsonl",
		outputFile:   "jikan-db.jsonl",
	}

	flag.BoolVar(&config.skipExisting, "skip-existing", false, "Skip processing if entry already exists in output file")
	flag.StringVar(&config.inputFile, "input", "anime365-db.jsonl", "Input file path")
	flag.StringVar(&config.outputFile, "output", "jikan-db.jsonl", "Output file path")
	flag.Parse()

	rateLimiter := NewRateLimiter(3, 60) // 3 запроса в секунду, 60 запросов в минуту

	var existing map[int]bool
	var err error

	if config.skipExisting {
		log.Printf("Reading existing entries from %s", config.outputFile)
		existing, err = readExistingEntries(config.outputFile)
		if err != nil {
			log.Fatalf("Failed to read existing entries: %v", err)
		}
		log.Printf("Found %d existing entries", len(existing))
	}

	input, err := os.Open(config.inputFile)
	if err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer input.Close()

	output, err := os.Create(config.outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer output.Close()

	client := &http.Client{}
	scanner := bufio.NewScanner(input)
	const maxCapacity = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		var animeInput AnimeInput
		if err := json.Unmarshal(scanner.Bytes(), &animeInput); err != nil {
			log.Printf("Failed to parse input JSON: %v", err)
			continue
		}

		if config.skipExisting {
			if _, exists := existing[animeInput.ID]; exists {
				log.Printf("Skipping existing entry for MAL ID: %d", animeInput.ID)
				continue
			}
		}

		if animeInput.MyAnimeListID == 0 {
			emptyData := AnimeOutput{
				ID:            animeInput.ID,
				MyAnimeListID: animeInput.MyAnimeListID,
				Episodes:      []Episode{},
			}
			jsonLine, _ := json.Marshal(emptyData)
			output.WriteString(string(jsonLine) + "\n")
			log.Printf("Skipped processing for zero MAL ID, MAL ID: %d", animeInput.MyAnimeListID)
			continue
		}

		episodes, err := fetchAllEpisodes(client, animeInput.MyAnimeListID, rateLimiter)
		if err != nil {
			log.Printf("Failed to fetch episodes for MAL ID %d: %v", animeInput.MyAnimeListID, err)
			emptyData := AnimeOutput{
				ID:            animeInput.ID,
				MyAnimeListID: animeInput.MyAnimeListID,
				Episodes:      []Episode{},
			}
			jsonLine, _ := json.Marshal(emptyData)
			output.WriteString(string(jsonLine) + "\n")
			continue
		}

		outputData := AnimeOutput{
			ID:            animeInput.ID,
			MyAnimeListID: animeInput.MyAnimeListID,
			Episodes:      episodes,
		}

		jsonLine, _ := json.Marshal(outputData)
		output.WriteString(string(jsonLine) + "\n")
		log.Printf("Successfully processed MAL ID: %d with %d episodes", animeInput.MyAnimeListID, len(episodes))
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	log.Printf("All data successfully saved to %s", config.outputFile)
}
