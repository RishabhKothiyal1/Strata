package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Provider interface
// ---------------------------------------------------------------------------

type ProviderAdapter interface {
	Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error)
	ChatStream(ctx context.Context, req *ChatRequest, apiKey string, baseURL string, onChunk func(StreamChunk)) error
	Embeddings(ctx context.Context, req *EmbeddingsRequest, apiKey string, baseURL string) (*EmbeddingsResponse, error)
	Models(ctx context.Context, apiKey string, baseURL string) ([]string, error)
	ValidateKey(ctx context.Context, apiKey string, baseURL string) error
}

type ProviderConfig struct {
	APIKey  string
	BaseURL string
}

// ---------------------------------------------------------------------------
// Shared types
// ---------------------------------------------------------------------------

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type ChatResponse struct {
	ID        string    `json:"id"`
	Model     string    `json:"model"`
	Choices   []Choice  `json:"choices"`
	Usage     *Usage    `json:"usage,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Choice struct {
	Index   int         `json:"index"`
	Message ChatMessage `json:"message"`
	Finish  string      `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type EmbeddingsRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbeddingsResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float32 `json:"embeddings"`
	Usage      *Usage      `json:"usage,omitempty"`
}

type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

// ---------------------------------------------------------------------------
// Provider registry
// ---------------------------------------------------------------------------

type providerFactory func() ProviderAdapter

var providerRegistry = map[string]providerFactory{
	"openai":     func() ProviderAdapter { return &OpenAIAdapter{} },
	"anthropic":  func() ProviderAdapter { return &AnthropicAdapter{} },
	"gemini":     func() ProviderAdapter { return &GeminiAdapter{} },
	"groq":       func() ProviderAdapter { return &GroqAdapter{} },
	"openrouter": func() ProviderAdapter { return &OpenRouterAdapter{} },
	"cohere":     func() ProviderAdapter { return &CohereAdapter{} },
	"mistral":    func() ProviderAdapter { return &MistralAdapter{} },
	"deepseek":   func() ProviderAdapter { return &DeepSeekAdapter{} },
	"ollama":     func() ProviderAdapter { return &OllamaAdapter{} },
	"lmstudio":   func() ProviderAdapter { return &LMStudioAdapter{} },
	"localai":    func() ProviderAdapter { return &LocalAIAdapter{} },
	"vllm":       func() ProviderAdapter { return &VLLMAdapter{} },
}

func getProvider(name string) (ProviderAdapter, error) {
	factory, ok := providerRegistry[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return factory(), nil
}

func listProviders() []string {
	names := make([]string, 0, len(providerRegistry))
	for name := range providerRegistry {
		names = append(names, name)
	}
	return names
}

// ---------------------------------------------------------------------------
// Provider default base URLs
// ---------------------------------------------------------------------------

var defaultBaseURLs = map[string]string{
	"openai":     "https://api.openai.com/v1",
	"anthropic":  "https://api.anthropic.com/v1",
	"gemini":     "https://generativelanguage.googleapis.com/v1beta",
	"groq":       "https://api.groq.com/openai/v1",
	"openrouter": "https://openrouter.ai/api/v1",
	"cohere":     "https://api.cohere.ai/v1",
	"mistral":    "https://api.mistral.ai/v1",
	"deepseek":   "https://api.deepseek.com/v1",
	"ollama":     "http://localhost:11434",
	"lmstudio":   "http://localhost:1234/v1",
	"localai":    "http://localhost:8080/v1",
	"vllm":       "http://localhost:8000/v1",
}

// ---------------------------------------------------------------------------
// OpenAI adapter (also serves as base for Groq, OpenRouter, DeepSeek)
// ---------------------------------------------------------------------------

type OpenAIAdapter struct{}

func (a *OpenAIAdapter) chatURL(baseURL string) string {
	return baseURL + "/chat/completions"
}

func (a *OpenAIAdapter) embedURL(baseURL string) string {
	return baseURL + "/embeddings"
}

func (a *OpenAIAdapter) modelsURL(baseURL string) string {
	return baseURL + "/models"
}

type openAIChatReq struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type openAIChatResp struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Model   string       `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   *openAIUsage `json:"usage,omitempty"`
}

type openAIChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (a *OpenAIAdapter) Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["openai"]
	}
	body := openAIChatReq{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.chatURL(baseURL), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, parseProviderError(resp.StatusCode, bodyData)
	}

	var oaResp openAIChatResp
	if err := json.Unmarshal(bodyData, &oaResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	result := &ChatResponse{
		ID:    oaResp.ID,
		Model: oaResp.Model,
		Choices: make([]Choice, len(oaResp.Choices)),
	}
	if oaResp.Usage != nil {
		result.Usage = &Usage{
			PromptTokens:     oaResp.Usage.PromptTokens,
			CompletionTokens: oaResp.Usage.CompletionTokens,
			TotalTokens:      oaResp.Usage.TotalTokens,
		}
	}
	for i, c := range oaResp.Choices {
		result.Choices[i] = Choice{
			Index:   c.Index,
			Message: c.Message,
			Finish:  c.FinishReason,
		}
	}
	return result, nil
}

func (a *OpenAIAdapter) Embeddings(ctx context.Context, req *EmbeddingsRequest, apiKey string, baseURL string) (*EmbeddingsResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["openai"]
	}
	body := map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.embedURL(baseURL), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Model  string `json:"model"`
		Data   []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Usage *openAIUsage `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	er := &EmbeddingsResponse{Model: result.Model}
	for _, d := range result.Data {
		er.Embeddings = append(er.Embeddings, d.Embedding)
	}
	if result.Usage != nil {
		er.Usage = &Usage{PromptTokens: result.Usage.PromptTokens, TotalTokens: result.Usage.TotalTokens}
	}
	return er, nil
}

