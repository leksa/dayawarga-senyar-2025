package odk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is an HTTP client for ODK Central API
type Client struct {
	config     *ODKConfig
	httpClient *http.Client
	token      string
	tokenExp   time.Time
}

// NewClient creates a new ODK Central client
func NewClient(config *ODKConfig) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// authenticate gets a session token from ODK Central
func (c *Client) authenticate() error {
	// Check if token is still valid
	if c.token != "" && time.Now().Before(c.tokenExp) {
		return nil
	}

	authURL := fmt.Sprintf("%s/v1/sessions", c.config.BaseURL)

	payload := fmt.Sprintf(`{"email":"%s","password":"%s"}`, c.config.Email, c.config.Password)

	req, err := http.NewRequest("POST", authURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(strings.NewReader(payload))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expiresAt"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	c.token = authResp.Token
	c.tokenExp = authResp.ExpiresAt

	return nil
}

// GetSubmissions fetches submissions from ODK Central OData API
func (c *Client) GetSubmissions(filter string, skip, top int) (*ODataResponse, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	// Build OData URL
	odataURL := fmt.Sprintf("%s/v1/projects/%d/forms/%s.svc/Submissions",
		c.config.BaseURL, c.config.ProjectID, c.config.FormID)

	// Add query parameters
	params := url.Values{}
	if filter != "" {
		params.Set("$filter", filter)
	}
	if skip > 0 {
		params.Set("$skip", fmt.Sprintf("%d", skip))
	}
	if top > 0 {
		params.Set("$top", fmt.Sprintf("%d", top))
	}

	if len(params) > 0 {
		odataURL += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", odataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch submissions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var odataResp ODataResponse
	if err := json.NewDecoder(resp.Body).Decode(&odataResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &odataResp, nil
}

// GetSubmissionsRaw fetches raw submission data as map for flexible parsing
func (c *Client) GetSubmissionsRaw(filter string, skip, top int) ([]map[string]interface{}, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	odataURL := fmt.Sprintf("%s/v1/projects/%d/forms/%s.svc/Submissions",
		c.config.BaseURL, c.config.ProjectID, c.config.FormID)

	params := url.Values{}
	if filter != "" {
		params.Set("$filter", filter)
	}
	if skip > 0 {
		params.Set("$skip", fmt.Sprintf("%d", skip))
	}
	if top > 0 {
		params.Set("$top", fmt.Sprintf("%d", top))
	}

	if len(params) > 0 {
		odataURL += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", odataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch submissions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var rawResp struct {
		Value []map[string]interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return rawResp.Value, nil
}

// GetSubmissionsSince fetches submissions updated after a specific time
func (c *Client) GetSubmissionsSince(since time.Time) ([]map[string]interface{}, error) {
	filter := fmt.Sprintf("__system/updatedAt gt %s", since.UTC().Format(time.RFC3339))
	return c.GetSubmissionsRaw(filter, 0, 0)
}

// GetApprovedSubmissions fetches only approved submissions
func (c *Client) GetApprovedSubmissions() ([]map[string]interface{}, error) {
	filter := "__system/reviewState eq 'approved'"
	return c.GetSubmissionsRaw(filter, 0, 0)
}

// GetAllSubmissions fetches all submissions with pagination
func (c *Client) GetAllSubmissions() ([]map[string]interface{}, error) {
	var allSubmissions []map[string]interface{}
	skip := 0
	pageSize := 100

	for {
		submissions, err := c.GetSubmissionsRaw("", skip, pageSize)
		if err != nil {
			return nil, err
		}

		if len(submissions) == 0 {
			break
		}

		allSubmissions = append(allSubmissions, submissions...)

		if len(submissions) < pageSize {
			break
		}

		skip += pageSize
	}

	return allSubmissions, nil
}

// GetAttachment downloads an attachment from a submission
func (c *Client) GetAttachment(submissionID, filename string) ([]byte, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	attachmentURL := fmt.Sprintf("%s/v1/projects/%d/forms/%s/submissions/%s/attachments/%s",
		c.config.BaseURL, c.config.ProjectID, c.config.FormID, submissionID, filename)

	req, err := http.NewRequest("GET", attachmentURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attachment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("attachment request failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
