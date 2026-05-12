package jira

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetIssue_Success(t *testing.T) {
	// Mock Jira API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-bearer-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Verify it's using API v2 with correct fields
		if r.URL.Path != "/rest/api/2/issue/TEST-123" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Return mock issue data (API v2 format with plain text)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "679134",
			"key": "TEST-123",
			"fields": {
				"summary": "Test issue summary",
				"description": "This is the plain text issue description",
				"environment": "Rancher version: 2.13.5/2.14.1"
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-bearer-token")
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
	if issue.Description != "This is the plain text issue description" {
		t.Errorf("expected plain text description, got '%s'", issue.Description)
	}
	if issue.Environment != "Rancher version: 2.13.5/2.14.1" {
		t.Errorf("expected environment 'Rancher version: 2.13.5/2.14.1', got '%s'", issue.Environment)
	}
}

func TestGetIssue_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errorMessages":["Issue does not exist"]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-bearer-token")
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

	client := NewClient(server.URL, "wrong-bearer-token")
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

	client := NewClient(server.URL, "test-bearer-token")
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
			"id": "456",
			"key": "TEST-456",
			"fields": {
				"summary": "Minimal issue",
				"description": "",
				"environment": ""
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-bearer-token")
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
}

func TestGetIssue_RealWorldExample(t *testing.T) {
	// Test with the actual SUSE Jira response format
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-bearer-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "679134",
			"key": "SURE-11610",
			"fields": {
				"summary": "Rancher removes chartValues in YAML editor",
				"description": "*Issue description:*\r\nCustomer found a bug in Rancher UI where chartValues got removed during the editing of cluster YAML.",
				"environment": "*Rancher Cluster:*\r\nRancher version: 2.13.5/2.14.1"
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-bearer-token")
	issue, err := client.GetIssue(context.Background(), "SURE-11610")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if issue.ID != "SURE-11610" {
		t.Errorf("expected ID 'SURE-11610', got '%s'", issue.ID)
	}
	if issue.Summary != "Rancher removes chartValues in YAML editor" {
		t.Errorf("unexpected summary: '%s'", issue.Summary)
	}
	if !contains(issue.Environment, "Rancher version: 2.13.5/2.14.1") {
		t.Errorf("expected environment to contain Rancher version, got '%s'", issue.Environment)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestListAttachments_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-bearer-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"fields": {
				"attachment": [
					{
						"id": "316999",
						"filename": "2026-03-24_15-52-37.yaml",
						"content": "https://jira.suse.com/secure/attachment/316999/2026-03-24_15-52-37.yaml",
						"mimeType": "text/yaml",
						"size": 1024
					},
					{
						"id": "316998",
						"filename": "logs.tar.gz",
						"content": "https://jira.suse.com/secure/attachment/316998/logs.tar.gz",
						"mimeType": "application/gzip",
						"size": 2048
					}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-bearer-token")
	attachments, err := client.ListAttachments(context.Background(), "SURE-11483")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(attachments) != 2 {
		t.Fatalf("expected 2 attachments, got %d", len(attachments))
	}

	if attachments[0].Filename != "2026-03-24_15-52-37.yaml" {
		t.Errorf("expected filename '2026-03-24_15-52-37.yaml', got '%s'", attachments[0].Filename)
	}

	if attachments[0].Size != 1024 {
		t.Errorf("expected size 1024, got %d", attachments[0].Size)
	}

	if attachments[1].MimeType != "application/gzip" {
		t.Errorf("expected mime type 'application/gzip', got '%s'", attachments[1].MimeType)
	}
}

func TestListAttachments_NoAttachments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"fields": {
				"attachment": []
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-bearer-token")
	attachments, err := client.ListAttachments(context.Background(), "SURE-11483")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(attachments) != 0 {
		t.Errorf("expected 0 attachments, got %d", len(attachments))
	}
}

func TestListAttachments_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "wrong-token")
	_, err := client.ListAttachments(context.Background(), "SURE-11483")

	if err == nil {
		t.Fatal("expected error for unauthorized, got none")
	}
}

func TestDownloadAttachment_Success(t *testing.T) {
	testContent := "test file content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-bearer-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-bearer-token")

	tmpDir := t.TempDir()
	destPath := tmpDir + "/test.txt"

	err := client.DownloadAttachment(context.Background(), server.URL+"/test", destPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify file was created and has correct content
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("expected content '%s', got '%s'", testContent, string(content))
	}
}

func TestDownloadAttachment_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "wrong-token")

	tmpDir := t.TempDir()
	destPath := tmpDir + "/test.txt"

	err := client.DownloadAttachment(context.Background(), server.URL+"/test", destPath)
	if err == nil {
		t.Fatal("expected error for unauthorized, got none")
	}
}
