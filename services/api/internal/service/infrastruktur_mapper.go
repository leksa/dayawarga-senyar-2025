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

	// Extract entity selection (sel_jembatan refers to entity 'nama' field which is UUID)
	if selJembatan, ok := submission["sel_jembatan"].(string); ok {
		infra.EntityID = selJembatan
	}

	// Extract calculated fields from entity lookup
	grpIdentifikasi, _ := submission["grp_identifikasi"].(map[string]interface{})

	// Basic info from entity (calculated fields)
	if nama, ok := submission["c_nama"].(string); ok && nama != "" {
		infra.Nama = nama
	} else if grpIdentifikasi != nil {
		if nama, ok := grpIdentifikasi["c_nama"].(string); ok {
			infra.Nama = nama
		}
	}

	if objectid, ok := submission["c_objectid"].(string); ok {
		infra.ObjectID = objectid
	}

	if jenis, ok := submission["c_jenis"].(string); ok {
		infra.Jenis = jenis
	}

	if statusjln, ok := submission["c_statusjln"].(string); ok {
		infra.StatusJln = statusjln
	}

	if kabupaten, ok := submission["c_kabupaten"].(string); ok {
		infra.NamaKabupaten = kabupaten
	}

	if provinsi, ok := submission["c_provinsi"].(string); ok {
		infra.NamaProvinsi = provinsi
	}

	if targetSelesai, ok := submission["c_target_selesai"].(string); ok {
		infra.TargetSelesai = targetSelesai
	}

	// Extract coordinates from entity
	if latStr, ok := submission["c_latitude"].(string); ok {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			infra.Latitude = &lat
		}
	}
	if lngStr, ok := submission["c_longitude"].(string); ok {
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

	if updateBy, ok := submission["update_by"].(string); ok {
		infra.UpdateBy = updateBy
	}

	// Extract system metadata
	if system, ok := submission["__system"].(map[string]interface{}); ok {
		if submitterName, ok := system["submitterName"].(string); ok {
			infra.SubmitterName = &submitterName
		}
		if submittedAt, ok := system["submissionDate"].(string); ok {
			if t, err := time.Parse(time.RFC3339, submittedAt); err == nil {
				infra.SubmittedAt = &t
			}
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
