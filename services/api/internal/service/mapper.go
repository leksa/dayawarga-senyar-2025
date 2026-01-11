package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/leksa/datamapper-senyar/internal/model"
)

// MapSubmissionToLocation converts an ODK submission to a Location model
// Uses final_* calculated fields from XLSForm v2, with fallback to nested grp_* fields for dump data
func MapSubmissionToLocation(submission map[string]interface{}) (*model.Location, error) {
	location := &model.Location{
		Type:   "posko",
		Status: "operational",
	}

	// Extract nested groups for fallback (dump data doesn't have final_* fields)
	grpIdentitas, _ := submission["grp_identitas"].(map[string]interface{})
	grpDemografi, _ := submission["grp_demografi"].(map[string]interface{})
	grpPengungsian, _ := submission["grp_pengungsian"].(map[string]interface{})
	grpFasilitas, _ := submission["grp_fasilitas"].(map[string]interface{})
	grpKomunikasi, _ := submission["grp_komunikasi"].(map[string]interface{})
	grpAkses, _ := submission["grp_akses"].(map[string]interface{})
	grpBaseline, _ := submission["grp_baseline"].(map[string]interface{})

	// Extract __id as ODK submission ID
	if id, ok := submission["__id"].(string); ok {
		location.ODKSubmissionID = &id
	}

	// Extract nama from calc_nama_posko or nama_posko (fallback for data dump)
	if nama, ok := submission["calc_nama_posko"].(string); ok && nama != "" {
		location.Nama = nama
	} else if nama, ok := submission["nama_posko"].(string); ok && nama != "" {
		location.Nama = nama
	}

	// Extract system metadata
	if system, ok := submission["__system"].(map[string]interface{}); ok {
		if submitterName, ok := system["submitterName"].(string); ok {
			location.SubmitterName = &submitterName
		}
		if submittedAt, ok := system["submissionDate"].(string); ok {
			if t, err := time.Parse(time.RFC3339, submittedAt); err == nil {
				location.SubmittedAt = &t
			}
		}
	}

	// Extract coordinates - try final_geometry first, then grp_identitas.koordinat
	if geom, ok := submission["final_geometry"].(string); ok && geom != "" {
		coords := strings.Fields(geom)
		if len(coords) >= 2 {
			if lat, err := strconv.ParseFloat(coords[0], 64); err == nil {
				location.Latitude = &lat
			}
			if lon, err := strconv.ParseFloat(coords[1], 64); err == nil {
				location.Longitude = &lon
			}
		}
	} else if grpIdentitas != nil {
		// Fallback: try koordinat from grp_identitas (can be GeoJSON or string)
		if koordinat, ok := grpIdentitas["koordinat"].(map[string]interface{}); ok {
			// GeoJSON format: {"type": "Point", "coordinates": [lon, lat, alt]}
			if coords, ok := koordinat["coordinates"].([]interface{}); ok && len(coords) >= 2 {
				if lon, ok := coords[0].(float64); ok {
					location.Longitude = &lon
				}
				if lat, ok := coords[1].(float64); ok {
					location.Latitude = &lat
				}
			}
		} else if koordinatStr, ok := grpIdentitas["koordinat"].(string); ok && koordinatStr != "" {
			// String format: "lat lon alt accuracy"
			coords := strings.Fields(koordinatStr)
			if len(coords) >= 2 {
				if lat, err := strconv.ParseFloat(coords[0], 64); err == nil {
					location.Latitude = &lat
				}
				if lon, err := strconv.ParseFloat(coords[1], 64); err == nil {
					location.Longitude = &lon
				}
			}
		}
	}

	// Build Alamat JSONB (codes and names)
	location.Alamat = model.JSONB{
		"id_provinsi":     getStringValue(submission, "sel_provinsi"),
		"id_kota_kab":     getStringValue(submission, "sel_kota_kab"),
		"id_kecamatan":    getStringValue(submission, "sel_kecamatan"),
		"id_desa":         getStringValue(submission, "sel_desa"),
		"nama_provinsi":   getStringValue(submission, "calc_nama_provinsi"),
		"nama_kota_kab":   getStringValue(submission, "calc_nama_kota_kab"),
		"nama_kecamatan":  getStringValue(submission, "calc_nama_kecamatan"),
		"nama_desa":       getStringValue(submission, "calc_nama_desa"),
	}

	// Build Identitas JSONB - try final_* first, fallback to grp_identitas
	location.Identitas = model.JSONB{
		"nama_penanggungjawab":    getWithFallback(submission, "final_nama_penanggungjawab", grpIdentitas, "nama_penanggungjawab"),
		"contact_penanggungjawab": getWithFallback(submission, "final_contact_penanggungjawab", grpIdentitas, "contact_penanggungjawab"),
		"nama_relawan":            getWithFallback(submission, "final_nama_relawan", grpIdentitas, "nama_relawan"),
		"contact_relawan":         getWithFallback(submission, "final_contact_relawan", grpIdentitas, "contact_relawan"),
		"alamat_dusun":            getWithFallback(submission, "final_alamat_dusun", grpIdentitas, "alamat_dusun"),
		"institusi":               getWithFallback(submission, "final_institusi", grpIdentitas, "institusi"),
		"mulai_tanggal":           getWithFallback(submission, "final_mulai_tanggal", grpIdentitas, "mulai_tanggal"),
		"kota_terdekat":           getWithFallback(submission, "final_kota_terdekat", grpIdentitas, "kota_terdekat"),
		"baseline_sumber":         getWithFallback(submission, "final_baseline_sumber", grpBaseline, "baseline_sumber"),
	}

	// Extract status_posko - try final_status_posko first, fallback to grp_identitas
	if statusPosko, ok := submission["final_status_posko"].(string); ok && statusPosko != "" {
		location.Status = statusPosko
	} else if grpIdentitas != nil {
		if statusPosko, ok := grpIdentitas["status_posko"].(string); ok && statusPosko != "" {
			location.Status = statusPosko
		}
	}

	// Build DataPengungsi JSONB - try final_* first, fallback to grp_pengungsian and grp_demografi
	totalPengungsi := getWithFallback(submission, "final_total_pengungsi", grpPengungsian, "total_pengungsi")
	dataPengungsi := model.JSONB{
		"jenis_pengungsian":   getWithFallback(submission, "final_jenis_pengungsian", grpPengungsian, "jenis_pengungsian"),
		"detail_pengungsian":  getWithFallback(submission, "final_detail_pengungsian", grpPengungsian, "detail_pengungsian"),
		"persen_keterlibatan": getWithFallback(submission, "final_persen_keterlibatan", grpPengungsian, "persen_keterlibatan"),
		"total_pengungsi":     totalPengungsi,
		"total_jiwa":          totalPengungsi, // alias
		"jumlah_kk":           getWithFallback(submission, "final_jumlah_kk", grpDemografi, "jumlah_kk"),
		"kk_perempuan":        getWithFallback(submission, "final_kk_perempuan", grpDemografi, "kk_perempuan"),
		"kk_anak":             getWithFallback(submission, "final_kk_anak", grpDemografi, "kk_anak"),
		"dewasa_perempuan":    getWithFallback(submission, "final_dewasa_perempuan", grpDemografi, "dewasa_perempuan"),
		"dewasa_laki":         getWithFallback(submission, "final_dewasa_laki", grpDemografi, "dewasa_laki"),
		"remaja_perempuan":    getWithFallback(submission, "final_remaja_perempuan", grpDemografi, "remaja_perempuan"),
		"remaja_laki":         getWithFallback(submission, "final_remaja_laki", grpDemografi, "remaja_laki"),
		"anak_perempuan":      getWithFallback(submission, "final_anak_perempuan", grpDemografi, "anak_perempuan"),
		"anak_laki":           getWithFallback(submission, "final_anak_laki", grpDemografi, "anak_laki"),
		"balita_perempuan":    getWithFallback(submission, "final_balita_perempuan", grpDemografi, "balita_perempuan"),
		"balita_laki":         getWithFallback(submission, "final_balita_laki", grpDemografi, "balita_laki"),
		"bayi_perempuan":      getWithFallback(submission, "final_bayi_perempuan", grpDemografi, "bayi_perempuan"),
		"bayi_laki":           getWithFallback(submission, "final_bayi_laki", grpDemografi, "bayi_laki"),
		"lansia":              getWithFallback(submission, "final_lansia", grpDemografi, "lansia"),
		"ibu_menyusui":        getWithFallback(submission, "final_ibu_menyusui", grpDemografi, "ibu_menyusui"),
		"ibu_hamil":           getWithFallback(submission, "final_ibu_hamil", grpDemografi, "ibu_hamil"),
		"remaja_tanpa_ortu":   getWithFallback(submission, "final_remaja_tanpa_ortu", grpDemografi, "remaja_tanpa_ortu"),
		"anak_tanpa_ortu":     getWithFallback(submission, "final_anak_tanpa_ortu", grpDemografi, "anak_tanpa_ortu"),
		"bayi_tanpa_ibu":      getWithFallback(submission, "final_bayi_tanpa_ibu", grpDemografi, "bayi_tanpa_ibu"),
		"difabel":             getWithFallback(submission, "final_difabel", grpDemografi, "difabel"),
		"komorbid":            getWithFallback(submission, "final_komorbid", grpDemografi, "komorbid"),
	}
	location.DataPengungsi = dataPengungsi

	// Build Fasilitas JSONB - try final_* first, fallback to grp_fasilitas
	location.Fasilitas = model.JSONB{
		"posko_logistik":      getWithFallback(submission, "final_posko_logistik", grpFasilitas, "posko_logistik"),
		"posko_faskes":        getWithFallback(submission, "final_posko_faskes", grpFasilitas, "posko_faskes"),
		"dapur_umum":          getWithFallback(submission, "final_dapur_umum", grpFasilitas, "dapur_umum"),
		"kapasitas_dapur":     getWithFallback(submission, "final_kapasitas_dapur", grpFasilitas, "kapasitas_dapur"),
		"ketersediaan_air":    getWithFallback(submission, "final_ketersediaan_air", grpFasilitas, "ketersediaan_air"),
		"kebutuhan_air":       submission["kebutuhan_air"], // root level calculated field
		"saluran_limbah":      getWithFallback(submission, "final_saluran_limbah", grpFasilitas, "saluran_limbah"),
		"sumber_air":          getWithFallback(submission, "final_sumber_air", grpFasilitas, "sumber_air"),
		"toilet_perempuan":    getWithFallback(submission, "final_toilet_perempuan", grpFasilitas, "toilet_perempuan"),
		"toilet_laki":         getWithFallback(submission, "final_toilet_laki", grpFasilitas, "toilet_laki"),
		"toilet_campur":       getWithFallback(submission, "final_toilet_campur", grpFasilitas, "toilet_campur"),
		"tempat_sampah":       getWithFallback(submission, "final_tempat_sampah", grpFasilitas, "tempat_sampah"),
		"sumber_listrik":      getWithFallback(submission, "final_sumber_listrik", grpFasilitas, "sumber_listrik"),
		"kondisi_penerangan":  getWithFallback(submission, "final_kondisi_penerangan", grpFasilitas, "kondisi_penerangan"),
		"titik_akses_listrik": getWithFallback(submission, "final_titik_akses_listrik", grpFasilitas, "titik_akses_listrik"),
		"posko_tenaga_medis":  getWithFallback(submission, "final_posko_kesehatan", grpFasilitas, "posko_kesehatan"),
		"posko_obat":          getWithFallback(submission, "final_posko_obat", grpFasilitas, "posko_obat"),
		"posko_psikososial":   getWithFallback(submission, "final_posko_psikososial", grpFasilitas, "posko_psikososial"),
		"ruang_laktasi":       getWithFallback(submission, "final_ruang_laktasi", grpFasilitas, "ruang_laktasi"),
		"layanan_lansia":      getWithFallback(submission, "final_layanan_lansia", grpFasilitas, "layanan_lansia"),
		"layanan_keluarga":    getWithFallback(submission, "final_layanan_keluarga", grpFasilitas, "layanan_keluarga"),
		"sekolah_darurat":     getWithFallback(submission, "final_sekolah_darurat", grpFasilitas, "sekolah_darurat"),
		"program_pengganti":   getWithFallback(submission, "final_program_pengganti", grpFasilitas, "program_pengganti"),
		"petugas_keamanan":    getWithFallback(submission, "final_petugas_keamanan", grpFasilitas, "petugas_keamanan"),
		"area_interaksi":      getWithFallback(submission, "final_area_interaksi", grpFasilitas, "area_interaksi"),
		"area_bermain":        getWithFallback(submission, "final_area_bermain", grpFasilitas, "area_bermain"),
	}

	// Calculate kebutuhan_air from total_pengungsi if not already set
	if location.Fasilitas["kebutuhan_air"] == nil {
		if totalPengungsi != nil {
			if totalFloat, ok := totalPengungsi.(float64); ok {
				location.Fasilitas["kebutuhan_air"] = int(totalFloat) * 15
			}
		}
	}

	// Build Komunikasi JSONB - try final_* first, fallback to grp_komunikasi
	location.Komunikasi = model.JSONB{
		"ketersediaan_sinyal":   getWithFallback(submission, "final_ketersediaan_sinyal", grpKomunikasi, "ketersediaan_sinyal"),
		"jaringan_orari":        getWithFallback(submission, "final_jaringan_orari", grpKomunikasi, "jaringan_orari"),
		"ketersediaan_internet": getWithFallback(submission, "final_ketersediaan_internet", grpKomunikasi, "ketersediaan_internet"),
	}

	// Build Akses JSONB - try final_* first, fallback to grp_akses
	location.Akses = model.JSONB{
		"jarak_pkm":            getWithFallback(submission, "final_jarak_pkm", grpAkses, "jarak_pkm"),
		"jarak_posko_logistik": getWithFallback(submission, "final_jarak_posko_logistik", grpAkses, "jarak_posko_logistik"),
		"nama_faskes_terdekat": getWithFallback(submission, "final_nama_faskes_terdekat", grpAkses, "nama_faskes_terdekat"),
		"terisolir":            getWithFallback(submission, "final_terisolir", grpAkses, "terisolir"),
		"akses_via":            getWithFallback(submission, "final_akses_via", grpAkses, "akses_via"),
	}

	// Store raw submission data
	location.RawData = model.JSONB(submission)

	return location, nil
}

