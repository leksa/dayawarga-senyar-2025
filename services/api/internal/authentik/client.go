package authentik

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for Authentik API
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// Config holds Authentik client configuration
type Config struct {
	BaseURL  string // e.g., "http://localhost:9000"
	APIToken string // Authentik API token
}

// NewClient creates a new Authentik API client
func NewClient(cfg *Config) *Client {
	return &Client{
		baseURL:  cfg.BaseURL,
		apiToken: cfg.APIToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// User represents an Authentik user
type User struct {
	PK          int                    `json:"pk"`
	UID         string                 `json:"uid"`
	Username    string                 `json:"username"`
	Name        string                 `json:"name"`
	Email       string                 `json:"email"`
	IsActive    bool                   `json:"is_active"`
	DateJoined  string                 `json:"date_joined"`
	LastLogin   *string                `json:"last_login,omitempty"`
	Groups      []int                  `json:"groups,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	Path        string                 `json:"path,omitempty"`
}

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	Username   string                 `json:"username"`
	Name       string                 `json:"name"`
	Email      string                 `json:"email"`
	IsActive   bool                   `json:"is_active"`
	Path       string                 `json:"path,omitempty"`
	Groups     []int                  `json:"groups,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// RecoveryLinkResponse represents response from recovery endpoint
type RecoveryLinkResponse struct {
	Link string `json:"link"`
}

// CreateUser creates a new user in Authentik
func (c *Client) CreateUser(input CreateUserInput) (*User, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v3/core/users/", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create user: %s (status %d)", string(respBody), resp.StatusCode)
	}

	var user User
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &user, nil
}

// GetUserByEmail finds a user by email
func (c *Client) GetUserByEmail(email string) (*User, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v3/core/users/?email="+email, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user: status %d", resp.StatusCode)
	}

	var result struct {
		Results []User `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Results) == 0 {
		return nil, nil // User not found
	}

	return &result.Results[0], nil
}

// GetRecoveryLink generates a password reset/setup link for a user
func (c *Client) GetRecoveryLink(userPK int) (string, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v3/core/users/%d/recovery/", c.baseURL, userPK), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get recovery link: %s (status %d)", string(respBody), resp.StatusCode)
	}

	var result RecoveryLinkResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Link, nil
}

// SendRecoveryEmail sends a password reset/setup email to a user
// Requires email stage to be configured in Authentik
func (c *Client) SendRecoveryEmail(userPK int, emailStageUUID string) error {
	body, err := json.Marshal(map[string]string{
		"email_stage": emailStageUUID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v3/core/users/%d/recovery_email/", c.baseURL, userPK), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send recovery email: %s (status %d)", string(respBody), resp.StatusCode)
	}

	return nil
}

// SetUserPassword sets a user's password directly
func (c *Client) SetUserPassword(userPK int, password string) error {
	body, err := json.Marshal(map[string]string{
		"password": password,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v3/core/users/%d/set_password/", c.baseURL, userPK), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set password: %s (status %d)", string(respBody), resp.StatusCode)
	}

	return nil
}

// UpdateUser updates user attributes
func (c *Client) UpdateUser(userPK int, updates map[string]interface{}) (*User, error) {
	body, err := json.Marshal(updates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v3/core/users/%d/", c.baseURL, userPK), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update user: %s (status %d)", string(respBody), resp.StatusCode)
	}

	var user User
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &user, nil
}

// DeleteUser deletes a user from Authentik
func (c *Client) DeleteUser(userPK int) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v3/core/users/%d/", c.baseURL, userPK), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete user: %s (status %d)", string(respBody), resp.StatusCode)
	}

	return nil
}
