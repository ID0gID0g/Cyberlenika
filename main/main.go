package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/joho/godotenv"

	// Предполагаемые импорты ваших сервисов
	"Cyberlenika/internal/core/services"
	"Cyberlenika/internal/realization/converter"
	"Cyberlenika/internal/realization/finder"
	"Cyberlenika/internal/realization/narrator"
)

// --- Глобальные переменные UI и данных ---
var pdfPreviewLabel *widget.Label
var annotationLabel *widget.Label
var summaryLabel *widget.Label

// !!! СДЕЛАНО ГЛОБАЛЬНЫМ ДЛЯ ОБНОВЛЕНИЯ ПОИСКОМ !!!
var articleListContent *fyne.Container
var currentArticles []ArticleData // Для хранения текущего набора данных

// ArticleData --- Структура данных для статьи ---
type ArticleData struct {
	Finder    services.Finder
	Converter services.Converter
	Narrator  services.Narrator

	Title      string
	Annotation string
	Link       string
	Summary    string
	PDFContent string
}

// --- Вспомогательные функции ---

func createWrappingLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

// --- Кастомный виджет ClickableTextCard (Fyne 2.x Compliant) ---

type ClickableTextCard struct {
	widget.BaseWidget
	label    *widget.Label
	OnTapped func()
}

func NewClickableTextCard(text string, tapped func()) *ClickableTextCard {
	c := &ClickableTextCard{
		label:    createWrappingLabel(text),
		OnTapped: tapped,
	}
	c.label.TextStyle = fyne.TextStyle{Bold: true}

	c.BaseWidget.ExtendBaseWidget(c)
	return c
}

func (c *ClickableTextCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewPadded(c.label))
}

func (c *ClickableTextCard) Tapped(*fyne.PointEvent) {
	if c.OnTapped != nil {
		c.OnTapped()
	}
}

func (c *ClickableTextCard) MinSize() fyne.Size {
	padding := theme.Padding() * 2
	minSize := c.label.MinSize()
	return fyne.NewSize(minSize.Width+padding, minSize.Height+padding)
}

// --- Модифицированная функция обновления контента (Асинхронная) ---

func updateContent(data ArticleData, win fyne.Window) {
	// 1. Немедленно обновляем UI со статусом загрузки
	pdfPreviewLabel.SetText("Предпоказ PDF: " + data.Title + "\n\n" + data.PDFContent)
	annotationLabel.SetText("Аннотация:\n\n" + data.Annotation)

	summaryLabel.SetText("Пересказ:\n\nЗагрузка и обработка данных... Пожалуйста, подождите.")

	pdfPreviewLabel.Refresh()
	annotationLabel.Refresh()
	summaryLabel.Refresh()

	// 2. Запускаем длительные операции в новой горутине
	go func() {
		summaryText := "Ошибка: Неизвестная ошибка."

		path, err := data.Finder.Download(data.Link)
		if err == nil {
			text := data.Converter.PDFtoTxt(path)
			summaryText, err = data.Narrator.Retell(text)
			if err != nil {
				summaryText = "Ошибка Retell: " + err.Error()
			}
		} else {
			summaryText = "Ошибка Download: " + err.Error()
		}

		// 3. БЕЗОПАСНОЕ ОБНОВЛЕНИЕ UI (возвращаемся в UI-поток)
		summaryLabel.SetText("Пересказ:\n\n" + summaryText)
		win.Canvas().Refresh(summaryLabel)
	}()
}

// --- Логика Поиска и Обновления Списка ---

// rebuildArticleList перестраивает содержимое контейнера articleListContent
func rebuildArticleList(articles []ArticleData, win fyne.Window) {
	// Безопасно выполняем в главном потоке
	// NOTE: В Fyne, если вы не уверены, лучше всегда использовать CallOnMainThread/RunOnMainThread
	// или Canvas().Refresh(). В данном случае Canvas().Refresh() - самый простой путь.

	// Очищаем текущий список
	articleListContent.RemoveAll()

	// Добавляем новые элементы
	for i := range articles {
		// Проверяем, что статья не пустая, если вы используете массив (Articles[100])
		if articles[i].Title != "" {
			// Передаем window в callback
			articleListContent.Add(createArticleButton(articles[i], updateContent, win))
		}
	}

	// Обновляем сам контейнер списка и родительский ScrollContainer
	articleListContent.Refresh()
}

