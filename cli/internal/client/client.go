package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/strata/strata/cli/internal/config"
)

type StrataClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func New() *StrataClient {
	return &StrataClient{
		baseURL: config.GetGatewayURL(),
		token:   config.GetAccessToken(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func NewWithURL(baseURL, token string) *StrataClient {
	return &StrataClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *StrataClient) SetToken(token string) {
	c.token = token
}

func (c *StrataClient) DoRequest(method, path string, body, result interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return resp, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
	}
	return resp, nil
}

func (c *StrataClient) Health() (map[string]interface{}, error) {
	var result map[string]interface{}
	_, err := c.DoRequest("GET", "/v1/health", nil, &result)
	return result, err
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         map[string]interface{} `json:"user"`
}

func (c *StrataClient) Login(email, password string) (*LoginResponse, error) {
	body := map[string]string{"email": email, "password": password}
	var result LoginResponse
	_, err := c.DoRequest("POST", "/v1/auth/login", body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *StrataClient) Register(email, password, role, orgID string) error {
	body := map[string]string{
		"email":    email,
		"password": password,
		"role":     role,
		"org_id":   orgID,
	}
	_, err := c.DoRequest("POST", "/v1/auth/register", body, nil)
	return err
}

func (c *StrataClient) GetMe() (map[string]interface{}, error) {
	var result map[string]interface{}
	_, err := c.DoRequest("GET", "/v1/auth/me", nil, &result)
	return result, err
}

type Function struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (c *StrataClient) ListFunctions() ([]Function, error) {
	var result []Function
	_, err := c.DoRequest("GET", "/v1/functions", nil, &result)
	return result, err
}

func (c *StrataClient) DeployFunction(name, description, code string) error {
	body := map[string]string{
		"name":        name,
		"description": description,
		"code":        code,
	}
	_, err := c.DoRequest("POST", "/v1/functions", body, nil)
	return err
}

func (c *StrataClient) InvokeFunction(name string, payload interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	_, err := c.DoRequest("POST", "/v1/functions/"+name+"/invoke", payload, &result)
	return result, err
}

func (c *StrataClient) DeleteFunction(name string) error {
	_, err := c.DoRequest("DELETE", "/v1/functions/"+name, nil, nil)
	return err
}

type Bucket struct {
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func (c *StrataClient) ListBuckets() ([]Bucket, error) {
	var result []Bucket
	_, err := c.DoRequest("GET", "/v1/storage/buckets", nil, &result)
	return result, err
}

func (c *StrataClient) CreateBucket(name string) error {
	body := map[string]string{"name": name}
	_, err := c.DoRequest("POST", "/v1/storage/buckets", body, nil)
	return err
}

func (c *StrataClient) DeleteBucket(name string) error {
	_, err := c.DoRequest("DELETE", "/v1/storage/buckets/"+name, nil, nil)
	return err
}

func (c *StrataClient) UploadFile(bucket, filepath string, data io.Reader) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/storage/buckets/%s/upload", c.baseURL, bucket)
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "multipart/form-data")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *StrataClient) DownloadFile(bucket, filepath string, w io.Writer) error {
	url := fmt.Sprintf("%s/v1/storage/buckets/%s/download/%s", c.baseURL, bucket, filepath)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)
	return err
}

type AICollection struct {
	Name      string `json:"name"`
	DocCount  int    `json:"doc_count"`
	CreatedAt string `json:"created_at"`
}

func (c *StrataClient) ListAICollections() ([]AICollection, error) {
	var result []AICollection
	_, err := c.DoRequest("GET", "/v1/ai/collections", nil, &result)
	return result, err
}

func (c *StrataClient) CreateAICollection(name string) error {
	body := map[string]string{"name": name}
	_, err := c.DoRequest("POST", "/v1/ai/collections", body, nil)
	return err
}

func (c *StrataClient) DeleteAICollection(name string) error {
	_, err := c.DoRequest("DELETE", "/v1/ai/collections/"+name, nil, nil)
	return err
}

type AIDocument struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt string `json:"created_at"`
}

func (c *StrataClient) ListAIDocuments(collection string) ([]AIDocument, error) {
	var result []AIDocument
	_, err := c.DoRequest("GET", "/v1/ai/collections/"+collection+"/documents", nil, &result)
	return result, err
}

func (c *StrataClient) AddAIDocument(collection, content string, metadata map[string]interface{}) error {
	body := map[string]interface{}{
		"content":  content,
		"metadata": metadata,
	}
	_, err := c.DoRequest("POST", "/v1/ai/collections/"+collection+"/documents", body, nil)
	return err
}

