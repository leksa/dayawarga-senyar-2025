package service

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
)

// FeedMappingResult contains the mapped feed and its photos
type FeedMappingResult struct {
	Feed        *model.Feed
	Photos      []FeedPhotoInfo
	RelasiType  string // "posko", "faskes", or "" (lapor situasi bebas)
	RelasiName  string // nama posko/faskes yang dipilih
	RelasiID    string // sel_posko atau sel_faskes value
}

// FeedPhotoInfo contains photo information extracted from ODK submission
type FeedPhotoInfo struct {
	PhotoType string
	Filename  string
}

// MapFeedSubmission converts an ODK feed submission to a Feed model
func MapFeedSubmission(submission map[string]interface{}) (*model.Feed, error) {
	feed := &model.Feed{
		Category: "informasi", // default
	}

	// Extract __id as ODK submission ID
	if id, ok := submission["__id"].(string); ok {
		feed.ODKSubmissionID = &id
	}

	// Extract system metadata
	if system, ok := submission["__system"].(map[string]interface{}); ok {
		if submitterName, ok := system["submitterName"].(string); ok {
			feed.Username = &submitterName
		}
		if submittedAt, ok := system["submissionDate"].(string); ok {
			if t, err := time.Parse(time.RFC3339, submittedAt); err == nil {
				feed.SubmittedAt = &t
			}
		}
	}

	// Extract grp_update fields
	if grpUpdate, ok := submission["grp_update"].(map[string]interface{}); ok {
		// Kategori -> Category
		if kategori, ok := grpUpdate["kategori"].(string); ok {
			feed.Category = kategori
		}

		// Tags -> Type (stored as comma-separated or space-separated string)
		if tags, ok := grpUpdate["tags"].(string); ok && tags != "" {
			feed.Type = &tags
		}

		// Deskripsi -> Content
		if deskripsi, ok := grpUpdate["deskripsi"].(string); ok {
			feed.Content = deskripsi
		}
	}

	// Extract location_id from calc_location_id (posko UUID)
	if locationID, ok := submission["calc_location_id"].(string); ok && locationID != "" {
		if uid, err := uuid.Parse(locationID); err == nil {
			feed.LocationID = &uid
		}
	}

	// Extract faskes_id from calc_faskes_id (faskes UUID)
	if faskesID, ok := submission["calc_faskes_id"].(string); ok && faskesID != "" {
		if uid, err := uuid.Parse(faskesID); err == nil {
			feed.FaskesID = &uid
		}
	}

	// Extract coordinates - priority: free geolocation (koordinat) > posko location (calc_geometry)
	// First try grp_relasi.koordinat (free geolocation from user when relasi='no')
	if grpRelasi, ok := submission["grp_relasi"].(map[string]interface{}); ok {
		// Try GeoJSON format first (ODK geopoint type)
		if koordinat, ok := grpRelasi["koordinat"].(map[string]interface{}); ok {
			if coords, ok := koordinat["coordinates"].([]interface{}); ok && len(coords) >= 2 {
				// GeoJSON format: [longitude, latitude]
				if lon, ok := coords[0].(float64); ok {
					feed.Longitude = &lon
				}
				if lat, ok := coords[1].(float64); ok {
					feed.Latitude = &lat
				}
			}
		} else if koordinatStr, ok := grpRelasi["koordinat"].(string); ok && koordinatStr != "" {
			// Fallback to string format "lat lon"
			coordParts := strings.Fields(koordinatStr)
			if len(coordParts) >= 2 {
				if lat, err := strconv.ParseFloat(coordParts[0], 64); err == nil {
					feed.Latitude = &lat
				}
				if lon, err := strconv.ParseFloat(coordParts[1], 64); err == nil {
					feed.Longitude = &lon
				}
			}
		}
	}

	// Fallback to calc_geometry if no free geolocation
	if feed.Latitude == nil || feed.Longitude == nil {
		if geom, ok := submission["calc_geometry"].(string); ok && geom != "" {
			coords := strings.Fields(geom)
			if len(coords) >= 2 {
				if lat, err := strconv.ParseFloat(coords[0], 64); err == nil {
					feed.Latitude = &lat
				}
				if lon, err := strconv.ParseFloat(coords[1], 64); err == nil {
					feed.Longitude = &lon
				}
			}
		}
	}

	// Store raw submission data
	feed.RawData = model.JSONB(submission)

	return feed, nil
}

