package odk

import "time"

// ODKConfig holds ODK Central connection settings
type ODKConfig struct {
	BaseURL   string
	Email     string
	Password  string
	ProjectID int
	FormID    string
}

// ODataResponse represents the OData response from ODK Central
type ODataResponse struct {
	Value         []Submission `json:"value"`
	ODataContext  string       `json:"@odata.context"`
	ODataNextLink string       `json:"@odata.nextLink,omitempty"`
}

// Submission represents a single ODK submission
type Submission struct {
	ID            string                 `json:"__id"`
	System        SystemInfo             `json:"__system"`
	Meta          Meta                   `json:"meta"`
	Mode          string                 `json:"mode"`
	SelProvinsi   string                 `json:"sel_provinsi"`
	SelKotaKab    string                 `json:"sel_kota_kab"`
	SelKecamatan  string                 `json:"sel_kecamatan"`
	SelDesa       string                 `json:"sel_desa"`
	SelPosko      *string                `json:"sel_posko"`
	NamaPosko     *string                `json:"nama_posko"`
	GrpIdentitas  GrpIdentitas           `json:"grp_identitas"`
	GrpTerisolir  GrpTerisolir           `json:"grp_terisolir"`
	GrpPengungsian GrpPengungsian        `json:"grp_pengungsian"`
	GrpDataPengungsi GrpDataPengungsi    `json:"grp_data_pengungsi"`
	GrpFasilitas  GrpFasilitas           `json:"grp_fasilitas"`
	GrpKomunikasi GrpKomunikasi          `json:"grp_komunikasi"`
	GrpAkses      GrpAkses               `json:"grp_akses"`
	GrpFoto       GrpFoto                `json:"grp_foto"`
	CalcNamaPosko string                 `json:"calc_nama_posko"`
	CalcIDDesa    string                 `json:"calc_id_desa"`
	CalcGeometry  string                 `json:"calc_geometry"`
	RawData       map[string]interface{} `json:"-"` // Will store full submission
}

// SystemInfo contains ODK system metadata
type SystemInfo struct {
	SubmissionDate      time.Time `json:"submissionDate"`
	UpdatedAt           time.Time `json:"updatedAt"`
	SubmitterID         string    `json:"submitterId"`
	SubmitterName       string    `json:"submitterName"`
	AttachmentsPresent  int       `json:"attachmentsPresent"`
	AttachmentsExpected int       `json:"attachmentsExpected"`
	Status              *string   `json:"status"`
	ReviewState         string    `json:"reviewState"`
	DeviceID            *string   `json:"deviceId"`
	Edits               int       `json:"edits"`
	FormVersion         string    `json:"formVersion"`
	DeletedAt           *string   `json:"deletedAt"`
}

// Meta contains instance metadata
type Meta struct {
	InstanceID   string     `json:"instanceID"`
	InstanceName string     `json:"instanceName"`
	Entity       EntityMeta `json:"entity"`
}

// EntityMeta contains entity label
type EntityMeta struct {
	Label string `json:"label"`
}

// GrpIdentitas contains identity information
type GrpIdentitas struct {
	NamaPenanggungjawab    string     `json:"nama_penanggungjawab"`
	ContactPenanggungjawab string     `json:"contact_penanggungjawab"`
	NamaRelawan            string     `json:"nama_relawan"`
	ContactRelawan         string     `json:"contact_relawan"`
	AlamatDusun            *string    `json:"alamat_dusun"`
	Koordinat              *GeoPoint  `json:"koordinat"`
}

// GeoPoint represents a GeoJSON point from ODK
type GeoPoint struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"` // [longitude, latitude]
}

// GrpTerisolir contains isolation status
type GrpTerisolir struct {
	Terisolir string  `json:"terisolir"`
	AksesVia  *string `json:"akses_via"`
}

// GrpPengungsian contains displacement information
type GrpPengungsian struct {
	JenisPengungsian  string `json:"jenis_pengungsian"`
	DetailPengungsian string `json:"detail_pengungsian"`
	TotalPengungsi    int    `json:"total_pengungsi"`
}

