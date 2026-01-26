package odk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AppUser represents an ODK Central App User (field key)
type AppUser struct {
	ID          int        `json:"id"`
	Type        string     `json:"type"` // "field_key"
	ProjectID   int        `json:"projectId"`
	DisplayName string     `json:"displayName"`
	Token       string     `json:"token,omitempty"` // Only returned on creation
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
}

// AppUserCreateRequest represents request to create an App User
type AppUserCreateRequest struct {
	DisplayName string `json:"displayName"`
}

// ListAppUsers fetches all app users for a project
func (c *Client) ListAppUsers(projectID int) ([]AppUser, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	appUsersURL := fmt.Sprintf("%s/v1/projects/%d/app-users", c.config.BaseURL, projectID)

	req, err := http.NewRequest("GET", appUsersURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app users: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var appUsers []AppUser
	if err := json.NewDecoder(resp.Body).Decode(&appUsers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return appUsers, nil
}

// CreateAppUser creates a new app user for a project
// Returns the created app user with token (token is only returned once!)
func (c *Client) CreateAppUser(projectID int, displayName string) (*AppUser, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	appUsersURL := fmt.Sprintf("%s/v1/projects/%d/app-users", c.config.BaseURL, projectID)

	payload, err := json.Marshal(AppUserCreateRequest{DisplayName: displayName})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", appUsersURL, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create app user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var appUser AppUser
	if err := json.NewDecoder(resp.Body).Decode(&appUser); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &appUser, nil
}

// DeleteAppUser removes an app user from a project
func (c *Client) DeleteAppUser(projectID, appUserID int) error {
	if err := c.authenticate(); err != nil {
		return err
	}

	appUserURL := fmt.Sprintf("%s/v1/projects/%d/app-users/%d", c.config.BaseURL, projectID, appUserID)

	req, err := http.NewRequest("DELETE", appUserURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete app user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// AssignFormToAppUser assigns a form to an app user
// roleId is typically 2 for "App User" role in ODK Central
func (c *Client) AssignFormToAppUser(projectID int, formID string, appUserID int) error {
	if err := c.authenticate(); err != nil {
		return err
	}

	// Role ID 2 is the "App User" role in ODK Central
	const appUserRoleID = 2

	assignURL := fmt.Sprintf("%s/v1/projects/%d/forms/%s/assignments/%d/%d",
		c.config.BaseURL, projectID, formID, appUserRoleID, appUserID)

	req, err := http.NewRequest("POST", assignURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to assign form: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UnassignFormFromAppUser removes form assignment from an app user
func (c *Client) UnassignFormFromAppUser(projectID int, formID string, appUserID int) error {
	if err := c.authenticate(); err != nil {
		return err
	}

	const appUserRoleID = 2

	unassignURL := fmt.Sprintf("%s/v1/projects/%d/forms/%s/assignments/%d/%d",
		c.config.BaseURL, projectID, formID, appUserRoleID, appUserID)

	req, err := http.NewRequest("DELETE", unassignURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to unassign form: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetAppUserFormAssignments lists all form assignments for an app user
func (c *Client) GetAppUserFormAssignments(projectID, appUserID int) ([]FormAssignment, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	// Get all forms and check assignments
	forms, err := c.ListProjectForms(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list forms: %w", err)
	}

	var assignments []FormAssignment
	for _, form := range forms {
		// Check if app user is assigned to this form
		assignmentsURL := fmt.Sprintf("%s/v1/projects/%d/forms/%s/assignments",
			c.config.BaseURL, projectID, form.XmlFormId)

		req, err := http.NewRequest("GET", assignmentsURL, nil)
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

		var formAssignments []struct {
			ActorID int `json:"actorId"`
			RoleID  int `json:"roleId"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&formAssignments); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		for _, fa := range formAssignments {
			if fa.ActorID == appUserID {
				assignments = append(assignments, FormAssignment{
					FormID:   form.XmlFormId,
					FormName: form.Name,
					RoleID:   fa.RoleID,
				})
				break
			}
		}
	}

	return assignments, nil
}

// FormAssignment represents a form assignment for an app user
type FormAssignment struct {
	FormID   string `json:"formId"`
	FormName string `json:"formName"`
	RoleID   int    `json:"roleId"`
}

// GenerateQRCodeData generates the QR code configuration data for ODK Collect
// Returns the JSON string that should be encoded into a QR code
func GenerateQRCodeData(baseURL string, projectID int, token string) string {
	// ODK Collect QR code format
	config := map[string]interface{}{
		"general": map[string]interface{}{
			"server_url": fmt.Sprintf("%s/v1/key/%s/projects/%d", baseURL, token, projectID),
		},
		"admin": map[string]interface{}{},
	}

	data, _ := json.Marshal(config)
	return string(data)
}
