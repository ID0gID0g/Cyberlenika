package converter

import (
	"Cyberlenika/internal/core/services"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

var _ services.Converter = (*Converter)(nil)

func (c Converter) PDFtoTxt(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("Файл %s не существует", path)
	}

	file, reader, err := pdf.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	pages := reader.NumPage()
	var allPagesText strings.Builder

	for i := 1; i <= pages; i++ {
		page := reader.Page(i)
		content := page.Content()

		text := content.Text

		allPagesText.WriteString(fmt.Sprintf("\n--- Страница %d ---\n", i))
		for _, row := range text {
			allPagesText.WriteString(row.S)
		}
	}

	return allPagesText.String()
}