// GrpDataPengungsi contains refugee demographic data
type GrpDataPengungsi struct {
	JumlahKK         int  `json:"jumlah_kk"`
	DewasaPerempuan  *int `json:"dewasa_perempuan"`
	DewasaLaki       *int `json:"dewasa_laki"`
	RemajaPerempuan  *int `json:"remaja_perempuan"`
	RemajaLaki       *int `json:"remaja_laki"`
	AnakPerempuan    *int `json:"anak_perempuan"`
	AnakLaki         *int `json:"anak_laki"`
	BalitaPerempuan  *int `json:"balita_perempuan"`
	BalitaLaki       *int `json:"balita_laki"`
	BayiPerempuan    *int `json:"bayi_perempuan"`
	BayiLaki         *int `json:"bayi_laki"`
	Lansia           *int `json:"lansia"`
	IbuMenyusui      *int `json:"ibu_menyusui"`
	IbuHamil         *int `json:"ibu_hamil"`
	RemajaTanpaOrtu  *int `json:"remaja_tanpa_ortu"`
	AnakTanpaOrtu    *int `json:"anak_tanpa_ortu"`
	BayiTanpaIbu     *int `json:"bayi_tanpa_ibu"`
	Difabel          *int `json:"difabel"`
	Komorbid         *int `json:"komorbid"`
}

// GrpFasilitas contains facility information
type GrpFasilitas struct {
	PoskoLogistik      string  `json:"posko_logistik"`
	PoskoFaskes        string  `json:"posko_faskes"`
	DapurUmum          string  `json:"dapur_umum"`
	KapasitasDapur     *int    `json:"kapasitas_dapur"`
	KetersediaanAir    string  `json:"ketersediaan_air"`
	SumberAir          *string `json:"sumber_air"`
	ToiletPerempuan    *int    `json:"toilet_perempuan"`
	ToiletLaki         *int    `json:"toilet_laki"`
	ToiletCampur       *int    `json:"toilet_campur"`
	TempatSampah       *int    `json:"tempat_sampah"`
	SumberListrik      string  `json:"sumber_listrik"`
	KondisiPenerangan  string  `json:"kondisi_penerangan"`
	TitikAksesListrik  *int    `json:"titik_akses_listrik"`
	PoskoKesehatan     string  `json:"posko_kesehatan"`
	PoskoObat          string  `json:"posko_obat"`
	PoskoPsikososial   string  `json:"posko_psikososial"`
	RuangLaktasi       string  `json:"ruang_laktasi"`
	LayananLansia      string  `json:"layanan_lansia"`
	LayananKeluarga    string  `json:"layanan_keluarga"`
	SekolahDarurat     string  `json:"sekolah_darurat"`
	ProgramPengganti   *string `json:"program_pengganti"`
	PetugasKeamanan    *string `json:"petugas_keamanan"`
	AreaInteraksi      *string `json:"area_interaksi"`
	AreaBermain        *string `json:"area_bermain"`
}

// GrpKomunikasi contains communication information
type GrpKomunikasi struct {
	KetersediaanSinyal   string  `json:"ketersediaan_sinyal"`
	JaringanOrari        string  `json:"jaringan_orari"`
	KetersediaanInternet *string `json:"ketersediaan_internet"`
}

// GrpAkses contains access information
type GrpAkses struct {
	JarakPKM           *float64 `json:"jarak_pkm"`
	JarakPoskoLogistik *float64 `json:"jarak_posko_logistik"`
}

// GrpFoto contains photo filenames
type GrpFoto struct {
	FotoDepan         *string `json:"foto_depan"`
	FotoArea1         *string `json:"foto_area1"`
	FotoArea2         *string `json:"foto_area2"`
	FotoArea3         *string `json:"foto_area3"`
	FotoToilet        *string `json:"foto_toilet"`
	FotoSampah        *string `json:"foto_sampah"`
	FotoFaskes        *string `json:"foto_faskes"`
	FotoDapur         *string `json:"foto_dapur"`
}

// SyncState represents synchronization state tracking
type SyncState struct {
	ID              string     `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	FormID          string     `json:"form_id" gorm:"uniqueIndex"`
	LastSyncTime    *time.Time `json:"last_sync_time" gorm:"column:last_sync_timestamp"`
	LastETag        *string    `json:"last_etag" gorm:"column:last_etag"`
	LastRecordCount int        `json:"last_record_count"`
	TotalRecords    int        `json:"total_records"`
	Status          string     `json:"status"` // idle, syncing, error
	ErrorMessage    *string    `json:"error_message"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (SyncState) TableName() string {
	return "sync_state"
}