func (a *OpenAIAdapter) Models(ctx context.Context, apiKey string, baseURL string) ([]string, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["openai"]
	}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", a.modelsURL(baseURL), nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	models := make([]string, len(result.Data))
	for i, m := range result.Data {
		models[i] = m.ID
	}
	return models, nil
}

func (a *OpenAIAdapter) ChatStream(ctx context.Context, req *ChatRequest, apiKey string, baseURL string, onChunk func(StreamChunk)) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["openai"]
	}
	body := openAIChatReq{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.chatURL(baseURL), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyData, _ := io.ReadAll(resp.Body)
		return parseProviderError(resp.StatusCode, bodyData)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			onChunk(StreamChunk{Content: "", Done: true})
			return nil
		}
		var streamResp struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue
		}
		if len(streamResp.Choices) > 0 {
			onChunk(StreamChunk{Content: streamResp.Choices[0].Delta.Content})
		}
	}
	return scanner.Err()
}

func (a *OpenAIAdapter) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["openai"]
	}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", a.modelsURL(baseURL), nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return parseProviderError(0, nil)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return parseProviderError(resp.StatusCode, body)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Groq (OpenAI-compatible)
// ---------------------------------------------------------------------------

type GroqAdapter struct {
	OpenAIAdapter
}

func (a *GroqAdapter) Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["groq"]
	}
	return a.OpenAIAdapter.Chat(ctx, req, apiKey, baseURL)
}

func (a *GroqAdapter) Embeddings(ctx context.Context, req *EmbeddingsRequest, apiKey string, baseURL string) (*EmbeddingsResponse, error) {
	return nil, fmt.Errorf("embeddings not supported by Groq")
}

func (a *GroqAdapter) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["groq"]
	}
	return a.OpenAIAdapter.ValidateKey(ctx, apiKey, baseURL)
}

// ---------------------------------------------------------------------------
// OpenRouter (OpenAI-compatible)
// ---------------------------------------------------------------------------

type OpenRouterAdapter struct {
	OpenAIAdapter
}

func (a *OpenRouterAdapter) Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["openrouter"]
	}
	return a.OpenAIAdapter.Chat(ctx, req, apiKey, baseURL)
}

func (a *OpenRouterAdapter) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["openrouter"]
	}
	return a.OpenAIAdapter.ValidateKey(ctx, apiKey, baseURL)
}

// ---------------------------------------------------------------------------
// DeepSeek (OpenAI-compatible)
// ---------------------------------------------------------------------------

type DeepSeekAdapter struct {
	OpenAIAdapter
}

