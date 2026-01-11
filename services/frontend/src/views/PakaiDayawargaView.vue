<script setup lang="ts">
import DataLayersSidebar from '@/components/DataLayersSidebar.vue'
import { ClipboardList, MessageSquare, Wifi, Users, Heart } from 'lucide-vue-next'

// Image URLs - stored in public/images/
const odkCollectImage = '/images/odk-collect-screenshots.png'
const odkCentralImage = '/images/odk-central-dashboard.png'
</script>

<template>
  <div class="flex-1 flex overflow-hidden">
    <DataLayersSidebar />

    <!-- Pakai Dayawarga Content -->
    <main class="flex-1 bg-gradient-to-b from-gray-50 to-white overflow-y-auto">
      <div class="max-w-3xl mx-auto px-8 py-12">
        <!-- Header -->
        <div class="mb-12">
          <h1 class="text-4xl font-bold text-gray-900 mb-4">Apa yang Bisa Dilakukan Relawan & Komunitas di Dayawarga</h1>
          <p class="text-xl text-gray-600 leading-relaxed">
            <strong>Aplikasi Relawan Dayawarga</strong> dirancang sebagai alat kerja praktis yang bisa langsung digunakan di lapangan maupun saat memantau kondisi dari jarak jauh. Melalui aplikasi ini, relawan dan komunitas dapat berkontribusi secara nyata dalam siklus data kebencanaan—dari pengumpulan hingga pembaruan informasi.
          </p>
        </div>

        <!-- Hal-hal yang dapat dilakukan -->
        <section class="mb-12">
          <h2 class="text-2xl font-bold text-gray-900 mb-6">Hal-hal yang dapat dilakukan antara lain:</h2>

          <!-- Input & Verifikasi Data -->
          <div class="bg-white border border-gray-200 rounded-xl p-6 mb-6 shadow-sm">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center flex-shrink-0">
                <ClipboardList class="w-6 h-6 text-blue-600" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Input, verifikasi, dan revisi data fasilitas di lapangan</h3>
                <p class="text-gray-600 mb-3">
                  Relawan dapat mengisi, memeriksa, dan memperbarui data berbagai fasilitas yang ada di wilayah terdampak—seperti <strong>fasilitas umum, posko, layanan kesehatan, akses logistik</strong>, dan lainnya—menggunakan perangkat Android maupun Web.
                </p>
                <p class="text-gray-500 text-sm">
                  Formulir standar untuk pendataan fasilitas sudah tersedia di aplikasi, sehingga proses input data menjadi lebih seragam dan mudah dibandingkan pencatatan manual.
                </p>
              </div>
            </div>
          </div>

          <!-- Feed Update -->
          <div class="bg-white border border-gray-200 rounded-xl p-6 mb-6 shadow-sm">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center flex-shrink-0">
                <MessageSquare class="w-6 h-6 text-orange-600" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Menyampaikan pembaruan kondisi lapangan melalui "Feed Update"</h3>
                <p class="text-gray-600 mb-3">
                  Setiap relawan atau komunitas dapat mengirimkan informasi terbaru tentang apa pun yang terjadi di lapangan—<strong>perubahan kondisi, kebutuhan mendesak, hambatan akses, atau perkembangan pemulihan</strong>—melalui formulir Feed Update.
                </p>
                <p class="text-gray-500 text-sm">
                  Informasi ini akan membantu relawan lain, komunitas pendukung, dan pihak yang memantau Peta Bencana Dayawarga untuk mendapatkan gambaran situasi yang lebih aktual.
                </p>
              </div>
            </div>
          </div>

          <!-- Offline Mode -->
          <div class="bg-white border border-gray-200 rounded-xl p-6 mb-6 shadow-sm">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center flex-shrink-0">
                <Wifi class="w-6 h-6 text-green-600" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Digunakan di wilayah dengan keterbatasan sinyal</h3>
                <p class="text-gray-600">
                  Aplikasi mobile Dayawarga dapat dibawa ke lapangan yang memiliki akses internet terbatas atau tidak stabil. Data dapat dikumpulkan terlebih dahulu dan dikirimkan ketika koneksi tersedia, sehingga <strong>kerja lapangan tetap berjalan tanpa bergantung pada sinyal setiap saat</strong>.
                </p>
              </div>
            </div>
          </div>

          <!-- Image: ODK Collect Screenshots -->
          <div class="mb-6">
            <img
              :src="odkCollectImage"
              alt="ODK Collect - Aplikasi pengumpulan data di lapangan"
              class="w-full rounded-xl border border-gray-200 shadow-sm"
              @error="($event.target as HTMLImageElement).style.display = 'none'"
            />
            <p class="text-sm text-gray-500 text-center mt-3">
              ODK Collect: Aplikasi Android untuk input data dan koordinat di lapangan
            </p>
          </div>

          <!-- Pengelolaan Tim -->
          <div class="bg-white border border-gray-200 rounded-xl p-6 mb-6 shadow-sm">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center flex-shrink-0">
                <Users class="w-6 h-6 text-purple-600" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Pengelolaan tim lapangan <span class="text-sm font-normal text-amber-600 bg-amber-50 px-2 py-0.5 rounded-full ml-2">segera hadir</span></h3>
                <p class="text-gray-600">
                  Ke depan, Dayawarga akan menyiapkan halaman manajemen tim lapangan, agar komunitas dan lembaga dapat mengelola anggota relawannya secara lebih mandiri dan terstruktur. Fitur ini masih dalam tahap pengembangan dan akan diumumkan secara terbuka ketika siap digunakan.
                </p>
              </div>
            </div>
          </div>

          <!-- Image: ODK Central Dashboard -->
          <div class="mb-6">
            <img
              :src="odkCentralImage"
              alt="ODK Central - Dashboard manajemen data"
              class="w-full rounded-xl border border-gray-200 shadow-sm"
              @error="($event.target as HTMLImageElement).style.display = 'none'"
            />
            <p class="text-sm text-gray-500 text-center mt-3">
              ODK Central: Dashboard manajemen formulir dan data submissions
            </p>
          </div>
        </section>

        <!-- Perangkat Kerja Bersama -->
        <div class="bg-blue-50 border-l-4 border-blue-500 p-6 rounded-r-lg mb-12">
          <p class="text-gray-700 leading-relaxed mb-4">
            <strong>Aplikasi Relawan Dayawarga</strong> adalah perangkat kerja bersama yang kami siapkan untuk mendukung pengumpulan, pengolahan, dan penyajian data lapangan secara cepat, rapi, dan dapat dipertanggungjawabkan—terutama dalam konteks kebencanaan.
          </p>
          <p class="text-gray-700 leading-relaxed">
            Aplikasi ini bukan sekadar alat teknis, melainkan bagian dari <strong>ekosistem kolaborasi</strong> antara warga, relawan, komunitas, NGO, dan lembaga publik agar semua pihak dapat bekerja di atas data yang sama dan saling terhubung.
          </p>
        </div>

        <!-- Bergabung sebagai Relawan -->
        <section class="mb-12">
          <h2 class="text-2xl font-bold text-gray-900 mb-6">Bergabung sebagai Relawan</h2>
          <p class="text-gray-600 mb-6">
            Saat ini, Dayawarga belum menyediakan halaman manajemen mandiri bagi komunitas atau lembaga untuk mengelola tim relawannya masing-masing.
          </p>
          <p class="text-gray-600 mb-6">
            Namun, jika kamu tertarik untuk menjadi bagian dari kerja bersama ini—baik sebagai <strong>relawan lapangan, pengolah data, pemeta, penulis cerita</strong>, maupun <strong>pendukung teknis</strong>—kamu dapat:
          </p>

          <div class="space-y-3 mb-8">
            <div class="flex items-center gap-3 bg-white border border-gray-200 rounded-lg p-4">
              <span class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center text-blue-600 font-semibold text-sm">1</span>
              <span class="text-gray-700">Mengisi formulir pendaftaran yang tersedia di bawah halaman ini, atau</span>
            </div>
            <div class="flex items-center gap-3 bg-white border border-gray-200 rounded-lg p-4">
              <span class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center text-blue-600 font-semibold text-sm">2</span>
              <span class="text-gray-700">Menghubungi kami secara langsung melalui kanal kontak Dayawarga</span>
            </div>
          </div>

          <p class="text-gray-600">
            Kami percaya kerja besar menghadapi bencana hanya mungkin dilakukan jika dibangun bersama, dengan peran yang beragam dan saling melengkapi.
          </p>
        </section>

        <!-- CTA -->
        <section class="bg-gradient-to-r from-blue-600 to-blue-700 rounded-2xl p-8 text-center text-white">
          <Heart class="w-10 h-10 mx-auto mb-4 opacity-80" />
          <p class="text-lg mb-2">
            Aplikasi Relawan Dayawarga adalah alatnya—
          </p>
          <p class="text-2xl font-bold">
            kolaborasi manusialah yang menjadi kekuatannya.
          </p>
        </section>

        <!-- Footer spacing -->
        <div class="h-12"></div>
      </div>
    </main>
  </div>
</template>