// processSearch выполняет поиск по введенному запросу
func processSearch(query string, win fyne.Window, newFinder services.Finder, newConverter services.Converter, newNarrator services.Narrator) {

	// 1. Немедленно показываем статус загрузки
	rebuildArticleList([]ArticleData{
		{Title: fmt.Sprintf("Поиск: '%s'. Загрузка результатов...", query)},
	}, win)

	// 2. Запускаем асинхронный поиск
	go func() {
		// Используем сервис Finder для поиска
		result, err := newFinder.Search(query)

		// Имитация задержки (на случай очень быстрого ответа)
		// time.Sleep(1 * time.Second)

		var newArticleData []ArticleData

		if err != nil || result.Found == 0 {
			newArticleData = []ArticleData{
				{Title: fmt.Sprintf("Статей не найдено или произошла ошибка: %v", err)},
			}
		} else {
			// Конвертируем результаты в ArticleData
			for _, article := range result.Articles {
				newArticleData = append(newArticleData, ArticleData{
					Finder: newFinder, Converter: newConverter, Narrator: newNarrator,
					Title:      article.Name,
					Annotation: article.Annotation,
					Link:       article.Link,
					Summary:    "...",
					PDFContent: "...",
				})
			}
		}

		// 3. Обновляем UI в главном потоке
		// Мы используем Canvas().Refresh() для сигнализации обновления
		currentArticles = newArticleData
		rebuildArticleList(currentArticles, win)
		win.Canvas().Refresh(articleListContent) // Обновляем ScrollContainer
	}()
}

// --- Main ---

func createArticleButton(data ArticleData, updater func(ArticleData, fyne.Window), win fyne.Window) fyne.CanvasObject {
	return NewClickableTextCard(data.Title, func() {
		updater(data, win)
	})
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Ошибка загрузки файла .env")
	}

	a := app.New()
	w := a.NewWindow("Cyberleninka")

	// Инициализация сервисов
	var newFinder services.Finder = finder.NewFinder()
	var newConverter services.Converter = converter.NewConverter()
	var newNarrator services.Narrator = narrator.NewNarrator()

	// Первичный поиск (может быть пустым)
	result, _ := newFinder.Search("")
	if result.Found == 0 {
		fmt.Println("На данную тему не найдено статей")
	}

	// --- Инициализация тестовых данных ---
	for _, article := range result.Articles {
		currentArticles = append(currentArticles, ArticleData{
			Finder:     newFinder,
			Converter:  newConverter,
			Narrator:   newNarrator,
			Title:      article.Name,
			Annotation: article.Annotation,
			Link:       article.Link,
			Summary:    "...",
			PDFContent: "...",
		})
	}
	if len(currentArticles) == 0 {
		currentArticles = append(currentArticles, ArticleData{
			Finder: newFinder, Converter: newConverter, Narrator: newNarrator,
			Title:      "Заглушка: Нажмите Enter в поле поиска, чтобы инициировать реальный поиск.",
			Annotation: "На данный момент список пуст.", Link: "https://test.link", Summary: "...", PDFContent: "...",
		})
	}

	// --- 1. Строка Поиска ---
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск по Cyberleninka.ru")

	// *** ЛОГИКА ON SUBMITTED ДЛЯ ПОИСКА ***
	searchEntry.OnSubmitted = func(query string) {
		fmt.Printf("Пользователь инициировал поиск: %s\n", query)
		// Запускаем процесс поиска, передавая все необходимые зависимости и окно 'w'
		processSearch(query, w, newFinder, newConverter, newNarrator)
	}

	searchBar := container.NewMax(searchEntry)

	// --- 2. Инициализация обновляемых виджетов ---
	pdfPreviewLabel = createWrappingLabel("Нажмите на статью, чтобы увидеть предпросмотр PDF")
	annotationLabel = createWrappingLabel("Нажмите на статью, чтобы увидеть аннотацию")
	summaryLabel = createWrappingLabel("Нажмите на статью, чтобы увидеть пересказ")

	pdfPreview := container.NewScroll(pdfPreviewLabel)
	annotationArea := container.NewScroll(annotationLabel)
	summaryArea := container.NewScroll(summaryLabel)

	// --- 3. Список Статей (1) ---
	articleListContent = container.NewVBox() // Глобальная переменная

	// Первичное заполнение списка
	rebuildArticleList(currentArticles, w)

	articlesScroll := container.NewScroll(articleListContent)
	articlesScroll.SetMinSize(fyne.NewSize(250, 0))

	// --- 4. Компоновка ---
	rightStack := container.NewVSplit(annotationArea, summaryArea)
	rightStack.SetOffset(0.5)

	mainRightArea := container.NewHSplit(pdfPreview, rightStack)
	mainRightArea.SetOffset(0.65)

	contentSplit := container.NewHSplit(articlesScroll, mainRightArea)
	contentSplit.SetOffset(0.25)

	mainLayout := container.NewBorder(
		container.NewVBox(searchBar, widget.NewSeparator()),
		nil,
		nil,
		nil,
		contentSplit,
	)

	w.SetContent(mainLayout)
	w.Resize(fyne.NewSize(1200, 800))
	w.ShowAndRun()
}
