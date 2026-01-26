package odk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Project represents an ODK Central project
type Project struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Archived    bool       `json:"archived"`
	KeyId       *int       `json:"keyId"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// Form represents an ODK Central form
type Form struct {
	ProjectID       int        `json:"projectId"`
	XmlFormId       string     `json:"xmlFormId"`
	Name            string     `json:"name"`
	Version         string     `json:"version"`
	EnketoId        *string    `json:"enketoId"`
	Hash            string     `json:"hash"`
	State           string     `json:"state"` // open, closing, closed
	PublishedAt     *time.Time `json:"publishedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       *time.Time `json:"updatedAt"`
	KeyId           *int       `json:"keyId"`
	DraftToken      *string    `json:"draftToken"`
	EnketoOnceId    *string    `json:"enketoOnceId"`
	EntityRelated   bool       `json:"entityRelated"`
	ReviewStates    FormReviewStates `json:"reviewStates,omitempty"`
}

// FormReviewStates contains counts of submissions by review state
type FormReviewStates struct {
	Received   int `json:"received"`
	HasIssues  int `json:"hasIssues"`
	Edited     int `json:"edited"`
	Approved   int `json:"approved"`
	Rejected   int `json:"rejected"`
}

// ListProjects fetches all projects from ODK Central
func (c *Client) ListProjects() ([]Project, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	projectsURL := fmt.Sprintf("%s/v1/projects", c.config.BaseURL)

	req, err := http.NewRequest("GET", projectsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var projects []Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return projects, nil
}

// GetProject fetches a specific project by ID
func (c *Client) GetProject(projectID int) (*Project, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	projectURL := fmt.Sprintf("%s/v1/projects/%d", c.config.BaseURL, projectID)

	req, err := http.NewRequest("GET", projectURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("project not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &project, nil
}

// ListProjectForms fetches all forms for a specific project
func (c *Client) ListProjectForms(projectID int) ([]Form, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	formsURL := fmt.Sprintf("%s/v1/projects/%d/forms", c.config.BaseURL, projectID)

	req, err := http.NewRequest("GET", formsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch forms: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var forms []Form
	if err := json.NewDecoder(resp.Body).Decode(&forms); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return forms, nil
}

// GetForm fetches a specific form by project ID and form ID
func (c *Client) GetForm(projectID int, formID string) (*Form, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	formURL := fmt.Sprintf("%s/v1/projects/%d/forms/%s", c.config.BaseURL, projectID, formID)

	req, err := http.NewRequest("GET", formURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch form: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("form not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var form Form
	if err := json.NewDecoder(resp.Body).Decode(&form); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &form, nil
}
