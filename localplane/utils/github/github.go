package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type contentItem struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
	URL         string `json:"url"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
}

// Client represents a GitHub repository access object. Optional fields are
// pointer types and can be nil when not set.
type Client struct {
	Owner      string
	Repo       string
	Ref        *string
	Token      *string
	HTTPClient *http.Client
}

// NewClient creates a minimal Client with required owner and repo.
func NewClient(owner, repo string) *Client {
	return &Client{Owner: owner, Repo: repo}
}

// DownloadPath downloads the file or directory at `srcPath` from the GitHub
// repository into `destDir`, preserving relative paths. Optional fields on the
// Client (Ref, Token) may be nil.
func (c *Client) DownloadPath(ctx context.Context, srcPath, destDir string) error {
	if c == nil || c.Owner == "" || c.Repo == "" {
		return errors.New("owner and repo are required on client")
	}
	srcPath = strings.TrimPrefix(srcPath, "/")

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{}
	}

	var fetch func(apiPath, basePath string) error
	fetch = func(apiPath, basePath string) error {
		apiPathEscaped := url.PathEscape(apiPath)
		apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", c.Owner, c.Repo, apiPathEscaped)
		if c.Ref != nil && *c.Ref != "" {
			apiURL = apiURL + "?ref=" + url.QueryEscape(*c.Ref)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return err
		}
		if c.Token != nil && *c.Token != "" {
			req.Header.Set("Authorization", "token "+*c.Token)
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("path not found: %s", apiPath)
		}
		if resp.StatusCode >= 400 {
			b, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("github api error: %s: %s", resp.Status, string(b))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var arr []contentItem
		if err := json.Unmarshal(body, &arr); err == nil {
			for _, it := range arr {
				if err := fetch(it.Path, basePath); err != nil {
					return err
				}
			}
			return nil
		}

		var it contentItem
		if err := json.Unmarshal(body, &it); err != nil {
			return fmt.Errorf("unable to parse github response: %w", err)
		}

		if it.Type == "dir" {
			req2, _ := http.NewRequestWithContext(ctx, "GET", it.URL, nil)
			if c.Token != nil && *c.Token != "" {
				req2.Header.Set("Authorization", "token "+*c.Token)
			}
			req2.Header.Set("Accept", "application/vnd.github.v3+json")
			resp2, err := client.Do(req2)
			if err != nil {
				return err
			}
			defer resp2.Body.Close()
			if resp2.StatusCode >= 400 {
				b, _ := io.ReadAll(resp2.Body)
				return fmt.Errorf("github api error: %s: %s", resp2.Status, string(b))
			}
			body2, _ := io.ReadAll(resp2.Body)
			var arr2 []contentItem
			if err := json.Unmarshal(body2, &arr2); err != nil {
				return err
			}
			for _, child := range arr2 {
				if err := fetch(child.Path, basePath); err != nil {
					return err
				}
			}
			return nil
		}

		if it.Type == "file" || it.Type == "symlink" {
			rel := strings.TrimPrefix(it.Path, strings.TrimPrefix(srcPath, "/"))
			rel = strings.TrimPrefix(rel, "/")
			destPath := filepath.Join(destDir, rel)
			if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
				return err
			}

			if it.DownloadURL != "" {
				reqf, _ := http.NewRequestWithContext(ctx, "GET", it.DownloadURL, nil)
				if c.Token != nil && *c.Token != "" {
					reqf.Header.Set("Authorization", "token "+*c.Token)
				}
				respf, err := client.Do(reqf)
				if err != nil {
					return err
				}
				defer respf.Body.Close()
				if respf.StatusCode >= 400 {
					b, _ := io.ReadAll(respf.Body)
					return fmt.Errorf("download error: %s: %s", respf.Status, string(b))
				}
				out, err := os.Create(destPath)
				if err != nil {
					return err
				}
				defer out.Close()
				if _, err := io.Copy(out, respf.Body); err != nil {
					return err
				}
				return nil
			}

			if it.Content != "" && it.Encoding == "base64" {
				decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(it.Content, "\n", ""))
				if err != nil {
					return err
				}
				if err := os.WriteFile(destPath, decoded, 0o644); err != nil {
					return err
				}
				return nil
			}

			return fmt.Errorf("file has no download data: %s", it.Path)
		}

		return fmt.Errorf("unsupported content type: %s", it.Type)
	}

	return fetch(srcPath, srcPath)
}

// DownloadRepoPath is a convenience wrapper that preserves the previous
// functional API but constructs a Client and calls its method. Optional fields
// may be passed as empty strings to remain nil.
func DownloadRepoPath(ctx context.Context, owner, repo, ref, srcPath, destDir, token string) error {
	c := &Client{Owner: owner, Repo: repo}
	if ref != "" {
		c.Ref = &ref
	}
	if token != "" {
		c.Token = &token
	}
	return c.DownloadPath(ctx, srcPath, destDir)
}
