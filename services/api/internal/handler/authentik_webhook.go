package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/leksa/datamapper-senyar/internal/service"
)

type AuthentikWebhookHandler struct {
	webhookService *service.AuthentikWebhookService
	webhookSecret  string
}

func NewAuthentikWebhookHandler(webhookService *service.AuthentikWebhookService, webhookSecret string) *AuthentikWebhookHandler {
	return &AuthentikWebhookHandler{
		webhookService: webhookService,
		webhookSecret:  webhookSecret,
	}
}

type AuthentikWebhookPayload struct {
	EventType string                 `json:"event_type"`
	Secret    string                 `json:"secret"`
	User      AuthentikUserPayload   `json:"user"`
	Context   map[string]interface{} `json:"context,omitempty"`
	CreatedAt string                 `json:"created_at,omitempty"`
}

type AuthentikUserPayload struct {
	PK       int    `json:"pk"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

func (h *AuthentikWebhookHandler) HandleWebhook(c *gin.Context) {
	var payload AuthentikWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid payload: " + err.Error(),
		})
		return
	}

	// Validate webhook secret from payload body
	// Authentik webhooks don't support custom headers, so we include the secret in the body
	if payload.Secret != h.webhookSecret {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid webhook secret",
		})
		return
	}

	if payload.EventType != "user_created" && payload.EventType != "model_created" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Event type ignored: " + payload.EventType,
		})
		return
	}

	if payload.User.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "User email is required",
		})
		return
	}

	ctx := c.Request.Context()
	result, err := h.webhookService.HandleUserCreated(ctx, service.AuthentikUserCreatedInput{
		AuthentikUserID: payload.User.PK,
		Username:        payload.User.Username,
		Email:           payload.User.Email,
		Name:            payload.User.Name,
		IsActive:        payload.User.IsActive,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "User synced to ODK Central successfully",
	})
}

func (h *AuthentikWebhookHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Authentik webhook endpoint is healthy",
	})
}
