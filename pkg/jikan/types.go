package jikan

type Data struct {
	ID            int       `json:"id"`
	MyAnimeListID int       `json:"myAnimeListId"`
	Episodes      []Episode `json:"episodes"`
}

type Episode struct {
	MalID         int     `json:"mal_id"`
	Title         string  `json:"title"`
	TitleJapanese string  `json:"title_japanese"`
	TitleRomanji  string  `json:"title_romanji"`
	Aired         string  `json:"aired"`
	Score         float64 `json:"score"`
	Filler        bool    `json:"filler"`
	Recap         bool    `json:"recap"`
	ForumURL      string  `json:"forum_url"`
}
