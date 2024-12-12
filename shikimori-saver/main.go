package main

import (
	"bufio"
	"dimensi/db-aggregator/pkg/fetcher"
	"dimensi/db-aggregator/pkg/ratelimiter"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type AnimeData struct {
	MyAnimeListID int                      `json:"myAnimeListId"`
	ShikimoriData map[string]interface{}   `json:"shikimoriData"`
	Roles         []map[string]interface{} `json:"roles"`
	Screenshots   []map[string]interface{} `json:"screenshots"`
	Similar       []map[string]interface{} `json:"similar"`
}

func createEmptyAnimeData(id int) AnimeData {
	return AnimeData{
		MyAnimeListID: id,
		ShikimoriData: make(map[string]interface{}),
		Roles:         make([]map[string]interface{}, 0),
		Screenshots:   make([]map[string]interface{}, 0),
		Similar:       make([]map[string]interface{}, 0),
	}
}

func main() {
	inputFile := "anime365-db.jsonl"
	outputFile := "shikimori-db.jsonl"
	baseURL := "https://shikimori.one/api/animes/"

	log.Printf("Opening input file: %s", inputFile)
	input, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer input.Close()

	log.Printf("Creating output file: %s", outputFile)
	output, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer output.Close()

	client := &http.Client{}
	scanner := bufio.NewScanner(input)
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	// Создаем rate limiter для Shikimori API (3 запроса/сек, 70 запросов/мин)
	rateLimiter := ratelimiter.New(3, 70)
	config := fetcher.DefaultConfig()

	for scanner.Scan() {
		var anime AnimeData
		log.Printf("Processing a new line from input file")
		if err := json.Unmarshal(scanner.Bytes(), &anime); err != nil {
			log.Printf("Failed to parse JSON: %v", err)
			continue
		}

		id := anime.MyAnimeListID
		if id == 0 {
			log.Printf("Skipping entry with missing MyAnimeListId")
			continue
		}

		// Fetch main anime data
		url := fmt.Sprintf("%s%d", baseURL, id)
		log.Printf("Fetching main anime data for ID: %d", id)
		body, err := fetcher.FetchWithRetry(client, url, rateLimiter, config)
		if err != nil {
			log.Printf("Failed to fetch anime data: %v", err)
			anime = createEmptyAnimeData(id)
			jsonLine, _ := json.Marshal(anime)
			output.WriteString(string(jsonLine) + "\n")
			continue
		}

		if err := json.Unmarshal(body, &anime.ShikimoriData); err != nil {
			log.Printf("Failed to parse anime data: %v", err)
			continue
		}
		log.Printf("Successfully fetched and parsed main anime data for ID: %d", id)

		// Fetch related data
		endpoints := []string{"roles", "screenshots", "similar"}
		for _, endpoint := range endpoints {
			url := fmt.Sprintf("%s%d/%s", baseURL, id, endpoint)
			log.Printf("Fetching %s data for ID: %d", endpoint, id)
			body, err := fetcher.FetchWithRetry(client, url, rateLimiter, config)
			if err != nil {
				log.Printf("Failed to fetch %s for anime %d: %v", endpoint, id, err)
				continue
			}

			var target interface{}
			switch endpoint {
			case "roles":
				target = &anime.Roles
			case "screenshots":
				target = &anime.Screenshots
			case "similar":
				target = &anime.Similar
			}

			if err := json.Unmarshal(body, target); err != nil {
				log.Printf("Failed to parse %s data: %v", endpoint, err)
				continue
			}
			log.Printf("Successfully fetched and parsed %s data for ID: %d", endpoint, id)
		}

		// Write updated anime data to the new file
		jsonLine, err := json.Marshal(anime)
		if err != nil {
			log.Printf("Failed to marshal updated anime data: %v", err)
			continue
		}
		if _, err := output.WriteString(string(jsonLine) + "\n"); err != nil {
			log.Printf("Failed to write to output file: %v", err)
		}
		log.Printf("Successfully wrote updated anime data for ID: %d to output file", id)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	log.Printf("All data successfully saved to %s", outputFile)
}
