<script setup lang="ts">
import ShareImageTemplate from '@/components/ShareImageTemplate.vue'

// Sample data untuk preview
const sampleData = {
  category: 'kebutuhan',
  photo: 'https://images.unsplash.com/photo-1469571486292-0ba58a3f068b?w=600',
  kabupaten: 'Aceh Tamiang',
  kecamatan: 'Karang Baru',
  desa: 'Suka Mulia',
  submitter: 'relawan_aceh - PMI Aceh',
  content: 'Butuh bantuan air bersih untuk 150 KK di Desa Suka Mulia. Stok air bersih habis sejak kemarin. Warga kesulitan mendapatkan air untuk minum dan memasak. Mohon bantuan segera dari relawan atau donatur yang bisa membantu. Koordinasi bisa hubungi nomor di bawah. Terima kasih atas perhatiannya.',
  tags: ['air_bersih', 'sembako', 'medis'],
  timestamp: '19-01-2026 14:30'
}

const sampleDataNoPhoto = {
  category: 'informasi',
  kabupaten: 'Pidie Jaya',
  kecamatan: 'Bandar Baru',
  desa: 'Blang Asan',
  submitter: 'admin_dayawarga',
  content: 'Update situasi: Jalan akses menuju Desa Blang Asan sudah bisa dilalui kendaraan roda 4. Evakuasi warga sudah selesai dilakukan. Posko pengungsian di Balai Desa masih aktif melayani 75 KK.',
  tags: ['transport_roda4'],
  timestamp: '19-01-2026 10:15'
}

const sampleDataFollowUp = {
  category: 'follow-up',
  photo: 'https://images.unsplash.com/photo-1532629345422-7515f3d16bb6?w=600',
  kabupaten: 'Aceh Utara',
  kecamatan: 'Lhoksukon',
  desa: 'Tanjung Baroh',
  submitter: 'koordinator_lapangan',
  content: 'Bantuan sembako dari donatur sudah diterima dan sedang didistribusikan ke 200 KK terdampak. Proses distribusi berjalan lancar dengan koordinasi RT/RW setempat.',
  tags: ['sembako', 'dapur_umum'],
  timestamp: '18-01-2026 16:45'
}

const sampleDataTerimaBantuan = {
  category: 'info_bantuan',
  photo: 'https://images.unsplash.com/photo-1593113598332-cd288d649433?w=600',
  kabupaten: 'Bireuen',
  kecamatan: 'Samalanga',
  desa: 'Meunasah Dayah',
  submitter: 'posko_bireuen',
  content: 'Alhamdulillah telah menerima bantuan berupa 500 liter air bersih, 100 paket sembako, dan 50 selimut dari para donatur. Terima kasih atas bantuan dan kepeduliannya.',
  tags: ['air_bersih', 'sembako'],
  timestamp: '18-01-2026 09:20'
}
</script>

<template>
  <div class="preview-page">
    <h1 class="preview-title">Preview Template Share Image</h1>
    <p class="preview-subtitle">Template gambar yang akan di-download saat user klik tombol download</p>

    <div class="preview-grid">
      <!-- Sample 1: Butuh Bantuan dengan foto -->
      <div class="preview-item">
        <h3>Butuh Bantuan (dengan foto)</h3>
        <ShareImageTemplate v-bind="sampleData" />
      </div>

      <!-- Sample 2: Informasi tanpa foto -->
      <div class="preview-item">
        <h3>Informasi (tanpa foto)</h3>
        <ShareImageTemplate v-bind="sampleDataNoPhoto" />
      </div>

      <!-- Sample 3: Follow-up -->
      <div class="preview-item">
        <h3>Follow-up</h3>
        <ShareImageTemplate v-bind="sampleDataFollowUp" />
      </div>

      <!-- Sample 4: Terima Bantuan -->
      <div class="preview-item">
        <h3>Terima Bantuan</h3>
        <ShareImageTemplate v-bind="sampleDataTerimaBantuan" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.preview-page {
  min-height: 100vh;
  background: #f3f4f6;
  padding: 40px 20px;
}

.preview-title {
  text-align: center;
  font-size: 28px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 8px;
}

.preview-subtitle {
  text-align: center;
  font-size: 14px;
  color: #6b7280;
  margin-bottom: 40px;
}

.preview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(480px, 1fr));
  gap: 40px;
  max-width: 1200px;
  margin: 0 auto;
}

.preview-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.preview-item h3 {
  font-size: 16px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 16px;
}
</style>
