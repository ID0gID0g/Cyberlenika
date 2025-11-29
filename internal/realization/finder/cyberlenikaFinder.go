package finder

import (
	"Cyberlenika/internal/core/models"
	"Cyberlenika/internal/core/services"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	BaseUrl  = "https://cyberleninka.ru"
	Mode     = "articles"
	Download = "C:/Users/ivano/Downloads/Cyberleninka"
)

var (
	Count  = 10
	Offset = 0
)

type Finder struct {
	client *resty.Client
}

// NewFinder Создание экземпляра клиента
func NewFinder() *Finder {
	return &Finder{
		client: resty.New(),
	}
}

// Проверка на соответствие интерфейсу
var _ services.Finder = (*Finder)(nil)

// Search Поиск запрашиваемых данных
func (f *Finder) Search(query string) (models.CyberlenikaSearchResponse, error) {
	var body = models.CyberlenikaRequestBody{
		Mode: Mode,
		Q:    query,
		Size: Count,
		From: Offset,
	}

	jsonData, _ := json.Marshal(body)

	r, err := f.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(jsonData).
		Post(BaseUrl + "/api/search")

	if err != nil {
		return models.CyberlenikaSearchResponse{}, err
	} else if r.StatusCode() != 200 {
		return models.CyberlenikaSearchResponse{}, err
	}

	var result models.CyberlenikaSearchResponse
	if e := json.Unmarshal([]byte(r.String()), &result); e != nil {
		return models.CyberlenikaSearchResponse{}, e
	}

	return result, nil
}

func (f *Finder) Download(url string) (string, error) {
	name := strings.TrimPrefix(url, "/article/n/")

	out := fmt.Sprintf("%s/PDF", Download)

	path := fmt.Sprintf("%s/%s.pdf", out, name)

	_, err := f.client.R().
		SetOutput(path).
		Get(BaseUrl + url + "/pdf")

	if err != nil {
		return "", fmt.Errorf("ошибка при скачивании PDF: %w", err)
	}

	return path, nil
}