func (a *DeepSeekAdapter) Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["deepseek"]
	}
	return a.OpenAIAdapter.Chat(ctx, req, apiKey, baseURL)
}

func (a *DeepSeekAdapter) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["deepseek"]
	}
	return a.OpenAIAdapter.ValidateKey(ctx, apiKey, baseURL)
}

// ---------------------------------------------------------------------------
// Ollama (OpenAI-compatible)
// ---------------------------------------------------------------------------

type OllamaAdapter struct {
	OpenAIAdapter
}

func (a *OllamaAdapter) Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["ollama"]
	}
	return a.OpenAIAdapter.Chat(ctx, req, apiKey, baseURL)
}

func (a *OllamaAdapter) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["ollama"]
	}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ollama unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("ollama: HTTP %d", resp.StatusCode)
	}
	return nil
}

// ---------------------------------------------------------------------------
// LM Studio (OpenAI-compatible)
// ---------------------------------------------------------------------------

type LMStudioAdapter struct {
	OpenAIAdapter
}

func (a *LMStudioAdapter) Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["lmstudio"]
	}
	return a.OpenAIAdapter.Chat(ctx, req, "", baseURL)
}

func (a *LMStudioAdapter) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["lmstudio"]
	}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("lm studio unavailable: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

// ---------------------------------------------------------------------------
// LocalAI (OpenAI-compatible)
// ---------------------------------------------------------------------------

type LocalAIAdapter struct {
	OpenAIAdapter
}

func (a *LocalAIAdapter) Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["localai"]
	}
	return a.OpenAIAdapter.Chat(ctx, req, "", baseURL)
}

func (a *LocalAIAdapter) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["localai"]
	}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("localai unavailable: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

// ---------------------------------------------------------------------------
// vLLM (OpenAI-compatible)
// ---------------------------------------------------------------------------

type VLLMAdapter struct {
	OpenAIAdapter
}

func (a *VLLMAdapter) Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["vllm"]
	}
	return a.OpenAIAdapter.Chat(ctx, req, apiKey, baseURL)
}

func (a *VLLMAdapter) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["vllm"]
	}
	return a.OpenAIAdapter.ValidateKey(ctx, apiKey, baseURL)
}

// ---------------------------------------------------------------------------
// Anthropic adapter
// ---------------------------------------------------------------------------

type AnthropicAdapter struct{}

func (a *AnthropicAdapter) Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["anthropic"]
	}

	systemMsg := ""
	messages := make([]ChatMessage, 0)
	for _, m := range req.Messages {
		if m.Role == "system" {
			systemMsg = m.Content
		} else {
			messages = append(messages, m)
		}
	}

	type anthropicContent struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	type anthropicMsg struct {
		Role    string             `json:"role"`
		Content []anthropicContent `json:"content"`
	}
	body := map[string]interface{}{
		"model":      req.Model,
		"max_tokens": req.MaxTokens,
		"messages":   messages,
	}
	if systemMsg != "" {
		body["system"] = systemMsg
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/messages", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, parseProviderError(resp.StatusCode, bodyData)
	}

	var anthroResp struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		StopReason string `json:"stop_reason"`
	}
	if err := json.Unmarshal(bodyData, &anthroResp); err != nil {
		return nil, err
	}

	content := ""
	for _, c := range anthroResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &ChatResponse{
		ID:    anthroResp.ID,
		Model: anthroResp.Model,
		Choices: []Choice{{
			Index:   0,
			Message: ChatMessage{Role: "assistant", Content: content},
			Finish:  anthroResp.StopReason,
		}},
		Usage: &Usage{
			PromptTokens:     anthroResp.Usage.InputTokens,
			CompletionTokens: anthroResp.Usage.OutputTokens,
			TotalTokens:      anthroResp.Usage.InputTokens + anthroResp.Usage.OutputTokens,
		},
	}, nil
}

