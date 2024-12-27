package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"dimensi/db-aggregator/pkg/fetcher"
	"dimensi/db-aggregator/pkg/jikan"
	"dimensi/db-aggregator/pkg/ratelimiter"
)

type Client struct {
	httpClient  *http.Client
	rateLimiter *ratelimiter.RateLimiter
	config      fetcher.Config
	baseURL     string
}

func NewClient(httpClient *http.Client, rateLimiter *ratelimiter.RateLimiter) *Client {
	return &Client{
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
		config:      fetcher.DefaultConfig(),
		baseURL:     "https://api.jikan.moe/v4/anime",
	}
}

func (c *Client) FetchAnimeData(malID int) (jikan.Data, bool) {
	var jikanData jikan.Data
	jikanData.MyAnimeListID = malID

	// Получаем эпизоды
	page := 1
	for {
		url := fmt.Sprintf("%s/%d/episodes?page=%d", c.baseURL, malID, page)
		body, err := fetcher.FetchWithRetry(c.httpClient, url, c.rateLimiter, c.config)
		if err != nil {
			return jikanData, false
		}

		var response struct {
			Data       []jikan.Episode `json:"data"`
			Pagination struct {
				HasNextPage bool `json:"has_next_page"`
			} `json:"pagination"`
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return jikanData, false
		}

		jikanData.Episodes = append(jikanData.Episodes, response.Data...)

		if !response.Pagination.HasNextPage {
			break
		}
		page++
	}

	return jikanData, true
}
