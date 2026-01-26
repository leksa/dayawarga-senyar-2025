package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
	"github.com/leksa/datamapper-senyar/internal/service"
)

// ODKHandler handles ODK-related API endpoints
type ODKHandler struct {
	projectService *service.ODKProjectService
	requestService *service.ProjectRequestService
}

// NewODKHandler creates a new ODK handler
func NewODKHandler(projectService *service.ODKProjectService, requestService *service.ProjectRequestService) *ODKHandler {
	return &ODKHandler{
		projectService: projectService,
		requestService: requestService,
	}
}

// ListProjects lists all available ODK projects
// GET /api/v1/odk/projects
func (h *ODKHandler) ListProjects(c *gin.Context) {
	ctx := c.Request.Context()

	projects, err := h.projectService.ListProjects(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch ODK projects: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    projects,
	})
}

// GetProject gets a specific ODK project
// GET /api/v1/odk/projects/:id
func (h *ODKHandler) GetProject(c *gin.Context) {
	ctx := c.Request.Context()

	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	project, err := h.projectService.GetProject(ctx, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch project: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    project,
	})
}

// ListProjectForms lists all forms in an ODK project
// GET /api/v1/odk/projects/:id/forms
func (h *ODKHandler) ListProjectForms(c *gin.Context) {
	ctx := c.Request.Context()

	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	forms, err := h.projectService.ListProjectForms(ctx, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch forms: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    forms,
	})
}

// CreateProjectRequestInput represents input for creating a project request
type CreateProjectRequestInput struct {
	ODKProjectID int     `json:"odk_project_id" binding:"required"`
	Notes        *string `json:"notes"`
}

// CreateProjectRequest creates a new project request for a group
// POST /api/v1/groups/:id/project-request
func (h *ODKHandler) CreateProjectRequest(c *gin.Context) {
	ctx := c.Request.Context()

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid group ID",
		})
		return
	}

	var input CreateProjectRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	request, err := h.requestService.CreateRequest(ctx, userID.(uuid.UUID), service.CreateProjectRequestInput{
		GroupID:      groupID,
		ODKProjectID: input.ODKProjectID,
		Notes:        input.Notes,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    request,
		"message": "Project request created successfully",
	})
}

// GetGroupProjectRequests gets project requests for a specific group
// GET /api/v1/groups/:id/project-requests
func (h *ODKHandler) GetGroupProjectRequests(c *gin.Context) {
	ctx := c.Request.Context()

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid group ID",
		})
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := repository.ProjectRequestFilter{
		GroupID: &groupID,
	}

	requests, total, err := h.requestService.ListRequests(ctx, filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch requests: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    requests,
		"meta": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// ListProjectRequests lists project requests
// GET /api/v1/admin/project-requests
func (h *ODKHandler) ListProjectRequests(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Parse filters
	var filter repository.ProjectRequestFilter

	if orgID := c.Query("organization_id"); orgID != "" {
		if id, err := uuid.Parse(orgID); err == nil {
			filter.OrganizationID = &id
		}
	}

	if status := c.Query("status"); status != "" {
		s := model.ProjectRequestStatus(status)
		filter.Status = &s
	}

	requests, total, err := h.requestService.ListRequests(ctx, filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch requests: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    requests,
		"meta": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetProjectRequest gets a specific project request
// GET /api/v1/admin/project-requests/:id
func (h *ODKHandler) GetProjectRequest(c *gin.Context) {
	ctx := c.Request.Context()

	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request ID",
		})
		return
	}

	request, err := h.requestService.GetRequest(ctx, requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Request not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    request,
	})
}

// ReviewProjectRequestInput represents input for reviewing a request
type ReviewProjectRequestInput struct {
	Action string  `json:"action" binding:"required,oneof=approve reject"`
	Notes  *string `json:"notes"`
}

// ReviewProjectRequest approves or rejects a project request
// PUT /api/v1/admin/project-requests/:id
func (h *ODKHandler) ReviewProjectRequest(c *gin.Context) {
	ctx := c.Request.Context()

	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request ID",
		})
		return
	}

	var input ReviewProjectRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var request *model.ProjectRequest

	switch input.Action {
	case "approve":
		request, err = h.requestService.ApproveRequest(ctx, requestID, userID.(uuid.UUID), service.ApproveRequestInput{
			Notes: input.Notes,
		})
	case "reject":
		request, err = h.requestService.RejectRequest(ctx, requestID, userID.(uuid.UUID), service.RejectRequestInput{
			Notes: input.Notes,
		})
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	action := "approved"
	if input.Action == "reject" {
		action = "rejected"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    request,
		"message": "Project request " + action + " successfully",
	})
}