func (a *AnthropicAdapter) ChatStream(ctx context.Context, req *ChatRequest, apiKey string, baseURL string, onChunk func(StreamChunk)) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["anthropic"]
	}

	systemMsg := ""
	messages := make([]ChatMessage, 0)
	for _, m := range req.Messages {
		if m.Role == "system" {
			systemMsg = m.Content
		} else {
			messages = append(messages, m)
		}
	}

	body := map[string]interface{}{
		"model":      req.Model,
		"max_tokens": req.MaxTokens,
		"messages":   messages,
		"stream":     true,
	}
	if systemMsg != "" {
		body["system"] = systemMsg
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/messages", bytes.NewReader(data))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyData, _ := io.ReadAll(resp.Body)
		return parseProviderError(resp.StatusCode, bodyData)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			onChunk(StreamChunk{Content: "", Done: true})
			return nil
		}
		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Text string `json:"text"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event.Type == "content_block_delta" {
			onChunk(StreamChunk{Content: event.Delta.Text})
		}
		if event.Type == "message_stop" {
			onChunk(StreamChunk{Content: "", Done: true})
			return nil
		}
	}
	return scanner.Err()
}

func (a *AnthropicAdapter) Embeddings(ctx context.Context, req *EmbeddingsRequest, apiKey string, baseURL string) (*EmbeddingsResponse, error) {
	return nil, fmt.Errorf("embeddings not directly supported by Anthropic API")
}

func (a *AnthropicAdapter) Models(ctx context.Context, apiKey string, baseURL string) ([]string, error) {
	return []string{"claude-sonnet-4", "claude-3-5-sonnet", "claude-3-5-haiku", "claude-3-opus"}, nil
}

func (a *AnthropicAdapter) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["anthropic"]
	}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("anthropic unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return parseProviderError(resp.StatusCode, body)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Gemini adapter
// ---------------------------------------------------------------------------

type GeminiAdapter struct{}

func (a *GeminiAdapter) chatURL(baseURL, apiKey, model string) string {
	return fmt.Sprintf("%s/models/%s:generateContent?key=%s", baseURL, model, apiKey)
}

func (a *GeminiAdapter) embedURL(baseURL, apiKey, model string) string {
	return fmt.Sprintf("%s/models/%s:embedContent?key=%s", baseURL, model, apiKey)
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiPart struct {
	Text string `json:"text"`
}

func (a *GeminiAdapter) Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["gemini"]
	}

	systemInstruction := ""
	contents := make([]geminiContent, 0)
	for _, m := range req.Messages {
		if m.Role == "system" {
			systemInstruction = m.Content
		} else {
			gRole := "user"
			if m.Role == "assistant" || m.Role == "model" {
				gRole = "model"
			}
			contents = append(contents, geminiContent{
				Role:  gRole,
				Parts: []geminiPart{{Text: m.Content}},
			})
		}
	}

	body := map[string]interface{}{
		"contents": contents,
	}
	if systemInstruction != "" {
		body["system_instruction"] = map[string]interface{}{
			"parts": []map[string]string{{"text": systemInstruction}},
		}
	}
	if req.Temperature > 0 {
		body["generationConfig"] = map[string]interface{}{
			"temperature": req.Temperature,
			"maxOutputTokens": req.MaxTokens,
		}
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := a.chatURL(baseURL, apiKey, req.Model)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, parseProviderError(resp.StatusCode, bodyData)
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := json.Unmarshal(bodyData, &geminiResp); err != nil {
		return nil, err
	}

	content := ""
	if len(geminiResp.Candidates) > 0 {
		for _, p := range geminiResp.Candidates[0].Content.Parts {
			content += p.Text
		}
	}

	return &ChatResponse{
		Model: req.Model,
		Choices: []Choice{{
			Index:   0,
			Message: ChatMessage{Role: "assistant", Content: content},
			Finish:  geminiResp.Candidates[0].FinishReason,
		}},
		Usage: &Usage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		},
	}, nil
}

func (a *GeminiAdapter) ChatStream(ctx context.Context, req *ChatRequest, apiKey string, baseURL string, onChunk func(StreamChunk)) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["gemini"]
	}

	systemInstruction := ""
	contents := make([]geminiContent, 0)
	for _, m := range req.Messages {
		if m.Role == "system" {
			systemInstruction = m.Content
		} else {
			gRole := "user"
			if m.Role == "assistant" || m.Role == "model" {
				gRole = "model"
			}
			contents = append(contents, geminiContent{
				Role:  gRole,
				Parts: []geminiPart{{Text: m.Content}},
			})
		}
	}

	body := map[string]interface{}{
		"contents": contents,
	}
	if systemInstruction != "" {
		body["system_instruction"] = map[string]interface{}{
			"parts": []map[string]string{{"text": systemInstruction}},
		}
	}
	if req.Temperature > 0 {
		body["generationConfig"] = map[string]interface{}{
			"temperature":    req.Temperature,
			"maxOutputTokens": req.MaxTokens,
		}
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s&alt=sse", baseURL, req.Model, apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyData, _ := io.ReadAll(resp.Body)
		return parseProviderError(resp.StatusCode, bodyData)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			onChunk(StreamChunk{Content: "", Done: true})
			return nil
		}
		var geminiResp struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
				FinishReason string `json:"finishReason"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal([]byte(data), &geminiResp); err != nil {
			continue
		}
		if len(geminiResp.Candidates) > 0 {
			for _, p := range geminiResp.Candidates[0].Content.Parts {
				onChunk(StreamChunk{Content: p.Text})
			}
			if geminiResp.Candidates[0].FinishReason != "" {
				onChunk(StreamChunk{Content: "", Done: true})
				return nil
			}
		}
	}
	return scanner.Err()
}

