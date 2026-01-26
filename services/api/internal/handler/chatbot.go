package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leksa/datamapper-senyar/internal/dto"
	"github.com/leksa/datamapper-senyar/internal/model"
	"github.com/leksa/datamapper-senyar/internal/repository"
	"gorm.io/gorm"
)

type ChatbotHandler struct {
	db           *gorm.DB
	locationRepo *repository.LocationRepository
	feedRepo     *repository.FeedRepository
}

func NewChatbotHandler(db *gorm.DB, locationRepo *repository.LocationRepository, feedRepo *repository.FeedRepository) *ChatbotHandler {
	return &ChatbotHandler{
		db:           db,
		locationRepo: locationRepo,
		feedRepo:     feedRepo,
	}
}

// --- Request DTOs ---

type CreateFeedRequest struct {
	Content      string      `json:"content" binding:"required"`
	Category     string      `json:"category"`
	Type         string      `json:"type"`
	Username     string      `json:"username"`
	Organization string      `json:"organization"`
	LocationID   string      `json:"location_id"`
	FaskesID     string      `json:"faskes_id"`
	Longitude    *float64    `json:"longitude"`
	Latitude     *float64    `json:"latitude"`
	Source       string      `json:"source"`
	RawData      model.JSONB `json:"raw_data"`
}

type CreateLocationRequest struct {
	Nama          string      `json:"nama" binding:"required"`
	Type          string      `json:"type"`
	Status        string      `json:"status"`
	Longitude     *float64    `json:"longitude"`
	Latitude      *float64    `json:"latitude"`
	Identitas     model.JSONB `json:"identitas"`
	Alamat        model.JSONB `json:"alamat"`
	DataPengungsi model.JSONB `json:"data_pengungsi"`
	Fasilitas     model.JSONB `json:"fasilitas"`
	Komunikasi    model.JSONB `json:"komunikasi"`
	Akses         model.JSONB `json:"akses"`
	SubmitterName string      `json:"submitter_name"`
}

type UpdateLocationRequest struct {
	Nama          *string     `json:"nama"`
	Status        *string     `json:"status"`
	Longitude     *float64    `json:"longitude"`
	Latitude      *float64    `json:"latitude"`
	Identitas     model.JSONB `json:"identitas"`
	Alamat        model.JSONB `json:"alamat"`
	DataPengungsi model.JSONB `json:"data_pengungsi"`
	Fasilitas     model.JSONB `json:"fasilitas"`
	Komunikasi    model.JSONB `json:"komunikasi"`
	Akses         model.JSONB `json:"akses"`
}

type WilayahResponse struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

// --- Feed Endpoints ---

func (h *ChatbotHandler) CreateFeed(c *gin.Context) {
	var req CreateFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	if req.Category == "" {
		req.Category = "informasi"
	}
	if req.Source == "" {
		req.Source = "whatsapp"
	}

	rawData := model.JSONB{"source": req.Source}
	if req.RawData != nil {
		rawData = req.RawData
		rawData["source"] = req.Source
	}
	rawDataJSON, _ := json.Marshal(rawData)

	now := time.Now()

	var query string
	var args []interface{}

	if req.Latitude != nil && req.Longitude != nil {
		query = `INSERT INTO information_feeds (content, category, type, username, organization, geom, raw_data, submitted_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, ST_SetSRID(ST_MakePoint($6, $7), 4326), $8, $9, NOW(), NOW())`
		args = []interface{}{req.Content, req.Category, req.Type, req.Username, req.Organization, *req.Longitude, *req.Latitude, string(rawDataJSON), now}
	} else {
		query = `INSERT INTO information_feeds (content, category, type, username, organization, geom, raw_data, submitted_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NULL, $6, $7, NOW(), NOW())`
		args = []interface{}{req.Content, req.Category, req.Type, req.Username, req.Organization, string(rawDataJSON), now}
	}

	result := h.db.Exec(query, args...)
	if result.Error != nil {
		log.Printf("[Chatbot] CreateFeed error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "DB_ERROR", Message: "Failed to create feed"},
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Data:    gin.H{"message": "Feed created successfully", "source": req.Source},
	})
}

// --- Location Endpoints ---

func (h *ChatbotHandler) CreateLocation(c *gin.Context) {
	var req CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	if req.Type == "" {
		req.Type = "posko"
	}
	if req.Status == "" {
		req.Status = "operasional"
	}

	identitasJSON, _ := json.Marshal(defaultJSONB(req.Identitas))
	alamatJSON, _ := json.Marshal(defaultJSONB(req.Alamat))
	dataPengungsiJSON, _ := json.Marshal(defaultJSONB(req.DataPengungsi))
	fasilitasJSON, _ := json.Marshal(defaultJSONB(req.Fasilitas))
	komunikasiJSON, _ := json.Marshal(defaultJSONB(req.Komunikasi))
	aksesJSON, _ := json.Marshal(defaultJSONB(req.Akses))

	now := time.Now()
	submitterName := req.SubmitterName
	if submitterName == "" {
		submitterName = "Chatbot WhatsApp"
	}

	var query string
	var args []interface{}

	if req.Latitude != nil && req.Longitude != nil {
		query = `INSERT INTO locations (nama, type, status, geom, identitas, alamat, data_pengungsi, fasilitas, komunikasi, akses, submitter_name, baseline_sumber, submitted_at, created_at, updated_at)
			VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326), $6, $7, $8, $9, $10, $11, $12, 'Chatbot WhatsApp', $13, NOW(), NOW())
			RETURNING id`
		args = []interface{}{
			req.Nama, req.Type, req.Status,
			*req.Longitude, *req.Latitude,
			string(identitasJSON), string(alamatJSON), string(dataPengungsiJSON),
			string(fasilitasJSON), string(komunikasiJSON), string(aksesJSON),
			submitterName, now,
		}
	} else {
		query = `INSERT INTO locations (nama, type, status, geom, identitas, alamat, data_pengungsi, fasilitas, komunikasi, akses, submitter_name, baseline_sumber, submitted_at, created_at, updated_at)
			VALUES ($1, $2, $3, NULL, $4, $5, $6, $7, $8, $9, $10, 'Chatbot WhatsApp', $11, NOW(), NOW())
			RETURNING id`
		args = []interface{}{
			req.Nama, req.Type, req.Status,
			string(identitasJSON), string(alamatJSON), string(dataPengungsiJSON),
			string(fasilitasJSON), string(komunikasiJSON), string(aksesJSON),
			submitterName, now,
		}
	}

	var locationID string
	row := h.db.Raw(query, args...).Row()
	if err := row.Scan(&locationID); err != nil {
		log.Printf("[Chatbot] CreateLocation error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "DB_ERROR", Message: "Failed to create location"},
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Data:    gin.H{"message": "Location created successfully", "id": locationID},
	})
}

