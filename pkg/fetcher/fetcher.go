package fetcher

import (
	"dimensi/db-aggregator/pkg/ratelimiter"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Config struct {
	MaxRetries      int
	RetryDelay      time.Duration
	RetryMultiplier int
}

func DefaultConfig() Config {
	return Config{
		MaxRetries:      3,
		RetryDelay:      time.Second * 10,
		RetryMultiplier: 2,
	}
}

func FetchWithRetry(client *http.Client, url string, rateLimiter *ratelimiter.RateLimiter, config Config) ([]byte, error) {
	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		log.Printf("Fetching URL: %s (attempt %d/%d)", url, attempt, config.MaxRetries)

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
			if attempt < config.MaxRetries {
				delay := config.RetryDelay * time.Duration(attempt*config.RetryMultiplier)
				log.Printf("Rate limit exceeded, waiting %v before retry...", delay)
				time.Sleep(delay)
				continue
			}
			return nil, fmt.Errorf("rate limit exceeded after %d attempts", config.MaxRetries)
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		log.Printf("Successfully fetched URL: %s", url)
		return body, nil
	}

	return nil, fmt.Errorf("max retries reached")
}
