package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client defines the interface for Jira API operations
type Client interface {
	GetIssue(ctx context.Context, issueID string) (*Issue, error)
}

// HTTPClient implements the Jira Client interface
type HTTPClient struct {
	baseURL  string
	username string
	token    string
	client   *http.Client
}

// NewClient creates a new Jira API client
func NewClient(baseURL, username, token string) Client {
	return &HTTPClient{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		username: username,
		token:    token,
		client:   &http.Client{},
	}
}

// jiraAPIResponse represents the JSON response from Jira API
type jiraAPIResponse struct {
	Key    string `json:"key"`
	Fields struct {
		Summary     string        `json:"summary"`
		Description interface{}   `json:"description"`
		Environment interface{}   `json:"environment"`
		Labels      []string      `json:"labels"`
		Components  []jiraComponent `json:"components"`
	} `json:"fields"`
}

type jiraComponent struct {
	Name string `json:"name"`
}

// GetIssue fetches a Jira issue by ID
func (c *HTTPClient) GetIssue(ctx context.Context, issueID string) (*Issue, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", c.baseURL, issueID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Basic Auth
	req.SetBasicAuth(c.username, c.token)
	req.Header.Set("Accept", "application/json")

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
		return nil, fmt.Errorf("authentication failed: check your Jira credentials")
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

	// Extract description text from Atlassian Document Format (ADF)
	description := extractDescription(apiResp.Fields.Description)

	// Extract environment field (can be string or object depending on field type)
	environment := extractEnvironment(apiResp.Fields.Environment)

	// Extract component names
	components := make([]string, len(apiResp.Fields.Components))
	for i, comp := range apiResp.Fields.Components {
		components[i] = comp.Name
	}

	issue := &Issue{
		ID:          apiResp.Key,
		Summary:     apiResp.Fields.Summary,
		Description: description,
		Environment: environment,
		Labels:      apiResp.Fields.Labels,
		Components:  components,
	}

	return issue, nil
}

// extractDescription extracts plain text from Atlassian Document Format
func extractDescription(adf interface{}) string {
	// Type assert to map to access content
	adfMap, ok := adf.(map[string]interface{})
	if !ok {
		return ""
	}

	content, ok := adfMap["content"].([]interface{})
	if !ok {
		return ""
	}

	var texts []string
	for _, block := range content {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}

		innerContent, ok := blockMap["content"].([]interface{})
		if !ok {
			continue
		}

		for _, item := range innerContent {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			if text, ok := itemMap["text"].(string); ok {
				texts = append(texts, text)
			}
		}
	}

	return strings.Join(texts, " ")
}

// extractEnvironment extracts environment field value
func extractEnvironment(env interface{}) string {
	if env == nil {
		return ""
	}

	// If it's a string, return directly
	if str, ok := env.(string); ok {
		return str
	}

	// If it's an object (like ADF), try to extract text
	if envMap, ok := env.(map[string]interface{}); ok {
		if content, ok := envMap["content"].([]interface{}); ok {
			var texts []string
			for _, block := range content {
				if blockMap, ok := block.(map[string]interface{}); ok {
					if innerContent, ok := blockMap["content"].([]interface{}); ok {
						for _, item := range innerContent {
							if itemMap, ok := item.(map[string]interface{}); ok {
								if text, ok := itemMap["text"].(string); ok {
									texts = append(texts, text)
								}
							}
						}
					}
				}
			}
			return strings.Join(texts, " ")
		}
	}

	return ""
}
