package tagger

import (
	"scraper-app/utils"
)

var openAIApiKey string

func InitTagger() {
	openAIApiKey = utils.GetEnv("OPENAI_KEY")
}

// Request
type TaggerRequest struct {
	Choices []struct {
		// TODO
	} `json:"choices"`
}

// Response Config
type TaggerResponse struct {
	Choices []struct {
		// TODO
	} `json:"choices"`
}

type Tagger struct {
	ApiURL string
	ApiKey string
}

// NewOpenAITagger creates a new instance of OpenAITagger
func NewTagger(apiURL string) *Tagger {
	return &Tagger{
		ApiURL: apiURL,
		ApiKey: openAIApiKey,
	}
}

// Tag performs the tagging operation using OpenAI API
func (t *Tagger) Tag(prompt string) (string, error) {
	// TODO
	return tag(prompt)
}

func tag(prompt string) (string, error) {
	// TODO
	return
}