// ExtractPhotos extracts photo information from a submission
func ExtractPhotos(submission map[string]interface{}) []PhotoInfo {
	var photos []PhotoInfo

	grpFoto, ok := submission["grp_foto"].(map[string]interface{})
	if !ok {
		return photos
	}

	photoFields := []struct {
		field     string
		photoType string
	}{
		{"foto_depan", "tampak_depan"},
		{"foto_area1", "area_1"},
		{"foto_area2", "area_2"},
		{"foto_area3", "area_3"},
		{"foto_toilet", "toilet"},
		{"foto_sampah", "sampah"},
		{"foto_faskes", "faskes"},
		{"foto_dapur", "dapur"},
	}

	submissionID := ""
	if id, ok := submission["__id"].(string); ok {
		submissionID = id
	}

	for _, pf := range photoFields {
		if filename, ok := grpFoto[pf.field].(string); ok && filename != "" {
			photos = append(photos, PhotoInfo{
				Filename:     filename,
				PhotoType:    pf.photoType,
				SubmissionID: submissionID,
			})
		}
	}

	return photos
}

// PhotoInfo holds photo metadata
type PhotoInfo struct {
	Filename     string
	PhotoType    string
	SubmissionID string
}

// Helper functions

// getWithFallback tries to get value from primary map first, then falls back to secondary map
func getWithFallback(primary map[string]interface{}, primaryKey string, fallback map[string]interface{}, fallbackKey string) interface{} {
	if val := primary[primaryKey]; val != nil {
		return val
	}
	if fallback != nil {
		return fallback[fallbackKey]
	}
	return nil
}

func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getIntValue(m map[string]interface{}, key string) *int {
	switch v := m[key].(type) {
	case float64:
		i := int(v)
		return &i
	case int:
		return &v
	}
	return nil
}

func getFloatValue(m map[string]interface{}, key string) *float64 {
	if v, ok := m[key].(float64); ok {
		return &v
	}
	return nil
}

// BuildGeomSQL creates PostgreSQL geometry from lat/lon
func BuildGeomSQL(lat, lon float64) string {
	return fmt.Sprintf("ST_SetSRID(ST_MakePoint(%f, %f), 4326)", lon, lat)
}
