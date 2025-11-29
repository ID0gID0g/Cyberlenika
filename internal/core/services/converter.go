package services

type Converter interface {
	PDFtoTxt(text string) string
}
