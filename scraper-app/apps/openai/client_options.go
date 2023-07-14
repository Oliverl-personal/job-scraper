package openai

import "net/http"

type ClientOptions func(*client) error

func WithBaseURL(url string) ClientOptions {
	return func(c *client) error {
		c.baseURL = url
		return nil
	}
}

func WithUserAgent(agent string) ClientOptions {
	return func(c *client) error {
		c.userAgent = agent
		return nil
	}
}

func WithHTTPClient(httpClient *http.Client) ClientOptions {
	return func(c *client) error {
		c.httpClient = httpClient
		return nil
	}
}

func WithDefaultModel(model string) ClientOptions {
	return func(c *client) error {
		c.defaultModel = model
		return nil
	}
}

func WithIdOrg(id string) ClientOptions {
	return func(c *client) error {
		c.idOrg = id
		return nil
	}
}
