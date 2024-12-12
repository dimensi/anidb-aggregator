package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type SeriesResponse struct {
	Data []map[string]interface{} `json:"data"`
}

func main() {
	// Определяем флаги командной строки
	initialOffset := flag.Int("offset", 0, "Starting offset for fetching data")
	batchSize := flag.Int("limit", 500, "Number of items to fetch per request")
	flag.Parse()

	// Base URL to fetch the data from
	baseURL := "https://smotret-anime.online/api/series/?limit=%d&offset=%d"

	// Open the output file
	file, err := os.Create("anime365-db.jsonl")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	offset := *initialOffset
	for {
		// Construct the URL with the current offset and batch size
		url := fmt.Sprintf(baseURL, *batchSize, offset)
		fmt.Printf("Fetching data from: %s\n", url)

		// Perform the HTTP GET request
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to fetch data: %v", err)
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read response body: %v", err)
		}

		// Parse the JSON response
		var series SeriesResponse
		if err := json.Unmarshal(body, &series); err != nil {
			log.Fatalf("Failed to parse JSON: %v", err)
		}

		// Break the loop if no data is returned
		if len(series.Data) == 0 {
			fmt.Println("No more data to fetch.")
			break
		}

		// Write each item as a JSON line
		for _, item := range series.Data {
			jsonLine, err := json.Marshal(item)
			if err != nil {
				log.Printf("Failed to marshal item: %v", err)
				continue
			}
			_, err = file.WriteString(string(jsonLine) + "\n")
			if err != nil {
				log.Printf("Failed to write to file: %v", err)
			}
		}

		// Increment the offset for the next batch
		offset += *batchSize
	}

	fmt.Println("All data successfully saved to anime365-db.jsonl")
}
