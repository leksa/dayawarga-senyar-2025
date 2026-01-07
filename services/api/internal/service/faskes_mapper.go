package service

import (
	"strconv"
	"strings"
	"time"

	"github.com/leksa/datamapper-senyar/internal/model"
)

// MapSubmissionToFaskes converts an ODK submission to a Faskes model
func MapSubmissionToFaskes(submission map[string]interface{}) (*model.Faskes, error) {
	faskes := &model.Faskes{
		StatusFaskes: "operasional",
	}

	// Extract __id as ODK submission ID
	if id, ok := submission["__id"].(string); ok {
		faskes.ODKSubmissionID = &id
	}

	// Extract nama from calc_nama_faskes
	if nama, ok := submission["calc_nama_faskes"].(string); ok {
		faskes.Nama = nama
	}

	// Extract system metadata
	if system, ok := submission["__system"].(map[string]interface{}); ok {
		if submitterName, ok := system["submitterName"].(string); ok {
			faskes.SubmitterName = &submitterName
		}
		if submittedAt, ok := system["submissionDate"].(string); ok {
			if t, err := time.Parse(time.RFC3339, submittedAt); err == nil {
				faskes.SubmittedAt = &t
			}
		}
	}

	// Extract coordinates from calc_geometry ("lat lon" format)
	if geom, ok := submission["calc_geometry"].(string); ok {
		coords := strings.Fields(geom)
		if len(coords) >= 2 {
			if lat, err := strconv.ParseFloat(coords[0], 64); err == nil {
				faskes.Latitude = &lat
			}
			if lon, err := strconv.ParseFloat(coords[1], 64); err == nil {
				faskes.Longitude = &lon
			}
		}
	}

	// Build Alamat JSONB (codes and names)
	faskes.Alamat = model.JSONB{
		"id_provinsi":    getStringValue(submission, "sel_provinsi"),
		"id_kota_kab":    getStringValue(submission, "sel_kota_kab"),
		"id_kecamatan":   getStringValue(submission, "sel_kecamatan"),
		"id_desa":        getStringValue(submission, "sel_desa"),
		"nama_provinsi":  getStringValue(submission, "calc_nama_provinsi"),
		"nama_kota_kab":  getStringValue(submission, "calc_nama_kota_kab"),
		"nama_kecamatan": getStringValue(submission, "calc_nama_kecamatan"),
		"nama_desa":      getStringValue(submission, "calc_nama_desa"),
	}

	// Extract from grp_identitas
	if grpIdentitas, ok := submission["grp_identitas"].(map[string]interface{}); ok {
		// Jenis dan status faskes
		if jenis, ok := grpIdentitas["jenis_faskes"].(string); ok {
			faskes.JenisFaskes = jenis
		}
		if status, ok := grpIdentitas["status_faskes"].(string); ok {
			faskes.StatusFaskes = status
		}
		if kondisi, ok := grpIdentitas["kondisi_faskes"].(string); ok {
			faskes.KondisiFaskes = &kondisi
		}

		faskes.Identitas = model.JSONB{
			"nama_pj_faskes":    grpIdentitas["nama_pj_faskes"],
			"contact_pj_faskes": grpIdentitas["contact_pj_faskes"],
			"alamat_dusun":      grpIdentitas["alamat_dusun"],
		}

		// Extract coordinates from grp_identitas.koordinat if calc_geometry not available
		if faskes.Latitude == nil || faskes.Longitude == nil {
			if koordinat, ok := grpIdentitas["koordinat"].(map[string]interface{}); ok {
				if coords, ok := koordinat["coordinates"].([]interface{}); ok && len(coords) >= 2 {
					if lon, ok := coords[0].(float64); ok {
						faskes.Longitude = &lon
					}
					if lat, ok := coords[1].(float64); ok {
						faskes.Latitude = &lat
					}
				}
			}
		}
	}

	// Extract from grp_terisolir
	if grpTerisolir, ok := submission["grp_terisolir"].(map[string]interface{}); ok {
		faskes.Isolasi = model.JSONB{
			"terisolir": grpTerisolir["terisolir"],
			"akses_via": grpTerisolir["akses_via"],
		}
	}

	// Extract from grp_infra_operasional
	if grpInfra, ok := submission["grp_infra_operasional"].(map[string]interface{}); ok {
		faskes.Infrastruktur = model.JSONB{
			"sumber_listrik":        grpInfra["sumber_listrik"],
			"kondisi_penerangan":    grpInfra["kondisi_penerangan"],
			"ketersediaan_sinyal":   grpInfra["ketersediaan_sinyal"],
			"ketersediaan_internet": grpInfra["ketersediaan_internet"],
			"jaringan_orari":        grpInfra["jaringan_orari"],
			"ketersediaan_air":      grpInfra["ketersediaan_air"],
			"sumber_air":            grpInfra["sumber_air"],
			"kapasitas_sterilisasi": grpInfra["kapasitas_sterilisasi"],
		}
	}

	// Extract from grp_sumber_daya_manusia
	if grpSDM, ok := submission["grp_sumber_daya_manusia"].(map[string]interface{}); ok {
		faskes.SDM = model.JSONB{
			// Tenaga Kesehatan
			"dokter_umum":                    grpSDM["dokter_umum"],
			"dokter_gigi":                    grpSDM["dokter_gigi"],
			"psikolog":                       grpSDM["psikolog"],
			"perawat":                        grpSDM["perawat"],
			"bidan":                          grpSDM["bidan"],
			"apoteker":                       grpSDM["apoteker"],
			"tenaga_kefarmasian":             grpSDM["tenaga_kefarmasian"],
			"analis_kimia":                   grpSDM["analis_kimia"],
			"tenaga_kesehatan_masyarakat":    grpSDM["tenaga_kesehatan_masyarakat"],
			"tenaga_kesehatan_lingkungan":    grpSDM["tenaga_kesehatan_lingkungan"],
			"ahli_gizi":                      grpSDM["ahli_gizi"],
			// Non-Tenaga Kesehatan
			"tenaga_administrasi":            grpSDM["tenaga_administrasi"],
			"tenaga_keuangan":                grpSDM["tenaga_keuangan"],
			"tenaga_sistem_informasi_kesehatan": grpSDM["tenaga_sistem_informasi_kesehatan"],
			"perekam_medis":                  grpSDM["perekam_medis"],
			"petugas_keamanan_kebersihan":    grpSDM["petugas_keamanan_kebersihan"],
		}
	}

	// Extract from grp_sumber_daya
	if grpSumberDaya, ok := submission["grp_sumber_daya"].(map[string]interface{}); ok {
		faskes.Perbekalan = model.JSONB{
			// Perbekalan Kesehatan
			"obat_bahan_habis_pakai": grpSumberDaya["obat_bahan_habis_pakai"],
			"alat_kesehatan":         grpSumberDaya["alat_kesehatan"],
			"persalinan_kit":         grpSumberDaya["persalinan_kit"],
			// Bahan Sanitasi dan Sterilisasi
			"kaporit":        grpSumberDaya["kaporit"],
			"pac":            grpSumberDaya["pac"],
			"aquatab":        grpSumberDaya["aquatab"],
			"kantong_sampah": grpSumberDaya["kantong_sampah"],
			"repellent_lalat": grpSumberDaya["repellent_lalat"],
			"hygiene_kit":    grpSumberDaya["hygiene_kit"],
		}
	}

	// Extract from grp_penanggulangan
	if grpPenanggulangan, ok := submission["grp_penanggulangan"].(map[string]interface{}); ok {
		faskes.Klaster = model.JSONB{
			"klaster_yankes":           grpPenanggulangan["klaster_yankes"],
			"klaster_kendali_penyakit": grpPenanggulangan["klaster_kendali_penyakit"],
			"klaster_gizi":             grpPenanggulangan["klaster_gizi"],
			"klaster_jiwa":             grpPenanggulangan["klaster_jiwa"],
			"klaster_reproduksi":       grpPenanggulangan["klaster_reproduksi"],
			"klaster_dvi":              grpPenanggulangan["klaster_dvi"],
			"klaster_logkes":           grpPenanggulangan["klaster_logkes"],
		}
	}

	// Store raw submission data
	faskes.RawData = model.JSONB(submission)

	return faskes, nil
}

// ExtractFaskesPhotos extracts photo information from a faskes submission
func ExtractFaskesPhotos(submission map[string]interface{}) []PhotoInfo {
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
