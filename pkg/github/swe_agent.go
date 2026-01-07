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

// SWEAgentConfig contains the configuration for calling the SWE Agent REST API.
type SWEAgentConfig struct {
	Endpoint           string // SWE Agent REST API endpoint (e.g., http://localhost:8000)
	AzureOpenAIAPIBase string // Azure OpenAI API base URL
	AzureOpenAIAPIKey  string // Azure OpenAI API key
	AzureOpenAIModel   string // Azure OpenAI model name (e.g., azure/gpt-5-chat)
	AzureAPIVersion    string // Azure OpenAI API version
	GitHubToken        string // GitHub token for SWE Agent
}

// SWEAgentConfigFromEnv reads SWE Agent configuration from environment variables.
func SWEAgentConfigFromEnv() (SWEAgentConfig, error) {
	cfg := SWEAgentConfig{
		Endpoint:           strings.TrimSpace(os.Getenv("SWE_AGENT_ENDPOINT")),
		AzureOpenAIAPIBase: strings.TrimSpace(os.Getenv("AZURE_OPENAI_API_BASE")),
		AzureOpenAIAPIKey:  strings.TrimSpace(os.Getenv("AZURE_OPENAI_API_KEY")),
		AzureOpenAIModel:   strings.TrimSpace(os.Getenv("AZURE_OPENAI_MODEL")),
		AzureAPIVersion:    strings.TrimSpace(os.Getenv("AZURE_OPENAI_API_VERSION")),
		GitHubToken:        strings.TrimSpace(os.Getenv("GITHUB_TOKEN")),
	}

	// Set default endpoint if not provided
	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://localhost:8000"
	}

	var missing []string
	if cfg.AzureOpenAIAPIBase == "" {
		missing = append(missing, "AZURE_OPENAI_API_BASE")
	}
	if cfg.AzureOpenAIAPIKey == "" {
		missing = append(missing, "AZURE_OPENAI_API_KEY")
	}
	if cfg.AzureOpenAIModel == "" {
		missing = append(missing, "AZURE_OPENAI_MODEL")
	}
	if cfg.AzureAPIVersion == "" {
		missing = append(missing, "AZURE_OPENAI_API_VERSION")
	}
	if cfg.GitHubToken == "" {
		missing = append(missing, "GITHUB_TOKEN")
	}
	if len(missing) > 0 {
		return SWEAgentConfig{}, fmt.Errorf("missing SWE Agent configuration: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

// SWEAgentRunRequest represents the request body for SWE Agent /run endpoint.
type SWEAgentRunRequest struct {
	Agent            SWEAgentModel            `json:"agent"`
	ProblemStatement SWEAgentProblemStatement `json:"problem_statement"`
	Env              SWEAgentEnv              `json:"env"`
	Actions          SWEAgentActions          `json:"actions"`
	EnvVars          map[string]string        `json:"env_vars"`
}

// SWEAgentModel represents the model configuration for SWE Agent.
type SWEAgentModel struct {
	Model SWEAgentModelConfig `json:"model"`
}

// SWEAgentModelConfig represents the model details.
type SWEAgentModelConfig struct {
	Name       string `json:"name"`
	APIBase    string `json:"api_base"`
	APIVersion string `json:"api_version"`
	APIKey     string `json:"api_key"`
}

// SWEAgentProblemStatement represents the problem statement for SWE Agent.
type SWEAgentProblemStatement struct {
	Type      string `json:"type"`
	GitHubURL string `json:"github_url"`
}

// SWEAgentEnv represents the environment configuration for SWE Agent.
type SWEAgentEnv struct {
	Repo SWEAgentRepo `json:"repo"`
}

// SWEAgentRepo represents the repository configuration.
type SWEAgentRepo struct {
	GitHubURL string `json:"github_url"`
}

// SWEAgentActions represents the actions configuration for SWE Agent.
type SWEAgentActions struct {
	OpenPR bool `json:"open_pr"`
}

// SWEAgentRunResponse represents the response from SWE Agent /run endpoint.
type SWEAgentRunResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// CallSWEAgent sends a request to the SWE Agent REST API to process a GitHub issue.
func CallSWEAgent(ctx context.Context, cfg SWEAgentConfig, owner, repo string, issueNumber int) (*SWEAgentRunResponse, error) {
	issueURL := fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, issueNumber)
	repoURL := fmt.Sprintf("https://github.com/%s/%s", owner, repo)

	reqBody := SWEAgentRunRequest{
		Agent: SWEAgentModel{
			Model: SWEAgentModelConfig{
				Name:       cfg.AzureOpenAIModel,
				APIBase:    cfg.AzureOpenAIAPIBase,
				APIVersion: cfg.AzureAPIVersion,
				APIKey:     cfg.AzureOpenAIAPIKey,
			},
		},
		ProblemStatement: SWEAgentProblemStatement{
			Type:      "github",
			GitHubURL: issueURL,
		},
		Env: SWEAgentEnv{
			Repo: SWEAgentRepo{
				GitHubURL: repoURL,
			},
		},
		Actions: SWEAgentActions{
			OpenPR: true,
		},
		EnvVars: map[string]string{
			"GITHUB_TOKEN": cfg.GitHubToken,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SWE Agent request: %w", err)
	}

	endpoint := strings.TrimSuffix(cfg.Endpoint, "/") + "/run"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create SWE Agent request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SWE Agent request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read SWE Agent response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("SWE Agent returned %s: %s", resp.Status, string(respBody))
	}

	var sweResp SWEAgentRunResponse
	if err := json.Unmarshal(respBody, &sweResp); err != nil {
		return nil, fmt.Errorf("failed to decode SWE Agent response: %w", err)
	}

	return &sweResp, nil
}

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