// MapFeedSubmissionWithPhotos converts an ODK feed submission to a Feed model with photos
func MapFeedSubmissionWithPhotos(submission map[string]interface{}) (*FeedMappingResult, error) {
	feed, err := MapFeedSubmission(submission)
	if err != nil {
		return nil, err
	}

	result := &FeedMappingResult{
		Feed:   feed,
		Photos: []FeedPhotoInfo{},
	}

	// Extract photos from grp_update.foto
	if grpUpdate, ok := submission["grp_update"].(map[string]interface{}); ok {
		if foto, ok := grpUpdate["foto"].(string); ok && foto != "" {
			result.Photos = append(result.Photos, FeedPhotoInfo{
				PhotoType: "foto",
				Filename:  foto,
			})
		}
	}

	// Extract relasi info (posko OR faskes, mutually exclusive)
	// Form v3 uses jenis_relasi to determine type, v2 uses sel_posko/sel_faskes directly
	if grpRelasi, ok := submission["grp_relasi"].(map[string]interface{}); ok {
		// Check if relasi is enabled (relasi='yes')
		relasi, _ := grpRelasi["relasi"].(string)

		if relasi == "yes" {
			// Check jenis_relasi first (form v3)
			jenisRelasi, _ := grpRelasi["jenis_relasi"].(string)

			// Check for posko selection
			if selPosko, ok := grpRelasi["sel_posko"].(string); ok && selPosko != "" {
				result.RelasiType = "posko"
				result.RelasiID = selPosko

				// Get posko name from calc_nama_posko
				if namaPosko, ok := submission["calc_nama_posko"].(string); ok {
					result.RelasiName = namaPosko
				}

				// Set location_id from sel_posko UUID (may already be set from calc_location_id)
				if feed.LocationID == nil {
					if uid, err := uuid.Parse(selPosko); err == nil {
						feed.LocationID = &uid
					}
				}
			} else if selFaskes, ok := grpRelasi["sel_faskes"].(string); ok && selFaskes != "" {
				// Check for faskes selection
				result.RelasiType = "faskes"
				result.RelasiID = selFaskes

				// Get faskes name from calc_nama_faskes
				if namaFaskes, ok := submission["calc_nama_faskes"].(string); ok {
					result.RelasiName = namaFaskes
				}

				// Set faskes_id from sel_faskes UUID (may already be set from calc_faskes_id)
				if feed.FaskesID == nil {
					if uid, err := uuid.Parse(selFaskes); err == nil {
						feed.FaskesID = &uid
					}
				}
			} else if jenisRelasi != "" {
				// Form v3: jenis_relasi is set but no selection made yet
				result.RelasiType = jenisRelasi
			}
		}
		// If relasi='no', it's a free geolocation report (lapor situasi bebas)
		// RelasiType remains empty string
	}

	return result, nil
}

// ExtractFeedPhotos extracts photo filenames from ODK submission
func ExtractFeedPhotos(submission map[string]interface{}) []FeedPhotoInfo {
	var photos []FeedPhotoInfo

	// Extract photos from grp_update.foto
	if grpUpdate, ok := submission["grp_update"].(map[string]interface{}); ok {
		if foto, ok := grpUpdate["foto"].(string); ok && foto != "" {
			photos = append(photos, FeedPhotoInfo{
				PhotoType: "foto",
				Filename:  foto,
			})
		}
	}

	return photos
}
