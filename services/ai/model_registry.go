package main

import "sort"

type ModelInfo struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

var modelRegistry = []ModelInfo{
	{ID: "gpt-4.1", Provider: "openai", Name: "GPT-4.1", Category: "chat"},
	{ID: "gpt-4o", Provider: "openai", Name: "GPT-4o", Category: "chat"},
	{ID: "gpt-4o-mini", Provider: "openai", Name: "GPT-4o Mini", Category: "chat"},
	{ID: "gpt-4-turbo", Provider: "openai", Name: "GPT-4 Turbo", Category: "chat"},
	{ID: "gpt-3.5-turbo", Provider: "openai", Name: "GPT-3.5 Turbo", Category: "chat"},
	{ID: "text-embedding-3-large", Provider: "openai", Name: "Text Embedding 3 Large", Category: "embedding"},
	{ID: "text-embedding-3-small", Provider: "openai", Name: "Text Embedding 3 Small", Category: "embedding"},

	{ID: "claude-sonnet-4", Provider: "anthropic", Name: "Claude Sonnet 4", Category: "chat"},
	{ID: "claude-3-5-sonnet", Provider: "anthropic", Name: "Claude 3.5 Sonnet", Category: "chat"},
	{ID: "claude-3-5-haiku", Provider: "anthropic", Name: "Claude 3.5 Haiku", Category: "chat"},
	{ID: "claude-3-opus", Provider: "anthropic", Name: "Claude 3 Opus", Category: "chat"},

	{ID: "gemini-2.5-pro", Provider: "gemini", Name: "Gemini 2.5 Pro", Category: "chat"},
	{ID: "gemini-2.5-flash", Provider: "gemini", Name: "Gemini 2.5 Flash", Category: "chat"},
	{ID: "gemini-2.0-flash", Provider: "gemini", Name: "Gemini 2.0 Flash", Category: "chat"},
	{ID: "text-embedding-004", Provider: "gemini", Name: "Text Embedding 004", Category: "embedding"},

	{ID: "llama-4", Provider: "groq", Name: "Llama 4", Category: "chat"},
	{ID: "llama-3.3-70b", Provider: "groq", Name: "Llama 3.3 70B", Category: "chat"},
	{ID: "mixtral-8x7b", Provider: "groq", Name: "Mixtral 8x7B", Category: "chat"},
	{ID: "deepseek-r1", Provider: "groq", Name: "DeepSeek R1", Category: "chat"},

	{ID: "openrouter/auto", Provider: "openrouter", Name: "OpenRouter Auto", Category: "chat"},
	{ID: "openrouter/gpt-4o", Provider: "openrouter", Name: "OpenRouter GPT-4o", Category: "chat"},
	{ID: "openrouter/claude-3.5-sonnet", Provider: "openrouter", Name: "OpenRouter Claude 3.5 Sonnet", Category: "chat"},

	{ID: "command-r-plus", Provider: "cohere", Name: "Command R+", Category: "chat"},
	{ID: "command-r", Provider: "cohere", Name: "Command R", Category: "chat"},
	{ID: "embed-english-v3.0", Provider: "cohere", Name: "Embed English v3", Category: "embedding"},

	{ID: "mistral-large", Provider: "mistral", Name: "Mistral Large", Category: "chat"},
	{ID: "mistral-medium", Provider: "mistral", Name: "Mistral Medium", Category: "chat"},
	{ID: "mistral-small", Provider: "mistral", Name: "Mistral Small", Category: "chat"},
	{ID: "mistral-embed", Provider: "mistral", Name: "Mistral Embed", Category: "embedding"},

	{ID: "deepseek-chat", Provider: "deepseek", Name: "DeepSeek Chat", Category: "chat"},
	{ID: "deepseek-reasoner", Provider: "deepseek", Name: "DeepSeek Reasoner", Category: "chat"},

	{ID: "ollama/llama3", Provider: "ollama", Name: "Ollama Llama 3", Category: "chat"},
	{ID: "ollama/mistral", Provider: "ollama", Name: "Ollama Mistral", Category: "chat"},
	{ID: "ollama/codellama", Provider: "ollama", Name: "Ollama CodeLlama", Category: "chat"},
	{ID: "ollama/nomic-embed-text", Provider: "ollama", Name: "Ollama Nomic Embed Text", Category: "embedding"},

	{ID: "lmstudio/local", Provider: "lmstudio", Name: "LM Studio Local", Category: "chat"},
	{ID: "localai/local", Provider: "localai", Name: "LocalAI", Category: "chat"},
	{ID: "vllm/local", Provider: "vllm", Name: "vLLM Server", Category: "chat"},
}

func getModelInfo(modelID string) *ModelInfo {
	for _, m := range modelRegistry {
		if m.ID == modelID {
			return &m
		}
	}
	return nil
}

func getModelsForProvider(provider string) []ModelInfo {
	var result []ModelInfo
	for _, m := range modelRegistry {
		if m.Provider == provider {
			result = append(result, m)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

func listAllModels() []ModelInfo {
	result := make([]ModelInfo, len(modelRegistry))
	copy(result, modelRegistry)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Provider != result[j].Provider {
			return result[i].Provider < result[j].Provider
		}
		return result[i].ID < result[j].ID
	})
	return result
}
