package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
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

func fetchWithRateLimit(client *http.Client, url string) ([]byte, error) {
	log.Printf("Fetching URL: %s", url)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for URL %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	log.Printf("Successfully fetched URL: %s", url)
	return body, nil
}

func fetchWithRetry(client *http.Client, url string, maxRetries int) ([]byte, error) {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		body, err := fetchWithRateLimit(client, url)
		if err == nil {
			return body, nil
		}

		lastErr = err
		retryDelay := time.Duration(attempt) * 10 * time.Second
		log.Printf("Attempt %d/%d failed: %v. Retrying in %v...",
			attempt, maxRetries, err, retryDelay)
		time.Sleep(retryDelay)
	}
	return nil, fmt.Errorf("failed after %d attempts. Last error: %v", maxRetries, lastErr)
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
	requestCount := 0
	startMinute := time.Now()

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

		// Rate limiting logic
		if requestCount >= 70 {
			elapsed := time.Since(startMinute)
			if elapsed < time.Minute {
				log.Printf("Rate limit reached, sleeping for %v", time.Minute-elapsed)
				time.Sleep(time.Minute - elapsed)
			}
			requestCount = 0
			startMinute = time.Now()
		}

		// Fetch main anime data
		url := fmt.Sprintf("%s%d", baseURL, id)
		log.Printf("Fetching main anime data for ID: %d", id)
		body, err := fetchWithRetry(client, url, 3)
		if err != nil {
			log.Printf("Failed to fetch anime data after all retries: %v", err)
			anime = createEmptyAnimeData(id)
			jsonLine, err := json.Marshal(anime)
			if err != nil {
				log.Printf("Failed to marshal empty anime data: %v", err)
				continue
			}
			if _, err := output.WriteString(string(jsonLine) + "\n"); err != nil {
				log.Printf("Failed to write empty anime data to output file: %v", err)
			}
			log.Printf("Wrote empty anime data for ID: %d to output file", id)
			continue
		}

		if err := json.Unmarshal(body, &anime.ShikimoriData); err != nil {
			log.Printf("Failed to parse anime data: %v", err)
			continue
		}
		log.Printf("Successfully fetched and parsed main anime data for ID: %d", id)
		requestCount++

		// Fetch related data
		endpoints := []string{"roles", "screenshots", "similar"}
		for _, endpoint := range endpoints {
			url := fmt.Sprintf("%s%d/%s", baseURL, id, endpoint)
			log.Printf("Fetching %s data for ID: %d", endpoint, id)
			body, err := fetchWithRetry(client, url, 3)
			if err != nil {
				log.Printf("Failed to fetch %s for anime %d: %v", endpoint, id, err)
				continue
			}

			if len(body) == 0 {
				log.Printf("Empty response for %s data, skipping", endpoint)
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
				log.Printf("Response body for %s: %s", endpoint, string(body))
			} else {
				log.Printf("Successfully fetched and parsed %s data for ID: %d", endpoint, id)
			}
			requestCount++
			if requestCount%5 == 0 {
				log.Printf("Sleeping for 1 second to avoid hitting rate limits")
				time.Sleep(1 * time.Second)
			}
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
