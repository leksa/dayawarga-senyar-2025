-- ===========================================
-- SAMPLE DATA - Lokasi Pengungsian Siklon Senyar
-- ===========================================

-- Lokasi di area Sumatera Barat (simulasi)
INSERT INTO locations (nama, type, status, geom, geo_meta, identitas, alamat, data_pengungsi, fasilitas, komunikasi)
VALUES
-- Posko 1: Central Medical Hub
(
    'Central Medical Hub',
    'kesehatan',
    'operational',
    ST_SetSRID(ST_MakePoint(100.3543, -0.9471), 4326),
    '{"altitude": 15.5, "accuracy": 3.2}'::jsonb,
    '{
        "nama_pj": "Dr. Ahmad Rizki",
        "contact_pj": "081234567890",
        "nama_relawan": "Tim Medis PMI",
        "contact_relawan": "081234567891",
        "lembaga": "PMI Sumbar"
    }'::jsonb,
    '{
        "provinsi": "Sumatera Barat",
        "kabupaten": "Padang",
        "kecamatan": "Padang Barat",
        "desa": "Olo",
        "dusun": "Kampung Jao"
    }'::jsonb,
    '{
        "jumlah_kk": 45,
        "total_jiwa": 180,
        "dewasa_laki": 40,
        "dewasa_perempuan": 45,
        "remaja_laki": 15,
        "remaja_perempuan": 20,
        "anak_laki": 25,
        "anak_perempuan": 20,
        "balita": 10,
        "lansia": 5,
        "ibu_hamil": 3,
        "ibu_menyusui": 5,
        "difabel": 2
    }'::jsonb,
    '{
        "dapur_umum": true,
        "kapasitas_dapur": 200,
        "air_bersih": {"ketersediaan": "cukup", "sumber": ["pdam", "sumur_bor"]},
        "toilet_tersedia": 8,
        "layanan_kesehatan": true,
        "posko_psikososial": true
    }'::jsonb,
    '{"sinyal": ["4g", "wifi"], "orari": true}'::jsonb
),

-- Posko 2: Shelter Beta-2
(
    'Shelter Beta-2',
    'posko',
    'operational',
    ST_SetSRID(ST_MakePoint(100.3621, -0.9512), 4326),
    '{"altitude": 12.3, "accuracy": 2.8}'::jsonb,
    '{
        "nama_pj": "Budi Santoso",
        "contact_pj": "082345678901",
        "nama_relawan": "Relawan BNPB",
        "contact_relawan": "082345678902",
        "lembaga": "BNPB"
    }'::jsonb,
    '{
        "provinsi": "Sumatera Barat",
        "kabupaten": "Padang",
        "kecamatan": "Padang Timur",
        "desa": "Sawahan",
        "dusun": "RT 05"
    }'::jsonb,
    '{
        "jumlah_kk": 78,
        "total_jiwa": 312,
        "dewasa_laki": 70,
        "dewasa_perempuan": 85,
        "remaja_laki": 30,
        "remaja_perempuan": 35,
        "anak_laki": 40,
        "anak_perempuan": 35,
        "balita": 12,
        "lansia": 5,
        "ibu_hamil": 4,
        "ibu_menyusui": 8,
        "difabel": 3
    }'::jsonb,
    '{
        "dapur_umum": true,
        "kapasitas_dapur": 350,
        "air_bersih": {"ketersediaan": "cukup", "sumber": ["pdam"]},
        "toilet_tersedia": 12,
        "layanan_kesehatan": false,
        "posko_psikososial": true
    }'::jsonb,
    '{"sinyal": ["4g"], "orari": false}'::jsonb
),

-- Posko 3: Water Point Delta
(
    'Water Point Delta',
    'air_bersih',
    'operational',
    ST_SetSRID(ST_MakePoint(100.3489, -0.9589), 4326),
    '{"altitude": 8.7, "accuracy": 4.1}'::jsonb,
    '{
        "nama_pj": "Hendra Wijaya",
        "contact_pj": "083456789012",
        "nama_relawan": "Tim Air Bersih",
        "contact_relawan": "083456789013",
        "lembaga": "Relawan Independen"
    }'::jsonb,
    '{
        "provinsi": "Sumatera Barat",
        "kabupaten": "Padang",
        "kecamatan": "Padang Selatan",
        "desa": "Seberang Padang",
        "dusun": "RT 08"
    }'::jsonb,
    '{
        "jumlah_kk": 0,
        "total_jiwa": 0
    }'::jsonb,
    '{
        "dapur_umum": false,
        "air_bersih": {"ketersediaan": "baik", "sumber": ["sumur_artesis"], "kapasitas_liter": 10000},
        "toilet_tersedia": 0,
        "layanan_kesehatan": false,
        "catatan": "Titik distribusi air bersih untuk 5 posko sekitar"
    }'::jsonb,
    '{"sinyal": ["4g"], "orari": false}'::jsonb
),

