package models

type CyberlenikaSearchResponse struct {
	Found    int `json:"found"`
	Articles []struct {
		Name        string      `json:"name"`
		Annotation  string      `json:"annotation"`
		Link        string      `json:"link"`
		Authors     []string    `json:"authors"`
		Year        int         `json:"year"`
		Journal     string      `json:"journal"`
		JournalLink string      `json:"journal_link"`
		Ocr         []string    `json:"ocr"`
		Catalogs    interface{} `json:"catalogs"`
	} `json:"articles"`
	AggTerm []struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"agg_term"`
	AggYear []struct {
		From  int    `json:"from"`
		To    int    `json:"to"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"agg_year"`
	AggJournal []struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"agg_journal"`
	AggAge []struct {
		From  int    `json:"from"`
		To    int    `json:"to"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"agg_age"`
	AggCat []struct {
		Id      int    `json:"id"`
		Acronym string `json:"acronym"`
		Count   int    `json:"count"`
	} `json:"agg_cat"`
}

type CyberlenikaRequestBody struct {
	Mode string `json:"mode"`
	Q    string `json:"q"`
	Size int    `json:"size"`
	From int    `json:"from"`
}
