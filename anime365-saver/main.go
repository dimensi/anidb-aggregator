package main

import (
	"encoding/json"
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
	// Base URL to fetch the data from
	baseURL := "https://smotret-anime.online/api/series/?limit=500&offset=%d"

	// Open the output file
	file, err := os.Create("output.jsonl")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	offset := 0
	for {
		// Construct the URL with the current offset
		url := fmt.Sprintf(baseURL, offset)
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
		offset += 500
	}

	fmt.Println("All data successfully saved to output.jsonl")
}
