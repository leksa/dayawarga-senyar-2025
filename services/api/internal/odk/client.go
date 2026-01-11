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

// GetAttachmentForForm downloads an attachment from a submission for a specific form
func (c *Client) GetAttachmentForForm(formID, submissionID, filename string) ([]byte, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	attachmentURL := fmt.Sprintf("%s/v1/projects/%d/forms/%s/submissions/%s/attachments/%s",
		c.config.BaseURL, c.config.ProjectID, formID, submissionID, filename)

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

// GetDatasets lists all datasets (entity lists) in the project
func (c *Client) GetDatasets() ([]map[string]interface{}, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	datasetsURL := fmt.Sprintf("%s/v1/projects/%d/datasets",
		c.config.BaseURL, c.config.ProjectID)

	req, err := http.NewRequest("GET", datasetsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch datasets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var datasets []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&datasets); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return datasets, nil
}

// GetEntities lists all entities in a dataset
func (c *Client) GetEntities(datasetName string) ([]map[string]interface{}, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	entitiesURL := fmt.Sprintf("%s/v1/projects/%d/datasets/%s/entities",
		c.config.BaseURL, c.config.ProjectID, datasetName)

	req, err := http.NewRequest("GET", entitiesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch entities: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var entities []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&entities); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return entities, nil
}

// CreateEntity creates a single entity in a dataset
func (c *Client) CreateEntity(datasetName string, entity EntityCreateRequest) (*map[string]interface{}, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	entitiesURL := fmt.Sprintf("%s/v1/projects/%d/datasets/%s/entities",
		c.config.BaseURL, c.config.ProjectID, datasetName)

	payload, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}

	req, err := http.NewRequest("POST", entitiesURL, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// CreateEntitiesBulk creates multiple entities in a dataset
func (c *Client) CreateEntitiesBulk(datasetName string, entities []EntityCreateRequest, sourceName string) ([]map[string]interface{}, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	entitiesURL := fmt.Sprintf("%s/v1/projects/%d/datasets/%s/entities",
		c.config.BaseURL, c.config.ProjectID, datasetName)

	bulkRequest := BulkEntityCreateRequest{
		Entities: entities,
		Source: EntitySource{
			Name: sourceName,
			Size: len(entities),
		},
	}

	payload, err := json.Marshal(bulkRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entities: %w", err)
	}

	req, err := http.NewRequest("POST", entitiesURL, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create entities: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var results []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetEntitySubmissionMapping builds a mapping from entity UUID to submission instance ID
// by fetching entity versions which contain the source submission info
func (c *Client) GetEntitySubmissionMapping(datasetName string) (map[string]string, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	// First, get all entities
	entities, err := c.GetEntities(datasetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	mapping := make(map[string]string)

	// For each entity, get its first version to find the source submission
	for _, entity := range entities {
		entityUUID, ok := entity["uuid"].(string)
		if !ok || entityUUID == "" {
			continue
		}

		// Get entity versions
		versionsURL := fmt.Sprintf("%s/v1/projects/%d/datasets/%s/entities/%s/versions",
			c.config.BaseURL, c.config.ProjectID, datasetName, entityUUID)

		req, err := http.NewRequest("GET", versionsURL, nil)
		if err != nil {
			continue
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		var versions []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Get submission ID from first version's source
		if len(versions) > 0 {
			if source, ok := versions[0]["source"].(map[string]interface{}); ok {
				if submission, ok := source["submission"].(map[string]interface{}); ok {
					if instanceID, ok := submission["instanceId"].(string); ok {
						mapping[entityUUID] = instanceID
					}
				}
			}
		}
	}

	return mapping, nil
}

// EntityCreateRequest represents request to create an entity
type EntityCreateRequest struct {
	UUID  string            `json:"uuid,omitempty"`
	Label string            `json:"label"`
	Data  map[string]string `json:"data"`
}

// BulkEntityCreateRequest represents request to create multiple entities
type BulkEntityCreateRequest struct {
	Entities []EntityCreateRequest `json:"entities"`
	Source   EntitySource          `json:"source"`
}

// EntitySource represents the source of bulk entity creation
type EntitySource struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}
