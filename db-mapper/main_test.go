package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dimensi/db-aggregator/pkg/anime365"
	"dimensi/db-aggregator/pkg/jikan"
	"dimensi/db-aggregator/pkg/shikimori"
)

func TestMapToResultAnime(t *testing.T) {
	// Загружаем тестовые данные
	anime365Data := loadAnime365Data(t)
	shikimoriData := loadShikimoriData(t)
	jikanData := loadJikanData(t)

	// Вызываем тестируемую функцию
	result := mapToResultAnime(anime365Data, shikimoriData, true, jikanData, true)

	// Проверяем результат
	assert.NotNil(t, result)

	// Проверяем основные поля
	assert.Equal(t, "Провожающая в последний путь Фрирен", result.Titles["ru"])
	assert.Equal(t, "Frieren: Beyond Journey's End", result.Titles["en"])
	assert.Equal(t, "葬送のフリーレン", result.Titles["ja"])

	// Проверяем описания
	assert.Len(t, result.Descriptions, 2)
	assert.Equal(t, "world-art.ru", result.Descriptions[0].Source)
	assert.Equal(t, "shikimori.me", result.Descriptions[1].Source)

	// Проверяем эпизоды
	assert.Equal(t, 28, result.NumberOfEpisodes)
	assert.Len(t, result.Episodes, 29)

	// Проверяем первый эпизод
	firstEp := result.Episodes[0]
	jsonData, _ := json.MarshalIndent(firstEp, "", "  ")
	fmt.Printf("First episode: %s\n", jsonData)

	assert.Equal(t, "The Journey's End", firstEp.Titles["en"])
	assert.Equal(t, "冒険の終わり", firstEp.Titles["ja"])
	assert.Equal(t, "Bouken no Owari", firstEp.Titles["romaji"])
	assert.Equal(t, "4.39", firstEp.Rating)

	// Проверяем жанры
	assert.Len(t, result.Genres, 4)
	assert.Equal(t, "Приключения", result.Genres[0].Title)
	assert.Equal(t, "Фэнтези", result.Genres[1].Title)
	assert.Equal(t, "Сёнен", result.Genres[2].Title)
	assert.Equal(t, "Драма", result.Genres[3].Title)
}

func loadAnime365Data(t *testing.T) anime365.Data {
	data, err := os.ReadFile("../test-data/anime365.json")
	require.NoError(t, err)

	var anime anime365.Data
	err = json.Unmarshal(data, &anime)
	require.NoError(t, err)

	return anime
}

func loadShikimoriData(t *testing.T) shikimori.Data {
	data, err := os.ReadFile("../test-data/shikimori.json")
	require.NoError(t, err)

	var anime shikimori.Data
	err = json.Unmarshal(data, &anime)
	require.NoError(t, err)

	return anime
}

func loadJikanData(t *testing.T) jikan.Data {
	data, err := os.ReadFile("../test-data/jikan.json")
	require.NoError(t, err)

	var episodes jikan.Data
	err = json.Unmarshal(data, &episodes)
	require.NoError(t, err)

	return episodes
}
