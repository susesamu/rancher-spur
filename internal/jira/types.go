package jira

// Issue represents a Jira issue with relevant fields for environment reproduction
type Issue struct {
	ID          string
	Summary     string
	Description string
	Environment string
	Labels      []string
	Components  []string
}

// Attachment represents a Jira attachment
type Attachment struct {
	ID       string
	Filename string
	URL      string
	MimeType string
	Size     int64
}
