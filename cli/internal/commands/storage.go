package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/strata/strata/cli/internal/client"
	"github.com/strata/strata/cli/internal/config"
	"github.com/strata/strata/cli/internal/output"
)

var storageCmd = &cobra.Command{
	Use:     "storage",
	Aliases: []string{"st", "store"},
	Short:   "Manage S3-compatible storage",
	Long:    `Manage buckets and files in Strata's S3-compatible storage.`,
}

var storageBucketsCmd = &cobra.Command{
	Use:     "buckets list",
	Aliases: []string{"list", "ls"},
	Short:   "List all storage buckets",
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.New()
		buckets, err := cl.ListBuckets()
		if err != nil {
			output.Fatal(fmt.Errorf("failed to list buckets: %w", err))
		}
		if len(buckets) == 0 {
			output.Info("No buckets found")
			output.Info("Create one with: strata storage buckets create <name>")
			return
		}
		rows := [][]string{}
		for _, b := range buckets {
			rows = append(rows, []string{b.Name, b.CreatedAt})
		}
		output.Table([]string{"Bucket", "Created"}, rows)
	},
}

var storageCreateBucketCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new storage bucket",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		cl := client.New()
		if err := cl.CreateBucket(name); err != nil {
			output.Fatal(fmt.Errorf("failed to create bucket: %w", err))
		}
		output.Success("Bucket '%s' created", name)
	},
}

var storageDeleteBucketCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a storage bucket",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		cl := client.New()
		if err := cl.DeleteBucket(name); err != nil {
			output.Fatal(fmt.Errorf("failed to delete bucket: %w", err))
		}
		output.Success("Bucket '%s' deleted", name)
	},
}

var storageUploadCmd = &cobra.Command{
	Use:   "upload <bucket> <file>",
	Short: "Upload a file to a bucket",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := args[0]
		filePath := args[1]

		file, err := os.Open(filePath)
		if err != nil {
			output.Fatal(fmt.Errorf("failed to open file: %w", err))
		}
		defer file.Close()

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		fw, err := w.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			output.Fatal(fmt.Errorf("failed to create form: %w", err))
		}
		if _, err := io.Copy(fw, file); err != nil {
			output.Fatal(fmt.Errorf("failed to copy file data: %w", err))
		}
		w.Close()

		baseURL := config.GetGatewayURL()
		url := fmt.Sprintf("%s/v1/storage/buckets/%s/upload", baseURL, bucket)
		req, err := http.NewRequest("POST", url, &buf)
		if err != nil {
			output.Fatal(fmt.Errorf("failed to create request: %w", err))
		}
		req.Header.Set("Content-Type", w.FormDataContentType())
		if token := config.GetAccessToken(); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			output.Fatal(fmt.Errorf("upload failed: %w", err))
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data, _ := json.MarshalIndent(result, "", "  ")
		output.Success("File uploaded to bucket '%s'", bucket)
		fmt.Println(string(data))
	},
}

var storageDownloadCmd = &cobra.Command{
	Use:   "download <bucket> <path> [output]",
	Short: "Download a file from a bucket",
	Args:  cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := args[0]
		filePath := args[1]

		outPath := filepath.Base(filePath)
		if len(args) > 2 {
			outPath = args[2]
		}

		f, err := os.Create(outPath)
		if err != nil {
			output.Fatal(fmt.Errorf("failed to create output file: %w", err))
		}
		defer f.Close()

		cl := client.New()
		if err := cl.DownloadFile(bucket, filePath, f); err != nil {
			output.Fatal(fmt.Errorf("download failed: %w", err))
		}

		output.Success("Downloaded to %s", outPath)
	},
}

func init() {
	storageCmd.AddCommand(storageBucketsCmd)
	storageCmd.AddCommand(storageCreateBucketCmd)
	storageCmd.AddCommand(storageDeleteBucketCmd)
	storageCmd.AddCommand(storageUploadCmd)
	storageCmd.AddCommand(storageDownloadCmd)

	rootCmd.AddCommand(storageCmd)
}
