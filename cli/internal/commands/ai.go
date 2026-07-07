package commands

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/strata/strata/cli/internal/client"
	"github.com/strata/strata/cli/internal/output"
)

var aiCmd = &cobra.Command{
	Use:     "ai",
	Aliases: []string{"vector", "embed"},
	Short:   "Manage AI vector collections and semantic search",
	Long:    `Manage pgvector collections, documents, and perform semantic search.`,
}

var aiCollectionsCmd = &cobra.Command{
	Use:     "collections",
	Aliases: []string{"list", "ls"},
	Short:   "List all AI collections",
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.New()
		cols, err := cl.ListAICollections()
		if err != nil {
			output.Fatal(fmt.Errorf("failed to list collections: %w", err))
		}
		if len(cols) == 0 {
			output.Info("No collections found")
			output.Info("Create one with: strata ai collections create <name>")
			return
		}
		rows := [][]string{}
		for _, c := range cols {
			rows = append(rows, []string{c.Name, fmt.Sprintf("%d", c.DocCount), c.CreatedAt})
		}
		output.Table([]string{"Collection", "Documents", "Created"}, rows)
	},
}

var aiCreateCollectionCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new AI vector collection",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		cl := client.New()
		if err := cl.CreateAICollection(name); err != nil {
			output.Fatal(fmt.Errorf("failed to create collection: %w", err))
		}
		output.Success("Collection '%s' created", name)
	},
}

var aiDeleteCollectionCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an AI collection and all its documents",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		cl := client.New()
		if err := cl.DeleteAICollection(name); err != nil {
			output.Fatal(fmt.Errorf("failed to delete collection: %w", err))
		}
		output.Success("Collection '%s' deleted", name)
	},
}

var aiSearchCmd = &cobra.Command{
	Use:   "search <collection> <query>",
	Short: "Semantic search across a collection",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		collection := args[0]
		query := args[1]
		topK, _ := cmd.Flags().GetInt("top-k")
		if topK <= 0 {
			topK = 5
		}

		cl := client.New()
		results, err := cl.SearchAICollection(collection, query, topK)
		if err != nil {
			output.Fatal(fmt.Errorf("search failed: %w", err))
		}

		if len(results) == 0 {
			output.Info("No results found")
			return
		}

		rows := [][]string{}
		for i, r := range results {
			content := r.Content
			if len(content) > 60 {
				content = content[:60] + "..."
			}
			score := fmt.Sprintf("%.4f", r.Score)
			rows = append(rows, []string{fmt.Sprintf("%d", i+1), content, score})
		}
		output.Table([]string{"#", "Content", "Score"}, rows)
	},
}

var aiDocumentsCmd = &cobra.Command{
	Use:     "documents",
	Aliases: []string{"docs", "doc"},
	Short:   "Manage documents in a collection",
}

var aiDocumentsListCmd = &cobra.Command{
	Use:   "list <collection>",
	Short: "List documents in a collection",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		collection := args[0]
		cl := client.New()
		docs, err := cl.ListAIDocuments(collection)
		if err != nil {
			output.Fatal(fmt.Errorf("failed to list documents: %w", err))
		}
		if len(docs) == 0 {
			output.Info("No documents in collection '%s'", collection)
			return
		}
		rows := [][]string{}
		for _, d := range docs {
			content := d.Content
			if len(content) > 50 {
				content = content[:50] + "..."
			}
			rows = append(rows, []string{d.ID[:8]+"...", content, d.CreatedAt})
		}
		output.Table([]string{"ID", "Content", "Created"}, rows)
	},
}

var aiDocumentsAddCmd = &cobra.Command{
	Use:   "add <collection> <content>",
	Short: "Add a document to a collection",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		collection := args[0]
		content := args[1]
		metaJSON, _ := cmd.Flags().GetString("metadata")

		var metadata map[string]interface{}
		if metaJSON != "" {
			json.Unmarshal([]byte(metaJSON), &metadata)
		}

		cl := client.New()
		if err := cl.AddAIDocument(collection, content, metadata); err != nil {
			output.Fatal(fmt.Errorf("failed to add document: %w", err))
		}
		output.Success("Document added to '%s'", collection)
	},
}

var aiDocumentsDeleteCmd = &cobra.Command{
	Use:   "delete <collection> <doc-id>",
	Short: "Delete a document from a collection",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		collection := args[0]
		docID := args[1]
		cl := client.New()
		if err := cl.DeleteAIDocument(collection, docID); err != nil {
			output.Fatal(fmt.Errorf("failed to delete document: %w", err))
		}
		output.Success("Document deleted from '%s'", collection)
	},
}

var aiEncryptionKeyCmd = &cobra.Command{
	Use:   "encryption-key",
	Short: "Generate a new AES-256 encryption key",
	Run: func(cmd *cobra.Command, args []string) {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			output.Fatal(fmt.Errorf("failed to generate key: %w", err))
		}
		output.Success("AES-256-GCM Encryption Key (64 hex chars)")
		fmt.Println(hex.EncodeToString(key))
		fmt.Println()
		output.Info("Add this to your .env file as ENCRYPTION_KEY=<key>")
	},
}

var aiProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Manage AI providers",
}

var aiProvidersListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List configured AI providers",
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.New()
		providers, err := cl.ListAIProviders()
		if err != nil {
			output.Fatal(err)
		}
		if len(providers) == 0 {
			output.Info("No providers configured")
			output.Info("Add one with: strata ai providers add <provider> --key <api-key>")
			return
		}
		rows := [][]string{}
		for _, p := range providers {
			status := "disabled"
			if p.Enabled {
				status = "enabled"
				if p.IsPrimary {
					status = "primary"
				}
			}
			rows = append(rows, []string{p.Provider, p.DefaultModel, p.BaseURL, status})
		}
		output.Table([]string{"Provider", "Default Model", "Base URL", "Status"}, rows)
	},
}

var aiProvidersAddCmd = &cobra.Command{
	Use:   "add <provider>",
	Short: "Add an AI provider (openai/anthropic/google-gemini/cohere)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		provider := args[0]
		apiKey, _ := cmd.Flags().GetString("key")
		baseURL, _ := cmd.Flags().GetString("base-url")
		model, _ := cmd.Flags().GetString("model")
		cl := client.New()
		if err := cl.CreateAIProvider(provider, apiKey, baseURL, model); err != nil {
			output.Fatal(err)
		}
		output.Success("Provider '%s' added", provider)
	},
}

var aiProvidersDeleteCmd = &cobra.Command{
	Use:   "delete <provider>",
	Short: "Delete an AI provider",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		provider := args[0]
		cl := client.New()
		if err := cl.DeleteAIProvider(provider); err != nil {
			output.Fatal(err)
		}
		output.Success("Provider '%s' deleted", provider)
	},
}

var aiProvidersTestCmd = &cobra.Command{
	Use:   "test <provider>",
	Short: "Test an AI provider connection",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		provider := args[0]
		cl := client.New()
		result, err := cl.TestAIProvider(provider)
		if err != nil {
			output.Fatal(err)
		}
		output.Success("Provider '%s' is reachable", provider)
		output.JSON(result)
	},
}

var aiEmbedCmd = &cobra.Command{
	Use:   "embed <text...>",
	Short: "Generate embeddings for text",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			output.Fatal(fmt.Errorf("provide at least one text input"))
		}
		provider, _ := cmd.Flags().GetString("provider")
		model, _ := cmd.Flags().GetString("model")
		cl := client.New()
		result, err := cl.GenerateEmbeddings(args, provider, model)
		if err != nil {
			output.Fatal(err)
		}
		for i, emb := range result.Embeddings {
			output.Success("Input %d: %d dimensions", i+1, len(emb))
		}
		output.Table([]string{"Inputs", "Dimensions", "Model"}, [][]string{
			{fmt.Sprintf("%d", len(args)), fmt.Sprintf("%d", len(result.Embeddings[0])), result.Model},
		})
	},
}

var aiChatCmd = &cobra.Command{
	Use:   "chat <message>",
	Short: "Chat with an AI model",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		provider, _ := cmd.Flags().GetString("provider")
		model, _ := cmd.Flags().GetString("model")
		cl := client.New()
		body := map[string]interface{}{
			"message":  args[0],
			"provider": provider,
			"model":    model,
		}
		var result map[string]interface{}
		_, err := cl.DoRequest("POST", "/v1/ai/chat", body, &result)
		if err != nil {
			output.Fatal(err)
		}
		reply, _ := result["reply"].(string)
		fmt.Println(reply)
	},
}

var aiAgentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage AI Hub agents",
}

var aiAgentsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List AI Hub agents",
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.New()
		agents, err := cl.ListHubAgents()
		if err != nil {
			output.Fatal(err)
		}
		if len(agents) == 0 {
			output.Info("No agents configured")
			return
		}
		rows := [][]string{}
		for _, a := range agents {
			desc := a.Description
			if len(desc) > 50 {
				desc = desc[:50] + "..."
			}
			rows = append(rows, []string{a.ID[:8], a.Name, desc, a.Model})
		}
		output.Table([]string{"ID", "Name", "Description", "Model"}, rows)
	},
}

var aiAgentsChatCmd = &cobra.Command{
	Use:   "chat <agent-id> <message>",
	Short: "Chat with an AI agent",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		agentID := args[0]
		message := args[1]
		cl := client.New()
		result, err := cl.ChatWithAgent(agentID, message)
		if err != nil {
			output.Fatal(err)
		}
		var reply string
		if r, ok := result["reply"].(string); ok {
			reply = r
		} else if r, ok := result["response"].(string); ok {
			reply = r
		} else {
			reply = fmt.Sprintf("%v", result)
		}
		fmt.Println(reply)
	},
}

var aiWorkflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Manage AI Hub workflows",
}

var aiWorkflowsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List AI Hub workflows",
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.New()
		workflows, err := cl.ListHubWorkflows()
		if err != nil {
			output.Fatal(err)
		}
		if len(workflows) == 0 {
			output.Info("No workflows configured")
			return
		}
		rows := [][]string{}
		for _, w := range workflows {
			status := "disabled"
			if w.Enabled {
				status = "enabled"
			}
			rows = append(rows, []string{w.ID[:8], w.Name, fmt.Sprintf("%d", w.Nodes), status})
		}
		output.Table([]string{"ID", "Name", "Nodes", "Status"}, rows)
	},
}

var aiWorkflowsExecuteCmd = &cobra.Command{
	Use:   "execute <workflow-id>",
	Short: "Execute a workflow",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.New()
		result, err := cl.ExecuteWorkflow(args[0])
		if err != nil {
			output.Fatal(err)
		}
		if outputVal, ok := result["output"]; ok {
			outputJSON, _ := json.MarshalIndent(outputVal, "", "  ")
			fmt.Println(string(outputJSON))
		} else {
			output.JSON(result)
		}
		output.Success("Workflow '%s' executed", args[0])
	},
}

var aiPromptsCmd = &cobra.Command{
	Use:   "prompts",
	Short: "List AI Hub prompts",
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.New()
		prompts, err := cl.ListHubPrompts()
		if err != nil {
			output.Fatal(err)
		}
		if len(prompts) == 0 {
			output.Info("No prompts found")
			return
		}
		rows := [][]string{}
		for _, p := range prompts {
			id, _ := p["id"].(string)
			name, _ := p["name"].(string)
			content, _ := p["content"].(string)
			if len(content) > 50 {
				content = content[:50] + "..."
			}
			rows = append(rows, []string{id[:8], name, content})
		}
		output.Table([]string{"ID", "Name", "Content"}, rows)
	},
}

var aiUsageCmd = &cobra.Command{
	Use:     "usage",
	Aliases: []string{"stats"},
	Short:   "View AI usage statistics",
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.New()
		stats, err := cl.GetUsageStats()
		if err != nil {
			output.Fatal(err)
		}
		if len(stats) == 0 {
			output.Info("No usage data")
			return
		}
		rows := [][]string{}
		for _, s := range stats {
			provider, _ := s["provider"].(string)
			model, _ := s["model"].(string)
			tokens, _ := s["total_tokens"].(float64)
			cost, _ := s["cost"].(float64)
			rows = append(rows, []string{provider, model,
				fmt.Sprintf("%.0f", tokens),
				fmt.Sprintf("$%.4f", cost)})
		}
		output.Table([]string{"Provider", "Model", "Tokens", "Cost"}, rows)
	},
}

func init() {
	aiCmd.AddCommand(aiCollectionsCmd)
	aiCmd.AddCommand(aiCreateCollectionCmd)
	aiCmd.AddCommand(aiDeleteCollectionCmd)
	aiCmd.AddCommand(aiSearchCmd)
	aiCmd.AddCommand(aiDocumentsCmd)
	aiCmd.AddCommand(aiEncryptionKeyCmd)
	aiCmd.AddCommand(aiProvidersCmd)
	aiCmd.AddCommand(aiEmbedCmd)
	aiCmd.AddCommand(aiChatCmd)
	aiCmd.AddCommand(aiAgentsCmd)
	aiCmd.AddCommand(aiWorkflowsCmd)
	aiCmd.AddCommand(aiPromptsCmd)
	aiCmd.AddCommand(aiUsageCmd)

	aiDocumentsCmd.AddCommand(aiDocumentsListCmd)
	aiDocumentsCmd.AddCommand(aiDocumentsAddCmd)
	aiDocumentsCmd.AddCommand(aiDocumentsDeleteCmd)

	aiProvidersCmd.AddCommand(aiProvidersListCmd)
	aiProvidersCmd.AddCommand(aiProvidersAddCmd)
	aiProvidersCmd.AddCommand(aiProvidersDeleteCmd)
	aiProvidersCmd.AddCommand(aiProvidersTestCmd)

	aiAgentsCmd.AddCommand(aiAgentsListCmd)
	aiAgentsCmd.AddCommand(aiAgentsChatCmd)

	aiWorkflowsCmd.AddCommand(aiWorkflowsListCmd)
	aiWorkflowsCmd.AddCommand(aiWorkflowsExecuteCmd)

	aiSearchCmd.Flags().IntP("top-k", "k", 5, "Number of results to return")
	aiDocumentsAddCmd.Flags().StringP("metadata", "m", "", "JSON metadata to attach")
	aiProvidersAddCmd.Flags().StringP("key", "k", "", "API key")
	aiProvidersAddCmd.Flags().StringP("base-url", "b", "", "Base URL (defaults to provider default)")
	aiProvidersAddCmd.Flags().StringP("model", "m", "", "Default model")
	aiEmbedCmd.Flags().StringP("provider", "p", "", "Provider to use")
	aiEmbedCmd.Flags().StringP("model", "m", "", "Model to use")
	aiChatCmd.Flags().StringP("provider", "p", "", "Provider to use")
	aiChatCmd.Flags().StringP("model", "m", "", "Model to use")

	rootCmd.AddCommand(aiCmd)
}
