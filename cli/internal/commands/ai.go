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

func init() {
	aiCmd.AddCommand(aiCollectionsCmd)
	aiCmd.AddCommand(aiCreateCollectionCmd)
	aiCmd.AddCommand(aiDeleteCollectionCmd)
	aiCmd.AddCommand(aiSearchCmd)
	aiCmd.AddCommand(aiDocumentsCmd)
	aiCmd.AddCommand(aiEncryptionKeyCmd)

	aiDocumentsCmd.AddCommand(aiDocumentsListCmd)
	aiDocumentsCmd.AddCommand(aiDocumentsAddCmd)
	aiDocumentsCmd.AddCommand(aiDocumentsDeleteCmd)

	aiSearchCmd.Flags().IntP("top-k", "k", 5, "Number of results to return")
	aiDocumentsAddCmd.Flags().StringP("metadata", "m", "", "JSON metadata to attach")

	rootCmd.AddCommand(aiCmd)
}