-- Posko 4: Shelter Alpha
(
    'Shelter Alpha',
    'posko',
    'operational',
    ST_SetSRID(ST_MakePoint(100.3712, -0.9438), 4326),
    '{"altitude": 18.2, "accuracy": 2.5}'::jsonb,
    '{
        "nama_pj": "Siti Rahmawati",
        "contact_pj": "084567890123",
        "nama_relawan": "Komunitas Peduli",
        "contact_relawan": "084567890124",
        "lembaga": "Mercy Corps"
    }'::jsonb,
    '{
        "provinsi": "Sumatera Barat",
        "kabupaten": "Padang",
        "kecamatan": "Kuranji",
        "desa": "Pasar Ambacang",
        "dusun": "Blok A"
    }'::jsonb,
    '{
        "jumlah_kk": 120,
        "total_jiwa": 480,
        "dewasa_laki": 100,
        "dewasa_perempuan": 120,
        "remaja_laki": 50,
        "remaja_perempuan": 55,
        "anak_laki": 60,
        "anak_perempuan": 55,
        "balita": 25,
        "lansia": 15,
        "ibu_hamil": 8,
        "ibu_menyusui": 12,
        "difabel": 5
    }'::jsonb,
    '{
        "dapur_umum": true,
        "kapasitas_dapur": 500,
        "air_bersih": {"ketersediaan": "kurang", "sumber": ["tangki_distribusi"]},
        "toilet_tersedia": 15,
        "layanan_kesehatan": true,
        "posko_psikososial": true,
        "area_bermain_anak": true
    }'::jsonb,
    '{"sinyal": ["4g", "wifi"], "orari": true}'::jsonb
),

-- Posko 5: Medical Post North
(
    'Medical Post North',
    'kesehatan',
    'limited',
    ST_SetSRID(ST_MakePoint(100.3398, -0.9325), 4326),
    '{"altitude": 22.1, "accuracy": 3.5}'::jsonb,
    '{
        "nama_pj": "Dr. Maya Putri",
        "contact_pj": "085678901234",
        "nama_relawan": "Tim Medis Basarnas",
        "contact_relawan": "085678901235",
        "lembaga": "Basarnas"
    }'::jsonb,
    '{
        "provinsi": "Sumatera Barat",
        "kabupaten": "Padang",
        "kecamatan": "Nanggalo",
        "desa": "Surau Gadang",
        "dusun": "RT 03"
    }'::jsonb,
    '{
        "jumlah_kk": 25,
        "total_jiwa": 100,
        "dewasa_laki": 20,
        "dewasa_perempuan": 25,
        "remaja_laki": 10,
        "remaja_perempuan": 12,
        "anak_laki": 15,
        "anak_perempuan": 10,
        "balita": 5,
        "lansia": 3,
        "ibu_hamil": 2,
        "ibu_menyusui": 3,
        "difabel": 1
    }'::jsonb,
    '{
        "dapur_umum": false,
        "air_bersih": {"ketersediaan": "kurang", "sumber": ["tangki_distribusi"]},
        "toilet_tersedia": 4,
        "layanan_kesehatan": true,
        "catatan": "Kekurangan obat-obatan"
    }'::jsonb,
    '{"sinyal": ["3g"], "orari": false}'::jsonb
),

