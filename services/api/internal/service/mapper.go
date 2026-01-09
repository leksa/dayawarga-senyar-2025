package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/leksa/datamapper-senyar/internal/model"
)

// MapSubmissionToLocation converts an ODK submission to a Location model
func MapSubmissionToLocation(submission map[string]interface{}) (*model.Location, error) {
	location := &model.Location{
		Type:   "posko",
		Status: "operational",
	}

	// Extract __id as ODK submission ID
	if id, ok := submission["__id"].(string); ok {
		location.ODKSubmissionID = &id
	}

	// Extract nama from calc_nama_posko
	if nama, ok := submission["calc_nama_posko"].(string); ok {
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

	// Extract coordinates from calc_geometry ("lat lon" format)
	if geom, ok := submission["calc_geometry"].(string); ok {
		coords := strings.Fields(geom)
		if len(coords) >= 2 {
			if lat, err := strconv.ParseFloat(coords[0], 64); err == nil {
				location.Latitude = &lat
			}
			if lon, err := strconv.ParseFloat(coords[1], 64); err == nil {
				location.Longitude = &lon
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

	// Build Identitas JSONB from grp_identitas
	if grpIdentitas, ok := submission["grp_identitas"].(map[string]interface{}); ok {
		location.Identitas = model.JSONB{
			"nama_penanggungjawab":    grpIdentitas["nama_penanggungjawab"],
			"contact_penanggungjawab": grpIdentitas["contact_penanggungjawab"],
			"nama_relawan":            grpIdentitas["nama_relawan"],
			"contact_relawan":         grpIdentitas["contact_relawan"],
			"alamat_dusun":            grpIdentitas["alamat_dusun"],
			"institusi":               grpIdentitas["institusi"],
			"mulai_tanggal":           grpIdentitas["mulai_tanggal"],
			"kota_terdekat":           grpIdentitas["kota_terdekat"],
		}

		// Extract status_posko from grp_identitas
		if statusPosko, ok := grpIdentitas["status_posko"].(string); ok && statusPosko != "" {
			location.Status = statusPosko
		}

		// Extract coordinates from grp_identitas.koordinat if calc_geometry not available
		if location.Latitude == nil || location.Longitude == nil {
			if koordinat, ok := grpIdentitas["koordinat"].(map[string]interface{}); ok {
				if coords, ok := koordinat["coordinates"].([]interface{}); ok && len(coords) >= 2 {
					if lon, ok := coords[0].(float64); ok {
						location.Longitude = &lon
					}
					if lat, ok := coords[1].(float64); ok {
						location.Latitude = &lat
					}
				}
			}
		}
	}

	// Build DataPengungsi JSONB
	dataPengungsi := model.JSONB{}

	// From grp_pengungsian
	if grpPengungsian, ok := submission["grp_pengungsian"].(map[string]interface{}); ok {
		dataPengungsi["jenis_pengungsian"] = grpPengungsian["jenis_pengungsian"]
		dataPengungsi["detail_pengungsian"] = grpPengungsian["detail_pengungsian"]
		dataPengungsi["total_pengungsi"] = grpPengungsian["total_pengungsi"]
		dataPengungsi["persen_keterlibatan"] = grpPengungsian["persen_keterlibatan"]
	}

	// From grp_terisolir
	if grpTerisolir, ok := submission["grp_terisolir"].(map[string]interface{}); ok {
		dataPengungsi["terisolir"] = grpTerisolir["terisolir"]
		dataPengungsi["akses_via"] = grpTerisolir["akses_via"]
	}

	// From grp_demografi (was grp_data_pengungsi in form v1)
	if grpData, ok := submission["grp_demografi"].(map[string]interface{}); ok {
		dataPengungsi["jumlah_kk"] = grpData["jumlah_kk"]
		dataPengungsi["kk_perempuan"] = grpData["kk_perempuan"]
		dataPengungsi["kk_anak"] = grpData["kk_anak"]
		dataPengungsi["dewasa_perempuan"] = grpData["dewasa_perempuan"]
		dataPengungsi["dewasa_laki"] = grpData["dewasa_laki"]
		dataPengungsi["remaja_perempuan"] = grpData["remaja_perempuan"]
		dataPengungsi["remaja_laki"] = grpData["remaja_laki"]
		dataPengungsi["anak_perempuan"] = grpData["anak_perempuan"]
		dataPengungsi["anak_laki"] = grpData["anak_laki"]
		dataPengungsi["balita_perempuan"] = grpData["balita_perempuan"]
		dataPengungsi["balita_laki"] = grpData["balita_laki"]
		dataPengungsi["bayi_perempuan"] = grpData["bayi_perempuan"]
		dataPengungsi["bayi_laki"] = grpData["bayi_laki"]
		dataPengungsi["lansia"] = grpData["lansia"]
		dataPengungsi["ibu_menyusui"] = grpData["ibu_menyusui"]
		dataPengungsi["ibu_hamil"] = grpData["ibu_hamil"]
		dataPengungsi["remaja_tanpa_ortu"] = grpData["remaja_tanpa_ortu"]
		dataPengungsi["anak_tanpa_ortu"] = grpData["anak_tanpa_ortu"]
		dataPengungsi["bayi_tanpa_ibu"] = grpData["bayi_tanpa_ibu"]
		dataPengungsi["difabel"] = grpData["difabel"]
		dataPengungsi["komorbid"] = grpData["komorbid"]
	}
	location.DataPengungsi = dataPengungsi

	// Build Fasilitas JSONB from grp_fasilitas
	if grpFasilitas, ok := submission["grp_fasilitas"].(map[string]interface{}); ok {
		location.Fasilitas = model.JSONB{
			"posko_logistik":      grpFasilitas["posko_logistik"],
			"posko_faskes":        grpFasilitas["posko_faskes"],
			"dapur_umum":          grpFasilitas["dapur_umum"],
			"kapasitas_dapur":     grpFasilitas["kapasitas_dapur"],
			"ketersediaan_air":    grpFasilitas["ketersediaan_air"],
			"saluran_limbah":      grpFasilitas["saluran_limbah"],
			"sumber_air":          grpFasilitas["sumber_air"],
			"toilet_perempuan":    grpFasilitas["toilet_perempuan"],
			"toilet_laki":         grpFasilitas["toilet_laki"],
			"toilet_campur":       grpFasilitas["toilet_campur"],
			"tempat_sampah":       grpFasilitas["tempat_sampah"],
			"sumber_listrik":      grpFasilitas["sumber_listrik"],
			"kondisi_penerangan":  grpFasilitas["kondisi_penerangan"],
			"titik_akses_listrik": grpFasilitas["titik_akses_listrik"],
			"posko_tenaga_medis":  grpFasilitas["posko_kesehatan"],
			"posko_obat":          grpFasilitas["posko_obat"],
			"posko_psikososial":   grpFasilitas["posko_psikososial"],
			"ruang_laktasi":       grpFasilitas["ruang_laktasi"],
			"layanan_lansia":      grpFasilitas["layanan_lansia"],
			"layanan_keluarga":    grpFasilitas["layanan_keluarga"],
			"sekolah_darurat":     grpFasilitas["sekolah_darurat"],
			"program_pengganti":   grpFasilitas["program_pengganti"],
			"petugas_keamanan":    grpFasilitas["petugas_keamanan"],
			"area_interaksi":      grpFasilitas["area_interaksi"],
			"area_bermain":        grpFasilitas["area_bermain"],
		}
	}

	// Build Komunikasi JSONB from grp_komunikasi
	if grpKomunikasi, ok := submission["grp_komunikasi"].(map[string]interface{}); ok {
		location.Komunikasi = model.JSONB{
			"ketersediaan_sinyal":   grpKomunikasi["ketersediaan_sinyal"],
			"jaringan_orari":        grpKomunikasi["jaringan_orari"],
			"ketersediaan_internet": grpKomunikasi["ketersediaan_internet"],
		}
	}

	// Build Akses JSONB from grp_akses
	if grpAkses, ok := submission["grp_akses"].(map[string]interface{}); ok {
		location.Akses = model.JSONB{
			"jarak_pkm":            grpAkses["jarak_pkm"],
			"jarak_posko_logistik": grpAkses["jarak_posko_logistik"],
			"nama_faskes_terdekat": grpAkses["nama_faskes_terdekat"],
			"terisolir":            grpAkses["terisolir"],
			"akses_via":            grpAkses["akses_via"],
		}
	}

	// Extract baseline_sumber from grp_baseline
	if grpBaseline, ok := submission["grp_baseline"].(map[string]interface{}); ok {
		if location.Identitas == nil {
			location.Identitas = model.JSONB{}
		}
		location.Identitas["baseline_sumber"] = grpBaseline["baseline_sumber"]
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
