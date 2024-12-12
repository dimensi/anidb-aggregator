package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Episode struct {
	ID      string            `json:"id"`
	Number  string            `json:"number"`
	AirDate string            `json:"airDate"`
	Titles  map[string]string `json:"titles"`
	Summary string            `json:"summary"`
	Rating  string            `json:"rating"`
}

type AnimeOutput struct {
	ID            int       `json:"id"`
	MyAnimeListID int       `json:"myAnimeListId"`
	Episodes      []Episode `json:"episodes"`
}

type AnimeInput struct {
	AniDbID       int `json:"aniDbId"`
	MyAnimeListID int `json:"myAnimeListId"`
}

type XMLAnime struct {
	Episodes []XMLEpisode `xml:"episodes>episode"`
}

type XMLEpisode struct {
	ID   string `xml:"id,attr"`
	Epno struct {
		Type string `xml:"type,attr"`
		Text string `xml:",chardata"`
	} `xml:"epno"`
	AirDate string     `xml:"airdate"`
	Titles  []XMLTitle `xml:"title"`
	Summary string     `xml:"summary"`
	Rating  string     `xml:"rating"`
}

type XMLTitle struct {
	Lang  string `xml:"lang,attr"`
	Value string `xml:",chardata"`
}

type XMLError struct {
	Code    string `xml:"code,attr"`
	Message string `xml:",chardata"`
}

type Config struct {
	skipExisting bool
	inputFile    string
	outputFile   string
}

func fetchWithRetry(client *http.Client, url string) ([]byte, error) {
	maxRetries := 3 // Максимальное количество попыток

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Fetching URL: %s (attempt %d/%d)", url, attempt, maxRetries)

		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch URL %s: %v", url, err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}

		// Проверяем на наличие ошибки бана
		var xmlError XMLError
		if err := xml.Unmarshal(body, &xmlError); err == nil && xmlError.Code == "500" && xmlError.Message == "banned" {
			if attempt < maxRetries {
				log.Printf("Received banned error, waiting 60 seconds before retry...")
				time.Sleep(300 * time.Second)
				continue
			}
			return nil, fmt.Errorf("banned error after %d attempts", maxRetries)
		}

		log.Printf("Successfully fetched URL: %s", url)
		return body, nil
	}

	return nil, fmt.Errorf("max retries reached")
}

func parseAnimeEpisodes(data []byte) ([]Episode, error) {
	var xmlAnime XMLAnime
	if err := xml.Unmarshal(data, &xmlAnime); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %v", err)
	}

	if len(xmlAnime.Episodes) == 0 {
		return nil, fmt.Errorf("no episodes found in XML data")
	}

	var episodes []Episode
	for _, xmlEp := range xmlAnime.Episodes {
		if xmlEp.Epno.Type != "1" { // Only process episodes with type="1"
			continue
		}

		titles := make(map[string]string)
		for _, title := range xmlEp.Titles {
			titles[title.Lang] = strings.TrimSpace(title.Value)
		}

		episodes = append(episodes, Episode{
			ID:      xmlEp.ID,
			Number:  xmlEp.Epno.Text,
			AirDate: xmlEp.AirDate,
			Titles:  titles,
			Summary: strings.TrimSpace(xmlEp.Summary),
			Rating:  xmlEp.Rating,
		})
	}

	if len(episodes) == 0 {
		return nil, fmt.Errorf("no valid episodes found after processing")
	}

	return episodes, nil
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
		outputFile:   "anidb-db.jsonl",
	}

	flag.BoolVar(&config.skipExisting, "skip-existing", false, "Skip processing if entry already exists in output file")
	flag.StringVar(&config.inputFile, "input", "anime365-db.jsonl", "Input file path")
	flag.StringVar(&config.outputFile, "output", "anidb-db.jsonl", "Output file path")
	flag.Parse()

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

	baseURL := "http://api.anidb.net:9001/httpapi?request=anime&client=ichimetvos&clientver=2&protover=1&aid="
	rateLimiter := time.Tick(5 * time.Second) // 1 запрос в 5 секунд

	log.Printf("Opening input file: %s", config.inputFile)
	input, err := os.Open(config.inputFile)
	if err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer input.Close()

	log.Printf("Creating output file: %s", config.outputFile)
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
			if _, exists := existing[animeInput.AniDbID]; exists {
				log.Printf("Skipping existing entry for AniDB ID: %d", animeInput.AniDbID)
				delete(existing, animeInput.AniDbID) // Удаляем обработанную запись из памяти
				continue
			}
		}

		if animeInput.AniDbID == 0 {
			emptyData := AnimeOutput{
				ID:            animeInput.AniDbID,
				MyAnimeListID: animeInput.MyAnimeListID,
				Episodes:      []Episode{},
			}
			jsonLine, _ := json.Marshal(emptyData)
			output.WriteString(string(jsonLine) + "\n")
			log.Printf("Skipped processing for zero AniDB ID, MyAnimeListID: %d", animeInput.MyAnimeListID)
			continue
		}

		<-rateLimiter

		url := fmt.Sprintf("%s%d", baseURL, animeInput.AniDbID)
		body, err := fetchWithRetry(client, url)
		if err != nil {
			log.Printf("Failed to fetch AniDB data for ID %d: %v", animeInput.AniDbID, err)
			emptyData := AnimeOutput{
				ID:            animeInput.AniDbID,
				MyAnimeListID: animeInput.MyAnimeListID,
				Episodes:      []Episode{},
			}
			jsonLine, _ := json.Marshal(emptyData)
			output.WriteString(string(jsonLine) + "\n")
			continue
		}

		episodes, err := parseAnimeEpisodes(body)
		if err != nil {
			log.Printf("Failed to parse episodes for AniDB ID %d: %v", animeInput.AniDbID, err)
			debugFileName := fmt.Sprintf("debug_response_%d.xml", animeInput.AniDbID)
			if err = os.WriteFile(debugFileName, body, 0644); err != nil {
				log.Printf("Failed to save debug response: %v", err)
			} else {
				log.Printf("Saved problematic response to %s", debugFileName)
			}

			emptyData := AnimeOutput{
				ID:            animeInput.AniDbID,
				MyAnimeListID: animeInput.MyAnimeListID,
				Episodes:      []Episode{},
			}
			jsonLine, _ := json.Marshal(emptyData)
			output.WriteString(string(jsonLine) + "\n")
			continue
		}

		outputData := AnimeOutput{
			ID:            animeInput.AniDbID,
			MyAnimeListID: animeInput.MyAnimeListID,
			Episodes:      episodes,
		}

		jsonLine, _ := json.Marshal(outputData)
		output.WriteString(string(jsonLine) + "\n")
		log.Printf("Successfully processed AniDB ID: %d with %d episodes", animeInput.AniDbID, len(episodes))
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	log.Printf("All data successfully saved to %s", config.outputFile)
}
