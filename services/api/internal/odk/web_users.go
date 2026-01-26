package odk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WebUser represents an ODK Central web user (human user with email login)
type WebUser struct {
	ID          int        `json:"id"`
	Type        string     `json:"type"` // "user"
	DisplayName string     `json:"displayName"`
	Email       string     `json:"email"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
}

// WebUserCreateRequest represents request to create a web user
type WebUserCreateRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName,omitempty"`
}

// WebUserCreateWithPasswordRequest represents request to create a web user with password
// When password is provided, the user is immediately active (not pending invitation)
type WebUserCreateWithPasswordRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName,omitempty"`
	Password    string `json:"password,omitempty"`
}

// Role IDs in ODK Central
// These IDs are from the /v1/roles endpoint
const (
	RoleAdmin          = 1 // Administrator (system: "admin")
	RoleAppUser        = 2 // App User (system: "app-user")
	RoleProjectManager = 5 // Project Manager (system: "manager")
	RoleProjectViewer  = 6 // Project Viewer (system: "viewer")
	RoleDataCollector  = 8 // Data Collector (system: "formfill")
)

// ListWebUsers fetches all web users in ODK Central
func (c *Client) ListWebUsers() ([]WebUser, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	usersURL := fmt.Sprintf("%s/v1/users", c.config.BaseURL)

	req, err := http.NewRequest("GET", usersURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var users []WebUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return users, nil
}

// GetWebUserByEmail finds a web user by email
func (c *Client) GetWebUserByEmail(email string) (*WebUser, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	// ODK Central supports filtering users by email via query param
	usersURL := fmt.Sprintf("%s/v1/users?q=%s", c.config.BaseURL, email)

	req, err := http.NewRequest("GET", usersURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var users []WebUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Find exact match
	for _, user := range users {
		if strings.EqualFold(user.Email, email) {
			return &user, nil
		}
	}

	return nil, nil // Not found
}

// CreateWebUser creates a new web user
// Note: This only creates the user, you need to separately assign project roles
func (c *Client) CreateWebUser(email, displayName string) (*WebUser, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	usersURL := fmt.Sprintf("%s/v1/users", c.config.BaseURL)

	createReq := WebUserCreateRequest{
		Email:       email,
		DisplayName: displayName,
	}

	payload, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", usersURL, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var user WebUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

// CreateWebUserWithPassword creates a new web user with a password
// User is immediately active and can login without email verification
func (c *Client) CreateWebUserWithPassword(email, displayName, password string) (*WebUser, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	usersURL := fmt.Sprintf("%s/v1/users", c.config.BaseURL)

	createReq := WebUserCreateWithPasswordRequest{
		Email:       email,
		DisplayName: displayName,
		Password:    password,
	}

	payload, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", usersURL, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var user WebUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

// InitiatePasswordReset sends a password reset email to the user
// This can be used to activate a pending user
func (c *Client) InitiatePasswordReset(email string, invalidate bool) error {
	if err := c.authenticate(); err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/v1/users/reset/initiate", c.config.BaseURL)
	if invalidate {
		resetURL += "?invalidate=true"
	}

	payload := fmt.Sprintf(`{"email":"%s"}`, email)

	req, err := http.NewRequest("POST", resetURL, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initiate password reset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteWebUser deletes a web user from ODK Central
func (c *Client) DeleteWebUser(userID int) error {
	if err := c.authenticate(); err != nil {
		return err
	}

	deleteURL := fmt.Sprintf("%s/v1/users/%d", c.config.BaseURL, userID)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete user failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// AssignProjectRole assigns a role to a user on a project
// roleId should be RoleProjectManager (5) or RoleProjectViewer (6)
func (c *Client) AssignProjectRole(projectID, userID, roleID int) error {
	if err := c.authenticate(); err != nil {
		return err
	}

	assignURL := fmt.Sprintf("%s/v1/projects/%d/assignments/%d/%d",
		c.config.BaseURL, projectID, roleID, userID)

	fmt.Printf("ODK AssignProjectRole URL: %s\n", assignURL)

	req, err := http.NewRequest("POST", assignURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("ODK AssignProjectRole response status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("ODK AssignProjectRole error body: %s\n", string(body))
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// RemoveProjectRole removes a role from a user on a project
func (c *Client) RemoveProjectRole(projectID, userID, roleID int) error {
	if err := c.authenticate(); err != nil {
		return err
	}

	removeURL := fmt.Sprintf("%s/v1/projects/%d/assignments/%d/%d",
		c.config.BaseURL, projectID, roleID, userID)

	req, err := http.NewRequest("DELETE", removeURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetOrCreateWebUser gets an existing user by email or creates a new one
// Also assigns Project Manager role to the specified project
func (c *Client) GetOrCreateProjectManager(projectID int, email, displayName string) (*WebUser, error) {
	// Check if user exists
	user, err := c.GetWebUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Create user if not exists
	if user == nil {
		user, err = c.CreateWebUser(email, displayName)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Assign Project Manager role
	if err := c.AssignProjectRole(projectID, user.ID, RoleProjectManager); err != nil {
		return nil, fmt.Errorf("failed to assign project manager role: %w", err)
	}

	return user, nil
}

// ListProjectAssignments lists all role assignments for a project
func (c *Client) ListProjectAssignments(projectID int) ([]ProjectAssignment, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	assignmentsURL := fmt.Sprintf("%s/v1/projects/%d/assignments", c.config.BaseURL, projectID)

	req, err := http.NewRequest("GET", assignmentsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assignments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var assignments []ProjectAssignment
	if err := json.NewDecoder(resp.Body).Decode(&assignments); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return assignments, nil
}

// ProjectAssignment represents a role assignment on a project
type ProjectAssignment struct {
	ActorID int `json:"actorId"`
	RoleID  int `json:"roleId"`
}
