package service

import (
	"strconv"
	"time"

	"github.com/leksa/datamapper-senyar/internal/model"
)

// InfrastrukturPhotoInfo holds photo information for infrastructure
type InfrastrukturPhotoInfo struct {
	PhotoType string
	Filename  string
}

// MapSubmissionToInfrastruktur converts an ODK submission to an Infrastruktur model
func MapSubmissionToInfrastruktur(submission map[string]interface{}) (*model.Infrastruktur, error) {
	infra := &model.Infrastruktur{}

	// Extract __id as ODK submission ID
	if id, ok := submission["__id"].(string); ok {
		infra.ODKSubmissionID = &id
	}

	// Extract grp_identifikasi group first - this contains entity selection and calculated fields
	grpIdentifikasi, _ := submission["grp_identifikasi"].(map[string]interface{})

	// Extract entity selection (sel_jembatan refers to entity 'nama' field which is UUID)
	// Check in grp_identifikasi first, then root
	if grpIdentifikasi != nil {
		if selJembatan, ok := grpIdentifikasi["sel_jembatan"].(string); ok {
			infra.EntityID = selJembatan
		}
	}
	if infra.EntityID == "" {
		if selJembatan, ok := submission["sel_jembatan"].(string); ok {
			infra.EntityID = selJembatan
		}
	}

	// Helper to get string from grpIdentifikasi or root
	getString := func(key string) string {
		if grpIdentifikasi != nil {
			if v, ok := grpIdentifikasi[key].(string); ok && v != "" {
				return v
			}
		}
		if v, ok := submission[key].(string); ok {
			return v
		}
		return ""
	}

	// Basic info from entity (calculated fields)
	infra.Nama = getString("c_nama")
	infra.ObjectID = getString("c_objectid")
	infra.Jenis = getString("c_jenis")
	infra.StatusJln = getString("c_statusjln")
	infra.NamaKabupaten = getString("c_kabupaten")
	infra.NamaProvinsi = getString("c_provinsi")
	infra.TargetSelesai = getString("c_target_selesai")

	// Extract coordinates from entity
	if latStr := getString("c_latitude"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			infra.Latitude = &lat
		}
	}
	if lngStr := getString("c_longitude"); lngStr != "" {
		if lng, err := strconv.ParseFloat(lngStr, 64); err == nil {
			infra.Longitude = &lng
		}
	}

	// Status fields from form input (grp_status)
	grpStatus, _ := submission["grp_status"].(map[string]interface{})
	if grpStatus != nil {
		if statusAkses, ok := grpStatus["status_akses"].(string); ok {
			infra.StatusAkses = statusAkses
		}
		if keterangan, ok := grpStatus["keterangan_bencana"].(string); ok {
			infra.KeteranganBencana = keterangan
		}
		if dampak, ok := grpStatus["dampak"].(string); ok {
			infra.Dampak = dampak
		}
	} else {
		// Try flat structure
		if statusAkses, ok := submission["status_akses"].(string); ok {
			infra.StatusAkses = statusAkses
		}
		if keterangan, ok := submission["keterangan_bencana"].(string); ok {
			infra.KeteranganBencana = keterangan
		}
		if dampak, ok := submission["dampak"].(string); ok {
			infra.Dampak = dampak
		}
	}

	// Penanganan fields (grp_penanganan)
	grpPenanganan, _ := submission["grp_penanganan"].(map[string]interface{})
	if grpPenanganan != nil {
		if status, ok := grpPenanganan["status_penanganan"].(string); ok {
			infra.StatusPenanganan = status
		}
		if detail, ok := grpPenanganan["penanganan_detail"].(string); ok {
			infra.PenangananDetail = detail
		}
		if bailey, ok := grpPenanganan["bailey"].(string); ok {
			infra.Bailey = bailey
		}
		if progress, ok := grpPenanganan["progress"].(string); ok {
			if p, err := strconv.Atoi(progress); err == nil {
				infra.Progress = p
			}
		}
	} else {
		// Try flat structure
		if status, ok := submission["status_penanganan"].(string); ok {
			infra.StatusPenanganan = status
		}
		if detail, ok := submission["penanganan_detail"].(string); ok {
			infra.PenangananDetail = detail
		}
		if bailey, ok := submission["bailey"].(string); ok {
			infra.Bailey = bailey
		}
		if progress, ok := submission["progress"].(string); ok {
			if p, err := strconv.Atoi(progress); err == nil {
				infra.Progress = p
			}
		}
	}

	// Source info
	if baseline, ok := submission["baseline_sumber"].(string); ok {
		infra.BaselineSumber = baseline
	} else {
		infra.BaselineSumber = "BNPB/PU"
	}

	// Extract system metadata and use submitterName as update_by
	if system, ok := submission["__system"].(map[string]interface{}); ok {
		if submitterName, ok := system["submitterName"].(string); ok {
			infra.SubmitterName = &submitterName
			// Use submitter name as update_by (who updated the data)
			infra.UpdateBy = submitterName
		}
		if submittedAt, ok := system["submissionDate"].(string); ok {
			if t, err := time.Parse(time.RFC3339, submittedAt); err == nil {
				infra.SubmittedAt = &t
			}
		}
	}

	// Fallback to update_by field if submitter not available
	if infra.UpdateBy == "" {
		if updateBy, ok := submission["update_by"].(string); ok {
			infra.UpdateBy = updateBy
		}
	}

	// Store raw data
	infra.RawData = model.JSONB(submission)

	return infra, nil
}

// ExtractInfrastrukturPhotos extracts photo information from an ODK submission
func ExtractInfrastrukturPhotos(submission map[string]interface{}) []InfrastrukturPhotoInfo {
	var photos []InfrastrukturPhotoInfo

	// Check grp_foto group first
	grpFoto, _ := submission["grp_foto"].(map[string]interface{})
	if grpFoto != nil {
		for i := 1; i <= 4; i++ {
			fieldName := "foto_" + strconv.Itoa(i)
			if filename, ok := grpFoto[fieldName].(string); ok && filename != "" {
				photos = append(photos, InfrastrukturPhotoInfo{
					PhotoType: fieldName,
					Filename:  filename,
				})
			}
		}
	} else {
		// Try flat structure
		for i := 1; i <= 4; i++ {
			fieldName := "foto_" + strconv.Itoa(i)
			if filename, ok := submission[fieldName].(string); ok && filename != "" {
				photos = append(photos, InfrastrukturPhotoInfo{
					PhotoType: fieldName,
					Filename:  filename,
				})
			}
		}
	}

	return photos
}
