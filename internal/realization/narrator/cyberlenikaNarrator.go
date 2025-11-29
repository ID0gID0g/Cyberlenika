package narrator

import (
	"Cyberlenika/internal/core/services"
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Narrator struct {
	key   string
	model string
}

func NewNarrator() *Narrator {
	return &Narrator{
		key:   "AIzaSyCL-VCOP2iDmvKqPVOlgeGe65rqtJahyYE",
		model: "gemini-2.5-flash",
	}
}

var _ services.Narrator = (*Narrator)(nil)

func (n Narrator) Retell(text string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(n.key))
	if err != nil {
		return "", err
	}
	defer func(client *genai.Client) {
		err := client.Close()
		if err != nil {
			log.Fatalf("Failed to close client: %v", err)
		}
	}(client)

	model := client.GenerativeModel(n.model)
	r, err := model.GenerateContent(ctx, genai.Text(fmt.Sprintf("Кратко перескажи текст: %s", text)))
	if err != nil {
		return "", err
	}

	return string(r.Candidates[0].Content.Parts[0].(genai.Text)), nil
}