func (a *GeminiAdapter) Embeddings(ctx context.Context, req *EmbeddingsRequest, apiKey string, baseURL string) (*EmbeddingsResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["gemini"]
	}
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("no input provided")
	}

	type geminiEmbedReq struct {
		Model  string `json:"model"`
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	}

	embedReq := geminiEmbedReq{
		Model: req.Model,
	}
	embedReq.Content.Parts = []struct{ Text string }{{Text: req.Input[0]}}

	data, err := json.Marshal(embedReq)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.embedURL(baseURL, apiKey, req.Model), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var geminiResp struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, err
	}
	return &EmbeddingsResponse{
		Model:      req.Model,
		Embeddings: [][]float32{geminiResp.Embedding.Values},
	}, nil
}

func (a *GeminiAdapter) Models(ctx context.Context, apiKey string, baseURL string) ([]string, error) {
	return []string{"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash", "text-embedding-004"}, nil
}

func (a *GeminiAdapter) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["gemini"]
	}
	url := fmt.Sprintf("%s/models?key=%s", baseURL, apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("gemini unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("invalid gemini API key")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Cohere adapter
// ---------------------------------------------------------------------------

type CohereAdapter struct{}

type cohereChatReq struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type cohereChatResp struct {
	ID      string `json:"id"`
	Message struct {
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"message"`
	FinishReason string       `json:"finish_reason"`
	Usage        *openAIUsage `json:"usage,omitempty"`
}

func (a *CohereAdapter) Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["cohere"]
	}
	body := cohereChatReq{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyData, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, parseProviderError(resp.StatusCode, bodyData)
	}

	var cResp cohereChatResp
	if err := json.Unmarshal(bodyData, &cResp); err != nil {
		return nil, err
	}

	content := ""
	for _, c := range cResp.Message.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	result := &ChatResponse{
		ID:    cResp.ID,
		Model: req.Model,
		Choices: []Choice{{
			Index:   0,
			Message: ChatMessage{Role: "assistant", Content: content},
			Finish:  cResp.FinishReason,
		}},
	}
	if cResp.Usage != nil {
		result.Usage = &Usage{
			PromptTokens:     cResp.Usage.PromptTokens,
			CompletionTokens: cResp.Usage.CompletionTokens,
			TotalTokens:      cResp.Usage.TotalTokens,
		}
	}
	return result, nil
}

func (a *CohereAdapter) ChatStream(ctx context.Context, req *ChatRequest, apiKey string, baseURL string, onChunk func(StreamChunk)) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["cohere"]
	}
	body := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
		"stream":      true,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat", bytes.NewReader(data))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyData, _ := io.ReadAll(resp.Body)
		return parseProviderError(resp.StatusCode, bodyData)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			onChunk(StreamChunk{Content: "", Done: true})
			return nil
		}
		var event struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event.Type == "text" {
			onChunk(StreamChunk{Content: event.Text})
		}
		if event.Type == "end" {
			onChunk(StreamChunk{Content: "", Done: true})
			return nil
		}
	}
	return scanner.Err()
}

func (a *CohereAdapter) Embeddings(ctx context.Context, req *EmbeddingsRequest, apiKey string, baseURL string) (*EmbeddingsResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["cohere"]
	}
	body := map[string]interface{}{
		"model":  req.Model,
		"texts":  req.Input,
		"input_type": "search_document",
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/embed", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &EmbeddingsResponse{Model: req.Model, Embeddings: result.Embeddings}, nil
}

func (a *CohereAdapter) Models(ctx context.Context, apiKey string, baseURL string) ([]string, error) {
	return []string{"command-r-plus", "command-r", "embed-english-v3.0"}, nil
}

func (a *CohereAdapter) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["cohere"]
	}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("cohere unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("invalid cohere API key")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Mistral adapter
// ---------------------------------------------------------------------------

type MistralAdapter struct {
	OpenAIAdapter
}

func (a *MistralAdapter) Chat(ctx context.Context, req *ChatRequest, apiKey string, baseURL string) (*ChatResponse, error) {
	if baseURL == "" {
		baseURL = defaultBaseURLs["mistral"]
	}
	return a.OpenAIAdapter.Chat(ctx, req, apiKey, baseURL)
}

func (a *MistralAdapter) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURLs["mistral"]
	}
	return a.OpenAIAdapter.ValidateKey(ctx, apiKey, baseURL)
}

// ---------------------------------------------------------------------------
// Error handling
// ---------------------------------------------------------------------------

type ProviderError struct {
	StatusCode int    `json:"status_code"`
	Type       string `json:"type"`
	Message    string `json:"message"`
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func parseProviderError(statusCode int, body []byte) *ProviderError {
	err := &ProviderError{StatusCode: statusCode}
	if statusCode == 0 {
		err.Type = "connection_error"
		err.Message = "provider unreachable"
		return err
	}

	switch statusCode {
	case 401:
		err.Type = "invalid_api_key"
		err.Message = "Invalid API key"
	case 429:
		err.Type = "rate_limit_exceeded"
		err.Message = "Rate limit exceeded"
	case 402:
		err.Type = "quota_exceeded"
		err.Message = "Quota exceeded"
	case 408:
		err.Type = "network_timeout"
		err.Message = "Request timed out"
	case 502, 503:
		err.Type = "provider_unavailable"
		err.Message = "Provider temporarily unavailable"
	default:
		err.Type = "api_error"
		err.Message = fmt.Sprintf("HTTP %d", statusCode)
	}

	if len(body) > 0 {
		var parsed struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &parsed) == nil && parsed.Error.Message != "" {
			err.Message = parsed.Error.Message
			if parsed.Error.Type != "" {
				err.Type = parsed.Error.Type
			}
		}
	}
	return err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func providerRequest(ctx context.Context, method, url, apiKey string, body []byte) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, parseProviderError(resp.StatusCode, data)
	}
	return data, nil
}

// ---------------------------------------------------------------------------
// Fallback chain
// ---------------------------------------------------------------------------

type FallbackProvider struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

func chatWithFallback(ctx context.Context, req *ChatRequest, providers []FallbackProvider, getConfig func(provider string) (*ProviderConfig, error)) (*ChatResponse, string, error) {
	var lastErr error
	for _, fp := range providers {
		cfg, err := getConfig(fp.Provider)
		if err != nil {
			slog.Warn("fallback: provider config unavailable", "provider", fp.Provider, "error", err)
			lastErr = err
			continue
		}

		adapter, err := getProvider(fp.Provider)
		if err != nil {
			lastErr = err
			continue
		}

		chatReq := *req
		if fp.Model != "" {
			chatReq.Model = fp.Model
		}

		result, err := adapter.Chat(ctx, &chatReq, cfg.APIKey, cfg.BaseURL)
		if err != nil {
			slog.Warn("fallback: provider failed", "provider", fp.Provider, "error", err)
			lastErr = err
			continue
		}

		return result, fp.Provider, nil
	}
	return nil, "", fmt.Errorf("all providers failed: %w", lastErr)
}