func (c *StrataClient) DeleteAIDocument(collection, docID string) error {
	_, err := c.DoRequest("DELETE", "/v1/ai/collections/"+collection+"/documents/"+docID, nil, nil)
	return err
}

type AISearchResult struct {
	Content   string  `json:"content"`
	Score     float64 `json:"score"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type AIProvider struct {
	ID           string `json:"id"`
	Provider     string `json:"provider"`
	Enabled      bool   `json:"enabled"`
	IsPrimary    bool   `json:"is_primary"`
	DefaultModel string `json:"default_model"`
	BaseURL      string `json:"base_url"`
	CreatedAt    string `json:"created_at"`
}

func (c *StrataClient) ListAIProviders() ([]AIProvider, error) {
	var result []AIProvider
	_, err := c.DoRequest("GET", "/v1/ai/providers", nil, &result)
	return result, err
}

func (c *StrataClient) CreateAIProvider(provider, apiKey, baseURL, defaultModel string) error {
	body := map[string]interface{}{
		"provider":      provider,
		"api_key":       apiKey,
		"base_url":      baseURL,
		"default_model": defaultModel,
	}
	_, err := c.DoRequest("POST", "/v1/ai/providers", body, nil)
	return err
}

func (c *StrataClient) DeleteAIProvider(provider string) error {
	_, err := c.DoRequest("DELETE", "/v1/ai/providers/"+provider, nil, nil)
	return err
}

func (c *StrataClient) TestAIProvider(provider string) (map[string]interface{}, error) {
	var result map[string]interface{}
	_, err := c.DoRequest("POST", "/v1/ai/providers/"+provider+"/test", nil, &result)
	return result, err
}

type AIEmbeddingResult struct {
	Model      string      `json:"model"`
	Embeddings [][]float32 `json:"embeddings"`
	Usage      *struct {
		PromptTokens int `json:"prompt_tokens"`
	} `json:"usage,omitempty"`
}

func (c *StrataClient) GenerateEmbeddings(input []string, provider, model string) (*AIEmbeddingResult, error) {
	body := map[string]interface{}{
		"input":    input,
		"provider": provider,
		"model":    model,
	}
	var result AIEmbeddingResult
	_, err := c.DoRequest("POST", "/v1/ai/embeddings", body, &result)
	return &result, err
}

type HubAgent struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	SystemPrompt string `json:"system_prompt"`
	Model        string `json:"model"`
	Enabled      bool   `json:"enabled"`
}

func (c *StrataClient) ListHubAgents() ([]HubAgent, error) {
	var result []HubAgent
	_, err := c.DoRequest("GET", "/v1/ai/hub/agents", nil, &result)
	return result, err
}

func (c *StrataClient) ChatWithAgent(agentID, message string) (map[string]interface{}, error) {
	body := map[string]string{"message": message}
	var result map[string]interface{}
	_, err := c.DoRequest("POST", "/v1/ai/hub/agents/"+agentID+"/chat", body, &result)
	return result, err
}

type HubWorkflow struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Nodes   int    `json:"nodes,omitempty"`
	Edges   int    `json:"edges,omitempty"`
}

func (c *StrataClient) ListHubWorkflows() ([]HubWorkflow, error) {
	var result []HubWorkflow
	_, err := c.DoRequest("GET", "/v1/ai/hub/workflows", nil, &result)
	return result, err
}

func (c *StrataClient) ExecuteWorkflow(id string) (map[string]interface{}, error) {
	var result map[string]interface{}
	_, err := c.DoRequest("POST", "/v1/ai/hub/workflows/"+id+"/execute", nil, &result)
	return result, err
}

func (c *StrataClient) ListHubPrompts() ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	_, err := c.DoRequest("GET", "/v1/ai/hub/prompts", nil, &result)
	return result, err
}

func (c *StrataClient) GetUsageStats() ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	_, err := c.DoRequest("GET", "/v1/ai/usage?limit=50", nil, &result)
	return result, err
}

func (c *StrataClient) SearchAICollection(collection, query string, topK int) ([]AISearchResult, error) {
	body := map[string]interface{}{
		"query": query,
		"top_k": topK,
	}
	var result struct {
		Results []AISearchResult `json:"results"`
	}
	_, err := c.DoRequest("POST", "/v1/ai/collections/"+collection+"/search", body, &result)
	return result.Results, err
}

func (c *StrataClient) CheckServiceHealth(serviceURL string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", serviceURL+"/health", nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
