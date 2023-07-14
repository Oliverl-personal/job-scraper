package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// Models
const (
	TextAda001Model           = "text-ada-001"
	TextBabbage001Model       = "text-babbage-001"
	TextCurie001Model         = "text-curie-001"
	TextDavinci001Model       = "text-davinci-001"
	TextDavinci002Model       = "text-davinci-002"
	TextDavinci003Model       = "text-davinci-003"
	AdaModel                  = "ada"
	BabbageModel              = "babbage"
	CurieModel                = "curie"
	DavinciModel              = "davinci"
	GPT3Dot5TurboModel        = "gpt-3.5-turbo"
	GPT3Dot5Turbo16KModel     = "gpt-3.5-turbo-16k"
	GPT3Dot5Turbo0613Model    = "gpt-3.5-turbo-0613"
	GPT3Dot5Turbo16k0613Model = "gpt-3.5-turbo-16k-0613"
	GPT4Model                 = "gpt-4"
	GPT4_0613Model            = "gpt-4-0613"
	GPT4_32KModel             = "gpt-4-32k"
	GPT4_32K_0613Model        = "gpt-4-32k-0613"
	// note GPT3Dot5Turbo0613Model will be outdated after Sept 2023
	DefaultModel = GPT3Dot5Turbo0613Model
)

const (
	TextModerationLatest = "text-moderation-latest"
	TextModerationStable = "text-moderation-stable"
)

const (
	defaultBaseURL        = "https://api.openai.com/v1"
	defaultUserAgent      = "go-openai"
	defaultTimeoutSeconds = 30
)

// ChatCompletionRequestMessage is a message to use as the context for the chat completion API
type ChatCompletionRequestMessage struct {
	// Role is the role is the role of the the message. Can be "system", "user", or "assistant"
	Role string `json:"role"`
	// Content is the content of the message
	Content string `json:"content"`
}

type ChatCompletionFunctionParameter struct {
	// The type model will generate for function calling, this value should always be "object"
	Type string `json:"type"`
	// Properties model should return
	Properties map[string]map[string]interface{} `json:"properties"`
	// Required properties that model should return in its response
	Required []string `json:"required"`
}

type ChatCompletionFunction struct {
	// Name of the function called
	Name string `json:"name"`
	// Description of what the function does, used by the model to determine when to call the function
	Description string `json:"description"`
	// Parameters that the function accepts, described as a JSON object
	Parameters ChatCompletionFunctionParameter `json:"parameters"`
}

// ChatCompletionRequest is a request for the chat completion API
type ChatCompletionRequest struct {
	// List of functions the models may generate json inputs for
	Functions []ChatCompletionFunction `json:"functions,omitempty"`
	// Controls how the model responds to function calls
	FunctionCall string `json:"function_call,omitempty"`
	// Model is the name of the model to use. If not specified, will default to defaultModel
	Model string `json:"model"`
	// Messages is a list of messages to use as the context for the chat completion.
	Messages []ChatCompletionRequestMessage `json:"messages"`
	// What sampling temperature to use, between 0 and 2. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic
	Temperature *float32 `json:"temperature,omitempty"`
	// An alternative to sampling with temperature, called nucleus sampling, where the model considers the results of the tokens with top_p probability mass. So 0.1 means only the tokens comprising the top 10% probability mass are considered.
	TopP float32 `json:"top_p,omitempty"`
	// Number of responses to generate
	N int `json:"n,omitempty"`
	// Whether or not to stream responses back as they are generated
	Stream bool `json:"stream,omitempty"`
	// Up to 4 sequences where the API will stop generating further tokens.
	Stop []string `json:"stop,omitempty"`
	// MaxTokens is the maximum number of tokens to return.
	MaxTokens int `json:"max_tokens,omitempty"`
	// (-2, 2) Penalize tokens that haven't appeared yet in the history.
	PresencePenalty float32 `json:"presence_penalty,omitempty"`
	// (-2, 2) Penalize tokens that appear too frequently in the history.
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`
	// Modify the probability of specific tokens appearing in the completion.
	LogitBias map[string]float32 `json:"logit_bias,omitempty"`
	// Can be used to identify an end-user
	User string `json:"user,omitempty"`
}

type ChatCompletionFunctionRequest struct {

	// Model is the name of the model to use. If not specified, will default to defaultModel
	Model string `json:"model"`
	// Messages is a list of messages to use as the context for the chat completion.
	Messages []ChatCompletionRequestMessage `json:"messages"`
	// What sampling temperature to use, between 0 and 2. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic
	Temperature *float32 `json:"temperature,omitempty"`
	// An alternative to sampling with temperature, called nucleus sampling, where the model considers the results of the tokens with top_p probability mass. So 0.1 means only the tokens comprising the top 10% probability mass are considered.
	TopP float32 `json:"top_p,omitempty"`
	// Number of responses to generate
	N int `json:"n,omitempty"`
	// Whether or not to stream responses back as they are generated
	Stream bool `json:"stream,omitempty"`
	// Up to 4 sequences where the API will stop generating further tokens.
	Stop []string `json:"stop,omitempty"`
	// MaxTokens is the maximum number of tokens to return.
	MaxTokens int `json:"max_tokens,omitempty"`
	// (-2, 2) Penalize tokens that haven't appeared yet in the history.
	PresencePenalty float32 `json:"presence_penalty,omitempty"`
	// (-2, 2) Penalize tokens that appear too frequently in the history.
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`
	// Modify the probability of specific tokens appearing in the completion.
	LogitBias map[string]float32 `json:"logit_bias,omitempty"`
	// Can be used to identify an end-user
	User string `json:"user,omitempty"`
}

