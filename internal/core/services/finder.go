package services

import "Cyberlenika/internal/core/models"

type Finder interface {
	Search(query string) (models.CyberlenikaSearchResponse, error)
	Download(url string) (string, error)
}
