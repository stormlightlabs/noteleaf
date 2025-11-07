//go:build !prod

package tools

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// GitHubCommit represents a GitHub API commit response
type GitHubCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
	} `json:"commit"`
}

// ArchiveConfig contains configuration for fetching and extracting archives
type ArchiveConfig struct {
	Repo       string
	Path       string
	Output     string
	SHA        string
	FormatJSON bool
}

// NewGHRepoCommand creates a command for fetching GitHub repository archives
func NewGHRepoCommand() *cobra.Command {
	var config ArchiveConfig

	cmd := &cobra.Command{
		Use:   "gh-repo",
		Short: "Fetch and extract files from a GitHub repository archive",
		Long: `Fetches a GitHub repository archive (tarball), extracts specific paths,
and optionally formats JSON files using Go's standard library.

This is useful for syncing lexicons, schemas, or other data files from GitHub repositories.`,
		Example: `  # Fetch lexicons from a specific path
  noteleaf tools fetch gh-repo \
    --repo hyperlink-academy/leaflet \
    --path lexicons/pub/leaflet/ \
    --output lexdocs/leaflet/

  # Fetch from a specific commit
  noteleaf tools fetch gh-repo \
    --repo owner/repo \
    --path schemas/ \
    --output local/schemas/ \
    --sha abc123def`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Repo == "" {
				return fmt.Errorf("--repo is required")
			}
			if config.Path == "" {
				return fmt.Errorf("--path is required")
			}
			if config.Output == "" {
				return fmt.Errorf("--output is required")
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			return fetchAndExtractArchive(ctx, config, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&config.Repo, "repo", "", "GitHub repository (owner/name)")
	cmd.Flags().StringVar(&config.Path, "path", "", "Path within repository to extract")
	cmd.Flags().StringVar(&config.Output, "output", "", "Output directory for extracted files")
	cmd.Flags().StringVar(&config.SHA, "sha", "", "Specific commit SHA (default: latest)")
	cmd.Flags().BoolVar(&config.FormatJSON, "format-json", true, "Format JSON files with indentation")
	return cmd
}

// fetchAndExtractArchive fetches a GitHub archive and extracts specific paths
func fetchAndExtractArchive(ctx context.Context, config ArchiveConfig, out io.Writer) error {
	sha := config.SHA
	if sha == "" {
		var err error
		sha, err = getLatestCommit(ctx, config.Repo, config.Path)
		if err != nil {
			return fmt.Errorf("failed to get latest commit: %w", err)
		}
		fmt.Fprintf(out, "Latest commit: %s\n", sha)
	}

	tmpDir, err := os.MkdirTemp("", "repo-archive-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Fprintf(out, "Fetching archive for %s@%s\n", config.Repo, sha[:7])
	if err := downloadAndExtract(ctx, config.Repo, sha, config.Path, tmpDir, config.FormatJSON, out); err != nil {
		return fmt.Errorf("failed to download and extract: %w", err)
	}

	fmt.Fprintf(out, "Writing README with source information\n")
	readme := fmt.Sprintf("Source: https://github.com/%s/tree/%s/%s\n", config.Repo, sha, config.Path)
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(readme), 0o644); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}

	fmt.Fprintf(out, "Moving extracted files to %s\n", config.Output)
	if err := os.RemoveAll(config.Output); err != nil {
		return fmt.Errorf("failed to remove existing output directory: %w", err)
	}
	if err := os.Rename(tmpDir, config.Output); err != nil {
		return fmt.Errorf("failed to move files to output directory: %w", err)
	}

	fmt.Fprintf(out, "Successfully extracted archive to %s\n", config.Output)
	return nil
}

// getLatestCommit fetches the latest commit SHA for a given repository and path
func getLatestCommit(ctx context.Context, repo, path string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits?path=%s&per_page=1", repo, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var commits []GitHubCommit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found for path %s", path)
	}

	return commits[0].SHA, nil
}

// downloadAndExtract downloads a GitHub archive and extracts files from a specific path
func downloadAndExtract(ctx context.Context, repo, sha, extractPath, outputDir string, formatJSON bool, out io.Writer) error {
	url := fmt.Sprintf("https://github.com/%s/archive/%s.tar.gz", repo, sha)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download archive: status %d", resp.StatusCode)
	}

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	repoName := strings.Split(repo, "/")[1]
	prefix := fmt.Sprintf("%s-%s/%s", repoName, sha, extractPath)

	fmt.Fprintf(out, "Extracting files from %s\n", prefix)

	fileCount := 0
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		if !strings.HasPrefix(header.Name, prefix) {
			continue
		}

		if !strings.HasSuffix(header.Name, ".json") {
			continue
		}

		relativePath := strings.TrimPrefix(header.Name, prefix)
		outputPath := filepath.Join(outputDir, relativePath)

		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", outputPath, err)
		}

		data, err := io.ReadAll(tr)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", header.Name, err)
		}

		if formatJSON {
			var jsonData any
			if err := json.Unmarshal(data, &jsonData); err != nil {
				return fmt.Errorf("failed to parse JSON in %s: %w", header.Name, err)
			}

			formattedData, err := json.MarshalIndent(jsonData, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON in %s: %w", header.Name, err)
			}
			data = append(formattedData, '\n')
		}

		if err := os.WriteFile(outputPath, data, 0o644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", outputPath, err)
		}

		fileCount++
	}

	fmt.Fprintf(out, "Extracted %d files\n", fileCount)
	return nil
}
