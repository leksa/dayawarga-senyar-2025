-- Script to populate alamat nama fields from wilayah tables
-- and fix baseline_sumber from raw_data
-- Run this on production database after importing wilayah tables

-- 1. Update nama_provinsi from wilayah_provinsi
UPDATE locations l
SET alamat = alamat || jsonb_build_object('nama_provinsi', p.nama)
FROM wilayah_provinsi p
WHERE l.alamat->>'id_provinsi' = p.kode
AND l.deleted_at IS NULL
AND (l.alamat->>'nama_provinsi' IS NULL OR l.alamat->>'nama_provinsi' = '');

-- 2. Update nama_kota_kab from wilayah_kota_kab
UPDATE locations l
SET alamat = alamat || jsonb_build_object('nama_kota_kab', kk.nama)
FROM wilayah_kota_kab kk
WHERE l.alamat->>'id_kota_kab' = kk.kode
AND l.deleted_at IS NULL
AND (l.alamat->>'nama_kota_kab' IS NULL OR l.alamat->>'nama_kota_kab' = '');

-- 3. Update nama_kecamatan from wilayah_kecamatan
UPDATE locations l
SET alamat = alamat || jsonb_build_object('nama_kecamatan', k.nama)
FROM wilayah_kecamatan k
WHERE l.alamat->>'id_kecamatan' = k.kode
AND l.deleted_at IS NULL
AND (l.alamat->>'nama_kecamatan' IS NULL OR l.alamat->>'nama_kecamatan' = '');

-- 4. Update nama_desa from wilayah_desa
UPDATE locations l
SET alamat = alamat || jsonb_build_object('nama_desa', d.nama)
FROM wilayah_desa d
WHERE l.alamat->>'id_desa' = d.kode
AND l.deleted_at IS NULL
AND (l.alamat->>'nama_desa' IS NULL OR l.alamat->>'nama_desa' = '');

-- 5. Update baseline_sumber from grp_baseline (for BNPB dump data)
UPDATE locations l
SET identitas = identitas || jsonb_build_object('baseline_sumber', raw_data->'grp_baseline'->>'baseline_sumber')
WHERE deleted_at IS NULL
AND raw_data->'grp_baseline'->>'baseline_sumber' IS NOT NULL
AND raw_data->'grp_baseline'->>'baseline_sumber' <> ''
AND (identitas->>'baseline_sumber' IS NULL OR identitas->>'baseline_sumber' = '');

-- Check results
SELECT
    COUNT(*) as total,
    COUNT(CASE WHEN alamat->>'nama_provinsi' IS NOT NULL AND alamat->>'nama_provinsi' <> '' THEN 1 END) as has_nama_provinsi,
    COUNT(CASE WHEN alamat->>'nama_kota_kab' IS NOT NULL AND alamat->>'nama_kota_kab' <> '' THEN 1 END) as has_nama_kota_kab,
    COUNT(CASE WHEN alamat->>'nama_kecamatan' IS NOT NULL AND alamat->>'nama_kecamatan' <> '' THEN 1 END) as has_nama_kecamatan,
    COUNT(CASE WHEN alamat->>'nama_desa' IS NOT NULL AND alamat->>'nama_desa' <> '' THEN 1 END) as has_nama_desa,
    COUNT(CASE WHEN identitas->>'baseline_sumber' IS NOT NULL AND identitas->>'baseline_sumber' <> '' THEN 1 END) as has_baseline_sumber
FROM locations
WHERE deleted_at IS NULL;
