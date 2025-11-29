package services

type Narrator interface {
	Retell(text string) (string, error)
}
