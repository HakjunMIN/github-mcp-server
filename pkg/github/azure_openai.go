package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

type AzureOpenAIConfig struct {
	Endpoint   string
	Deployment string
	APIVersion string
	APIKey     string
}

func AzureOpenAIConfigFromEnv() (AzureOpenAIConfig, error) {
	cfg := AzureOpenAIConfig{
		Endpoint:   strings.TrimSpace(os.Getenv("AZURE_OPENAI_ENDPOINT")),
		Deployment: strings.TrimSpace(os.Getenv("AZURE_OPENAI_DEPLOYMENT")),
		APIVersion: strings.TrimSpace(os.Getenv("AZURE_OPENAI_API_VERSION")),
		APIKey:     strings.TrimSpace(os.Getenv("AZURE_OPENAI_API_KEY")),
	}

	var missing []string
	if cfg.Endpoint == "" {
		missing = append(missing, "AZURE_OPENAI_ENDPOINT")
	}
	if cfg.Deployment == "" {
		missing = append(missing, "AZURE_OPENAI_DEPLOYMENT")
	}
	if cfg.APIVersion == "" {
		missing = append(missing, "AZURE_OPENAI_API_VERSION")
	}
	if cfg.APIKey == "" {
		missing = append(missing, "AZURE_OPENAI_API_KEY")
	}
	if len(missing) > 0 {
		return AzureOpenAIConfig{}, fmt.Errorf("missing Azure OpenAI configuration: %s", strings.Join(missing, ", "))
	}

	if _, err := url.Parse(cfg.Endpoint); err != nil {
		return AzureOpenAIConfig{}, fmt.Errorf("invalid AZURE_OPENAI_ENDPOINT: %w", err)
	}

	return cfg, nil
}

type azureOpenAIChatCompletionRequest struct {
	Messages    []azureOpenAIChatMessage `json:"messages"`
	Temperature *float64                 `json:"temperature,omitempty"`
}

type azureOpenAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type azureOpenAIChatCompletionResponse struct {
	Choices []struct {
		Message azureOpenAIChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func AzureOpenAIChatCompletion(ctx context.Context, cfg AzureOpenAIConfig, systemPrompt, userPrompt string) (string, error) {
	endpoint, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid Azure OpenAI endpoint: %w", err)
	}

	endpoint.Path = path.Join(endpoint.Path, "/openai/deployments/", cfg.Deployment, "/chat/completions")
	q := endpoint.Query()
	q.Set("api-version", cfg.APIVersion)
	endpoint.RawQuery = q.Encode()

	temperature := 0.2
	payload := azureOpenAIChatCompletionRequest{
		Messages: []azureOpenAIChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: &temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Azure OpenAI request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create Azure OpenAI request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", cfg.APIKey)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("azure openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Azure OpenAI response: %w", err)
	}

	var decoded azureOpenAIChatCompletionResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return "", fmt.Errorf("failed to decode Azure OpenAI response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(decoded.ErrorMessage())
		if msg == "" {
			msg = strings.TrimSpace(string(respBody))
		}
		return "", fmt.Errorf("azure openai returned %s: %s", resp.Status, msg)
	}

	if len(decoded.Choices) == 0 {
		return "", errors.New("azure openai returned no choices")
	}

	content := strings.TrimSpace(decoded.Choices[0].Message.Content)
	if content == "" {
		return "", errors.New("azure openai returned empty content")
	}

	return content, nil
}

func (r azureOpenAIChatCompletionResponse) ErrorMessage() string {
	if r.Error == nil {
		return ""
	}
	return r.Error.Message
}
