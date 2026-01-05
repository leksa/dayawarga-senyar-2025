package service

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
)

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

	// Extract coordinates from calc_geometry ("lat lon" format)
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

	// Store raw submission data
	feed.RawData = model.JSONB(submission)

	return feed, nil
}
