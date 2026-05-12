package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Client defines the interface for Jira API operations
type Client interface {
	GetIssue(ctx context.Context, issueID string) (*Issue, error)
	ListAttachments(ctx context.Context, issueID string) ([]Attachment, error)
	DownloadAttachment(ctx context.Context, url, destPath string) error
}

// HTTPClient implements the Jira Client interface
type HTTPClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewClient creates a new Jira API client with Bearer token authentication
func NewClient(baseURL, token string) Client {
	return &HTTPClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		client:  &http.Client{},
	}
}

// jiraAPIResponse represents the JSON response from Jira API v2
type jiraAPIResponse struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		Environment string `json:"environment"`
	} `json:"fields"`
}

// GetIssue fetches a Jira issue by ID using API v2
func (c *HTTPClient) GetIssue(ctx context.Context, issueID string) (*Issue, error) {
	url := fmt.Sprintf("%s/rest/api/2/issue/%s?fields=summary,description,environment", c.baseURL, issueID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Bearer token authentication
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("issue %s not found", issueID)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("authentication failed: check your Jira Bearer token")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp jiraAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	issue := &Issue{
		ID:          apiResp.Key,
		Summary:     apiResp.Fields.Summary,
		Description: apiResp.Fields.Description,
		Environment: apiResp.Fields.Environment,
		Labels:      []string{},
		Components:  []string{},
	}

	return issue, nil
}

// ListAttachments fetches all attachments for a Jira issue
func (c *HTTPClient) ListAttachments(ctx context.Context, issueID string) ([]Attachment, error) {
	url := fmt.Sprintf("%s/rest/api/2/issue/%s?fields=attachment", c.baseURL, issueID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attachments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("issue %s not found", issueID)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("authentication failed: check your Jira Bearer token")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Fields struct {
			Attachment []struct {
				ID       string `json:"id"`
				Filename string `json:"filename"`
				Content  string `json:"content"`
				MimeType string `json:"mimeType"`
				Size     int64  `json:"size"`
			} `json:"attachment"`
		} `json:"fields"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	attachments := make([]Attachment, 0, len(response.Fields.Attachment))
	for _, att := range response.Fields.Attachment {
		attachments = append(attachments, Attachment{
			ID:       att.ID,
			Filename: att.Filename,
			URL:      att.Content,
			MimeType: att.MimeType,
			Size:     att.Size,
		})
	}

	return attachments, nil
}

// DownloadAttachment downloads a Jira attachment to the specified path
func (c *HTTPClient) DownloadAttachment(ctx context.Context, url, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download attachment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
