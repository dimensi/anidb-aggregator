package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"dimensi/db-aggregator/pkg/fetcher"
	"dimensi/db-aggregator/pkg/ratelimiter"
	"dimensi/db-aggregator/pkg/shikimori"
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
		baseURL:     "https://shikimori.one/api/animes/",
	}
}

func (c *Client) FetchAnimeData(malID int) (shikimori.Data, bool) {
	var shikiData shikimori.Data

	// Получаем основные данные
	url := fmt.Sprintf("%s%d", c.baseURL, malID)
	body, err := fetcher.FetchWithRetry(c.httpClient, url, c.rateLimiter, c.config)
	if err != nil {
		return shikiData, false
	}

	if err := json.Unmarshal(body, &shikiData.ShikimoriData); err != nil {
		return shikiData, false
	}

	// Получаем роли
	url = fmt.Sprintf("%s%d/roles", c.baseURL, malID)
	body, err = fetcher.FetchWithRetry(c.httpClient, url, c.rateLimiter, c.config)
	if err == nil {
		json.Unmarshal(body, &shikiData.Roles)
	}

	// Получаем похожие аниме
	url = fmt.Sprintf("%s%d/similar", c.baseURL, malID)
	body, err = fetcher.FetchWithRetry(c.httpClient, url, c.rateLimiter, c.config)
	if err == nil {
		json.Unmarshal(body, &shikiData.Similar)
	}

	return shikiData, true
}
