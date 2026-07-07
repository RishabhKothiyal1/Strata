package main

import (
	"context"
	"database/sql"
	"time"
)

type UsageLog struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	LatencyMs        int64     `json:"latency_ms"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	EstimatedCost    float64   `json:"estimated_cost"`
	Success          bool      `json:"success"`
	Error            string    `json:"error,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type UsageStats struct {
	Provider         string  `json:"provider"`
	Model            string  `json:"model"`
	TotalRequests    int     `json:"total_requests"`
	SuccessfulReq    int     `json:"successful_requests"`
	FailedReq        int     `json:"failed_requests"`
	TotalTokens      int     `json:"total_tokens"`
	AvgLatencyMs     float64 `json:"avg_latency_ms"`
	EstimatedCost    float64 `json:"estimated_cost"`
}

var costPerToken = map[string]map[string]float64{
	"openai":    {"input": 0.0000025, "output": 0.000010},
	"anthropic": {"input": 0.000003, "output": 0.000015},
	"gemini":    {"input": 0.00000125, "output": 0.000005},
	"groq":      {"input": 0.00000059, "output": 0.00000079},
	"cohere":    {"input": 0.0000005, "output": 0.0000015},
	"mistral":   {"input": 0.000002, "output": 0.000006},
	"deepseek":  {"input": 0.00000014, "output": 0.00000028},
}

func estimateCost(provider string, promptTokens, completionTokens int) float64 {
	rates, ok := costPerToken[provider]
	if !ok {
		return 0
	}
	inputCost := float64(promptTokens) * rates["input"]
	outputCost := float64(completionTokens) * rates["output"]
	return inputCost + outputCost
}

func logUsage(ctx context.Context, db *sql.DB, userID, provider, model string, latencyMs int64, promptTokens, completionTokens int, success bool, errMsg string) {
	totalTokens := promptTokens + completionTokens
	cost := estimateCost(provider, promptTokens, completionTokens)

	_, err := db.ExecContext(ctx, `
		INSERT INTO ai_usage_logs (user_id, provider, model, latency_ms, prompt_tokens, completion_tokens, total_tokens, estimated_cost, success, error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, userID, provider, model, latencyMs, promptTokens, completionTokens, totalTokens, cost, success, errMsg, time.Now().UTC())

	if err != nil {
		slog.Warn("failed to log usage", "error", err)
	}
}

func getUsageStats(ctx context.Context, db *sql.DB, userID string, limit int) ([]UsageStats, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT provider, model,
		       COUNT(*) AS total_requests,
		       SUM(CASE WHEN success THEN 1 ELSE 0 END) AS successful_requests,
		       SUM(CASE WHEN NOT success THEN 1 ELSE 0 END) AS failed_requests,
		       COALESCE(SUM(total_tokens), 0) AS total_tokens,
		       COALESCE(AVG(latency_ms), 0) AS avg_latency_ms,
		       COALESCE(SUM(estimated_cost), 0) AS estimated_cost
		FROM ai_usage_logs
		WHERE user_id = $1
		GROUP BY provider, model
		ORDER BY total_requests DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []UsageStats
	for rows.Next() {
		var s UsageStats
		if err := rows.Scan(&s.Provider, &s.Model, &s.TotalRequests, &s.SuccessfulReq, &s.FailedReq, &s.TotalTokens, &s.AvgLatencyMs, &s.EstimatedCost); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}
