package tagger

import (
	"context"
	"fmt"
	"scraper-app/apps/openai"
	"scraper-app/utils"
)

var openAIApiKey string

const OpenAIURL = "https://api.openai.com/v1/chat/completions"

func InitTagger() {
	openAIApiKey = utils.GetEnv("OPENAI_KEY")
}

func TagJobDescription(jd string) (string, error) {
	if jd == "" {
		return "", fmt.Errorf("unable to create tags for empty job description")
	}
	prompt := "Job Posting: " + jd

	client := openai.NewClient(openAIApiKey)
	ctx := context.Background()
	param := openai.ChatCompletionFunctionParameter{
		Type: "object",
		Properties: map[string]map[string]interface{}{
			"coop":      {"type": "string", "description": "Determine if the job posting is a internship or coop position, outputs should be [Yes/No/Maybe]"},
			"tech":      {"type": "string", "description": "Determine the software technologies mentioned in the job posting. Each technology should be comma separated"},
			"languages": {"type": "string", "description": "Determine the programming languages mentioned in the job posting. Each language should be comma separated"},
			"skills":    {"type": "string", "description": "Determine the skills mentioned in the job posting. Each skill should be comma separated"},
		},
		Required: []string{"coop", "tech", "languages", "skills"},
	}
	function := openai.ChatCompletionFunction{
		Name:        "tag_job_posting",
		Description: "Given a job posting, provide tags",
		Parameters:  param,
	}
	msgs := make([]openai.ChatCompletionRequestMessage, 1)
	msgs[0] = openai.ChatCompletionRequestMessage{
		Role:    "user",
		Content: prompt,
	}
	resp, err := client.ChatCompletion(ctx, openai.ChatCompletionRequest{
		Functions:    []openai.ChatCompletionFunction{function},
		FunctionCall: "auto",
		Messages:     msgs,
		MaxTokens:    3000,
	})
	if err != nil {
		utils.FatalError(fmt.Errorf("%v, ", err))
	}
	return resp.Choices[0].Message.FunctionCall.Arguments, nil
}