type FunctionCallResponse struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatCompletionResponseMessage is a message returned in the response to the Chat Completions API
type ChatCompletionResponseMessage struct {
	Role         string               `json:"role"`
	Content      string               `json:"content"`
	FunctionCall FunctionCallResponse `json:"function_call"`
}

// ChatCompletionsResponseUsage is the object that returns how many tokens the completion's request used
type ChatCompletionsResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponseChoice is one of the choices returned in the response to the Chat Completions API
type ChatCompletionResponseChoice struct {
	Index        int                           `json:"index"`
	FinishReason string                        `json:"finish_reason"`
	Message      ChatCompletionResponseMessage `json:"message"`
}

// ChatCompletionResponse is the full response from a request to the Chat Completions API
type ChatCompletionResponse struct {
	ID      string                         `json:"id"`
	Object  string                         `json:"object"`
	Created int                            `json:"created"`
	Model   string                         `json:"model"`
	Choices []ChatCompletionResponseChoice `json:"choices"`
	Usage   ChatCompletionsResponseUsage   `json:"usage"`
}

// APIError represents an error that occured on an API
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Type       string `json:"type"`
}

// APIErrorResponse is the full error response that has been returned by an API.
type APIErrorResponse struct {
	Error APIError `json:"error"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("[%d:%s] %s", e.StatusCode, e.Type, e.Message)
}

type Client interface {
	// Takes list of messages and feeds them into a GPT Model for an response
	ChatCompletion(ctx context.Context, request ChatCompletionRequest) (*ChatCompletionResponse, error)
	// Takes a predefined json formatted objects containing arguments, the GPT model will take it to and generate a response in the json format.
	// ChatCompletionFunctionCall(ctx context.Context, request ChatCompletionFunctionRequest) (*ChatCompletionResponse, error)
}

type client struct {
	baseURL      string
	apiKey       string
	userAgent    string
	httpClient   *http.Client
	defaultModel string
	idOrg        string
}

func NewClient(apiKey string, options ...ClientOptions) Client {
	httpClient := &http.Client{
		Timeout: time.Duration(defaultTimeoutSeconds * time.Second),
	}

	c := &client{
		baseURL:      defaultBaseURL,
		apiKey:       apiKey,
		userAgent:    defaultUserAgent,
		httpClient:   httpClient,
		defaultModel: DefaultModel,
		idOrg:        "",
	}

	for _, o := range options {
		o(c)
	}

	return c
}

func (c *client) ChatCompletion(ctx context.Context, request ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if request.Model == "" {
		request.Model = DefaultModel
	}
	request.Stream = false

	req, err := c.newRequest(ctx, "POST", "/chat/completions", request)
	if err != nil {
		return nil, err
	}

	resp, err := c.performRequest(req)
	if err != nil {
		return nil, err
	}

	output := new(ChatCompletionResponse)
	if err := getResponseObject(resp, output); err != nil {
		return nil, err
	}

	return output, nil
}

func getResponseObject(rsp *http.Response, v interface{}) error {
	defer rsp.Body.Close()
	if err := json.NewDecoder(rsp.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid json response: %w", err)
	}
	return nil
}

func (c *client) newRequest(ctx context.Context, method, path string, payload interface{}) (*http.Request, error) {
	bodyReader, err := jsonBodyReader(payload)
	if err != nil {
		return nil, err
	}
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	if len(c.idOrg) > 0 {
		req.Header.Set("OpenAI-Organization", c.idOrg)
	}
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	return req, nil
}

func jsonBodyReader(body interface{}) (io.Reader, error) {
	if body == nil {
		return bytes.NewBuffer(nil), nil
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed encoding json: %w", err)
	}
	return bytes.NewBuffer(raw), nil
}

func (c *client) performRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if err := checkForSuccess(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// returns an error if this response includes an error.
func checkForSuccess(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read from body: %w", err)
	}
	var result APIErrorResponse
	if err := json.Unmarshal(data, &result); err != nil {
		// if we can't decode the json error then create an unexpected error
		apiError := APIError{
			StatusCode: resp.StatusCode,
			Type:       "Unexpected",
			Message:    string(data),
		}
		return apiError
	}
	result.Error.StatusCode = resp.StatusCode
	return result.Error
}