-- Posko 6: Emergency Shelter C
(
    'Emergency Shelter C',
    'posko',
    'operational',
    ST_SetSRID(ST_MakePoint(100.3256, -0.9678), 4326),
    '{"altitude": 5.3, "accuracy": 4.2}'::jsonb,
    '{
        "nama_pj": "Andi Pratama",
        "contact_pj": "086789012345",
        "nama_relawan": "Satgas Bencana",
        "contact_relawan": "086789012346",
        "lembaga": "BPBD Sumbar"
    }'::jsonb,
    '{
        "provinsi": "Sumatera Barat",
        "kabupaten": "Padang",
        "kecamatan": "Bungus Teluk Kabung",
        "desa": "Bungus Barat",
        "dusun": "Pantai"
    }'::jsonb,
    '{
        "jumlah_kk": 65,
        "total_jiwa": 260,
        "dewasa_laki": 55,
        "dewasa_perempuan": 70,
        "remaja_laki": 25,
        "remaja_perempuan": 30,
        "anak_laki": 35,
        "anak_perempuan": 30,
        "balita": 10,
        "lansia": 5,
        "ibu_hamil": 3,
        "ibu_menyusui": 6,
        "difabel": 2
    }'::jsonb,
    '{
        "dapur_umum": true,
        "kapasitas_dapur": 280,
        "air_bersih": {"ketersediaan": "cukup", "sumber": ["sumur_dangkal", "tangki"]},
        "toilet_tersedia": 10,
        "layanan_kesehatan": false,
        "posko_psikososial": false
    }'::jsonb,
    '{"sinyal": ["4g"], "orari": false}'::jsonb
);

-- ===========================================
-- SAMPLE FEEDS
-- ===========================================
INSERT INTO information_feeds (location_id, content, category, type, username, organization, submitted_at, geom)
SELECT
    l.id,
    'Dibutuhkan obat-obatan standar untuk 200 pengungsi. Stok antibiotik dan analgesik menipis.',
    'kebutuhan',
    'kesehatan',
    'leksa',
    'PMI Sumbar',
    NOW() - INTERVAL '2 hours',
    l.geom
FROM locations l WHERE l.nama = 'Central Medical Hub';

INSERT INTO information_feeds (location_id, content, category, type, username, organization, submitted_at, geom)
SELECT
    l.id,
    'Truk logistik telah tiba. Proses bongkar muat sedang berlangsung. Bantuan makanan siap didistribusikan besok pagi.',
    'informasi',
    'logistik',
    'sarah_coord',
    'BNPB',
    NOW() - INTERVAL '3 hours',
    l.geom
FROM locations l WHERE l.nama = 'Shelter Beta-2';

INSERT INTO information_feeds (location_id, content, category, type, username, organization, submitted_at, geom)
SELECT
    l.id,
    'Pompa air mengalami kerusakan minor. Teknisi sedang dalam perjalanan untuk perbaikan.',
    'informasi',
    'air bersih',
    'field_ops_1',
    'Relawan Independen',
    NOW() - INTERVAL '4 hours',
    l.geom
FROM locations l WHERE l.nama = 'Water Point Delta';

INSERT INTO information_feeds (location_id, content, category, type, username, organization, submitted_at, geom)
SELECT
    l.id,
    'Evakuasi warga dari daerah terdampak selesai. Total 150 jiwa berhasil dievakuasi.',
    'informasi',
    'SAR',
    'sar_sumbar',
    'Basarnas',
    NOW() - INTERVAL '5 hours',
    l.geom
FROM locations l WHERE l.nama = 'Shelter Beta-2';

INSERT INTO information_feeds (location_id, content, category, type, username, organization, submitted_at, geom)
SELECT
    l.id,
    'Tim medis tambahan dari Jakarta telah tiba. Total 5 dokter dan 10 perawat bergabung.',
    'informasi',
    'kesehatan',
    'medic_team',
    'Mercy Corps',
    NOW() - INTERVAL '6 hours',
    l.geom
FROM locations l WHERE l.nama = 'Central Medical Hub';

INSERT INTO information_feeds (content, category, type, username, organization, submitted_at, geom)
VALUES (
    'Peringatan dini cuaca ekstrem dikeluarkan untuk 24 jam ke depan. Harap semua unit siaga.',
    'informasi',
    'komunikasi',
    'admin_pusat',
    'BPBD Sumbar',
    NOW() - INTERVAL '8 hours',
    ST_SetSRID(ST_MakePoint(100.3500, -0.9500), 4326)
);

-- ===========================================
-- SUCCESS MESSAGE
-- ===========================================
DO $$
BEGIN
    RAISE NOTICE 'Sample data inserted: 6 locations, 6 feeds';
END $$;
