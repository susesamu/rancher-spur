package jira

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetIssue_Success(t *testing.T) {
	// Mock Jira API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		username, password, ok := r.BasicAuth()
		if !ok || username != "test@example.com" || password != "test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Return mock issue data
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"key": "TEST-123",
			"fields": {
				"summary": "Test issue summary",
				"description": {
					"content": [
						{
							"content": [
								{"text": "This is the issue description"}
							]
						}
					]
				},
				"environment": "Production environment",
				"labels": ["bug", "critical"],
				"components": [
					{"name": "backend"},
					{"name": "api"}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test@example.com", "test-token")
	issue, err := client.GetIssue(context.Background(), "TEST-123")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if issue.ID != "TEST-123" {
		t.Errorf("expected ID 'TEST-123', got '%s'", issue.ID)
	}
	if issue.Summary != "Test issue summary" {
		t.Errorf("expected summary 'Test issue summary', got '%s'", issue.Summary)
	}
	if issue.Description != "This is the issue description" {
		t.Errorf("expected description, got '%s'", issue.Description)
	}
	if issue.Environment != "Production environment" {
		t.Errorf("expected environment 'Production environment', got '%s'", issue.Environment)
	}
	if len(issue.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(issue.Labels))
	}
	if len(issue.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(issue.Components))
	}
}

func TestGetIssue_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errorMessages":["Issue does not exist"]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test@example.com", "test-token")
	_, err := client.GetIssue(context.Background(), "NOTFOUND-123")

	if err == nil {
		t.Fatal("expected error for not found issue, got none")
	}
}

func TestGetIssue_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errorMessages":["Authentication failed"]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "wrong@example.com", "wrong-token")
	_, err := client.GetIssue(context.Background(), "TEST-123")

	if err == nil {
		t.Fatal("expected error for unauthorized, got none")
	}
}

func TestGetIssue_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errorMessages":["Internal server error"]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test@example.com", "test-token")
	_, err := client.GetIssue(context.Background(), "TEST-123")

	if err == nil {
		t.Fatal("expected error for server error, got none")
	}
}

func TestGetIssue_EmptyFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"key": "TEST-456",
			"fields": {
				"summary": "Minimal issue",
				"description": null,
				"environment": null,
				"labels": [],
				"components": []
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test@example.com", "test-token")
	issue, err := client.GetIssue(context.Background(), "TEST-456")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if issue.Description != "" {
		t.Errorf("expected empty description, got '%s'", issue.Description)
	}
	if issue.Environment != "" {
		t.Errorf("expected empty environment, got '%s'", issue.Environment)
	}
	if len(issue.Labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(issue.Labels))
	}
	if len(issue.Components) != 0 {
		t.Errorf("expected 0 components, got %d", len(issue.Components))
	}
}