func (h *ChatbotHandler) UpdateLocation(c *gin.Context) {
	id := c.Param("id")

	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	setClauses := []string{}
	args := []interface{}{}
	paramIdx := 1

	if req.Nama != nil {
		setClauses = append(setClauses, fmt.Sprintf("nama = $%d", paramIdx))
		args = append(args, *req.Nama)
		paramIdx++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", paramIdx))
		args = append(args, *req.Status)
		paramIdx++
	}
	if req.Latitude != nil && req.Longitude != nil {
		setClauses = append(setClauses, fmt.Sprintf("geom = ST_SetSRID(ST_MakePoint($%d, $%d), 4326)", paramIdx, paramIdx+1))
		args = append(args, *req.Longitude, *req.Latitude)
		paramIdx += 2
	}
	if req.Identitas != nil {
		j, _ := json.Marshal(req.Identitas)
		setClauses = append(setClauses, fmt.Sprintf("identitas = $%d", paramIdx))
		args = append(args, string(j))
		paramIdx++
	}
	if req.Alamat != nil {
		j, _ := json.Marshal(req.Alamat)
		setClauses = append(setClauses, fmt.Sprintf("alamat = $%d", paramIdx))
		args = append(args, string(j))
		paramIdx++
	}
	if req.DataPengungsi != nil {
		j, _ := json.Marshal(req.DataPengungsi)
		setClauses = append(setClauses, fmt.Sprintf("data_pengungsi = $%d", paramIdx))
		args = append(args, string(j))
		paramIdx++
	}
	if req.Fasilitas != nil {
		j, _ := json.Marshal(req.Fasilitas)
		setClauses = append(setClauses, fmt.Sprintf("fasilitas = $%d", paramIdx))
		args = append(args, string(j))
		paramIdx++
	}
	if req.Komunikasi != nil {
		j, _ := json.Marshal(req.Komunikasi)
		setClauses = append(setClauses, fmt.Sprintf("komunikasi = $%d", paramIdx))
		args = append(args, string(j))
		paramIdx++
	}
	if req.Akses != nil {
		j, _ := json.Marshal(req.Akses)
		setClauses = append(setClauses, fmt.Sprintf("akses = $%d", paramIdx))
		args = append(args, string(j))
		paramIdx++
	}

	if len(setClauses) == 0 {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: "No fields to update"},
		})
		return
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE locations SET %s WHERE id = $%d AND deleted_at IS NULL",
		joinStrings(setClauses, ", "), paramIdx)
	args = append(args, id)

	result := h.db.Exec(query, args...)
	if result.Error != nil {
		log.Printf("[Chatbot] UpdateLocation error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "DB_ERROR", Message: "Failed to update location"},
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "NOT_FOUND", Message: "Location not found"},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    gin.H{"message": "Location updated successfully"},
	})
}

// --- Wilayah Endpoints ---

func (h *ChatbotHandler) GetWilayahKabupaten(c *gin.Context) {
	var results []WilayahResponse
	h.db.Raw("SELECT kode, nama FROM wilayah_kota_kab ORDER BY nama").Scan(&results)
	c.JSON(http.StatusOK, dto.APIResponse{Success: true, Data: results})
}

func (h *ChatbotHandler) GetWilayahKecamatan(c *gin.Context) {
	kabID := c.Query("kota_kab_id")
	if kabID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: "kota_kab_id query param required"},
		})
		return
	}
	var results []WilayahResponse
	h.db.Raw("SELECT kode, nama FROM wilayah_kecamatan WHERE id_kota_kab = ? ORDER BY nama", kabID).Scan(&results)
	c.JSON(http.StatusOK, dto.APIResponse{Success: true, Data: results})
}

func (h *ChatbotHandler) GetWilayahDesa(c *gin.Context) {
	kecID := c.Query("kecamatan_id")
	kabID := c.Query("kota_kab_id")

	if kecID == "" && kabID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: "kecamatan_id or kota_kab_id query param required"},
		})
		return
	}

	var results []WilayahResponse
	if kecID != "" {
		h.db.Raw("SELECT kode, nama FROM wilayah_desa WHERE id_kec = ? ORDER BY nama", kecID).Scan(&results)
	} else {
		h.db.Raw("SELECT kode, nama FROM wilayah_desa WHERE id_kota_kab = ? ORDER BY nama", kabID).Scan(&results)
	}
	c.JSON(http.StatusOK, dto.APIResponse{Success: true, Data: results})
}

// --- Helpers ---

func defaultJSONB(j model.JSONB) model.JSONB {
	if j == nil {
		return model.JSONB{}
	}
	return j
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
