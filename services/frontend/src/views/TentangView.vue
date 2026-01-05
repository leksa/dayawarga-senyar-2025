<script setup lang="ts">
import DataLayersSidebar from '@/components/DataLayersSidebar.vue'
import { MapPin, BookOpen, Users, Heart, Instagram, Twitter, Code } from 'lucide-vue-next'
</script>

<template>
  <div class="flex-1 flex overflow-hidden">
    <DataLayersSidebar />

    <!-- Tentang Content -->
    <main class="flex-1 bg-gradient-to-b from-gray-50 to-white overflow-y-auto">
      <div class="max-w-3xl mx-auto px-8 py-12">
        <!-- Header -->
        <div class="mb-12">
          <h1 class="text-4xl font-bold text-gray-900 mb-4">Tentang Dayawarga</h1>
          <p class="text-xl text-gray-600 leading-relaxed">
            <strong>Dayawarga</strong> adalah ruang bersama untuk melihat, memahami, dan merespons bencana secara lebih terbuka, kolaboratif, dan berbasis data. Kami hadir untuk membantu semua pihak—warga, relawan, komunitas, NGO, hingga lembaga pemerintah—bergerak lebih cepat dan tepat dalam proses mitigasi, tanggap darurat, dan pemulihan pascabencana.
          </p>
        </div>

        <!-- Bencana Siklon -->
        <div class="bg-blue-50 border-l-4 border-blue-500 p-6 rounded-r-lg mb-12">
          <p class="text-gray-700 leading-relaxed">
            <strong>Bencana Siklon Senyar Sumatra 2025</strong> mengajarkan kita semua sebuah peristiwa yang meninggalkan dampak luas bagi kehidupan warga dan lingkungan. Dari peristiwa inilah kami belajar bahwa data yang tersebar, informasi yang terputus, dan cerita warga yang tak terdengar sering kali memperlambat pemulihan. Dayawarga lahir untuk menjembatani celah itu.
          </p>
        </div>

        <!-- Apa yang Kami Tampilkan -->
        <section class="mb-12">
          <h2 class="text-2xl font-bold text-gray-900 mb-6">Apa yang Kami Tampilkan</h2>

          <!-- Peta Pemulihan -->
          <div class="bg-white border border-gray-200 rounded-xl p-6 mb-6 shadow-sm">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center flex-shrink-0">
                <MapPin class="w-6 h-6 text-green-600" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Peta Pemulihan Berbasis Data</h3>
                <p class="text-gray-600 mb-4">
                  Dayawarga menampilkan peta pemulihan bencana dalam format <strong>spatial GIS</strong>, dilengkapi dengan berbagai layer data penting—mulai dari kondisi wilayah terdampak, infrastruktur, akses logistik, hingga kebutuhan lapangan.
                </p>
                <div class="grid grid-cols-2 gap-3">
                  <div class="flex items-center gap-2 text-sm text-gray-600">
                    <span class="w-2 h-2 bg-green-500 rounded-full"></span>
                    Lebih cepat
                  </div>
                  <div class="flex items-center gap-2 text-sm text-gray-600">
                    <span class="w-2 h-2 bg-green-500 rounded-full"></span>
                    Lebih tanggap
                  </div>
                  <div class="flex items-center gap-2 text-sm text-gray-600">
                    <span class="w-2 h-2 bg-green-500 rounded-full"></span>
                    Lebih informatif
                  </div>
                  <div class="flex items-center gap-2 text-sm text-gray-600">
                    <span class="w-2 h-2 bg-green-500 rounded-full"></span>
                    Lebih terkoordinasi
                  </div>
                </div>
                <p class="text-gray-500 text-sm mt-4">
                  Pengumpulan dan pengolahan data dilakukan secara kolaboratif oleh <strong>relawan, komunitas, NGO, dan lembaga pemerintah</strong>, dengan prinsip keterbukaan dan verifikasi bersama.
                </p>
              </div>
            </div>
          </div>

          <!-- Cerita dari Lapangan -->
          <div class="bg-white border border-gray-200 rounded-xl p-6 mb-6 shadow-sm">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center flex-shrink-0">
                <BookOpen class="w-6 h-6 text-purple-600" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Cerita dari Lapangan</h3>
                <p class="text-gray-600">
                  Bencana bukan hanya soal statistik dan peta. Di Dayawarga, kami juga menghadirkan sisi lain dari bencana melalui <strong>cerita, esai, dan catatan lapangan</strong> yang ditulis oleh warga dan relawan. Cerita-cerita ini menjadi ruang refleksi, dokumentasi sosial, sekaligus pengingat bahwa pemulihan adalah proses manusiawi, bukan sekadar teknis.
                </p>
              </div>
            </div>
          </div>

          <!-- Sumber Kode dan Data Terbuka -->
          <div class="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-cyan-100 rounded-xl flex items-center justify-center flex-shrink-0">
                <Code class="w-6 h-6 text-cyan-600" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Sumber Kode dan Data Terbuka</h3>
                <p class="text-gray-600 mb-3">
                  Aplikasi Peta Dayawarga dibangun dengan semangat <strong>open source</strong>. Kami memanfaatkan berbagai teknologi terbuka yang tersedia, dengan backbone pengumpulan data menggunakan <strong>OpenDataKit</strong>. Sementara data GIS processing dan visualisasi dikembangkan secara mandiri dengan bantuan agent AI.
                </p>
                <p class="text-gray-600 mb-3">
                  Dengan prinsip kolaborasi dan keterbukaan, seluruh <strong>sumber kode dan data olahan spatial</strong> kami rilis di repository GitHub agar dapat digunakan dan dimanfaatkan bersama oleh siapa saja yang membutuhkan.
                </p>
                <p class="text-gray-500 text-sm">
                  Demi menghormati privasi, tidak ada data pribadi seperti nomor KTP atau nomor handphone yang dirilis ke repository publik.
                </p>
              </div>
            </div>
          </div>
        </section>

        <!-- Siapa di Balik Dayawarga -->
        <section class="mb-12">
          <h2 class="text-2xl font-bold text-gray-900 mb-6">Siapa di Balik Dayawarga</h2>
          <p class="text-gray-600 mb-6">
            Dayawarga adalah <strong>inisiatif publik</strong>—tumbuh dari kepedulian warga dan relawan. Inisiatif ini bermula dari diskusi dan gerakan kecil di media sosial, yang kemudian berkembang menjadi platform kolaboratif. Nama domain Dayawarga dipilih ketika kami harus menyiapkan domain agar aplikasi ini bisa menyala.
          </p>

          <p class="text-gray-600 mb-4">Untuk informasi dan kontak berkontribusi bisa menghubungi:</p>

          <div class="space-y-4">
            <!-- Fakhrizal -->
            <div class="bg-gray-50 rounded-lg p-4">
              <div class="font-semibold text-gray-900">Fakhrizal</div>
              <div class="text-sm text-gray-500 mb-2">Developer Platform (Jakarta)</div>
              <div class="flex gap-4">
                <a href="https://instagram.com/leksagram" target="_blank" class="flex items-center gap-1.5 text-sm text-gray-600 hover:text-pink-600 transition-colors">
                  <Instagram class="w-4 h-4" />
                  @leksagram
                </a>
                <a href="https://x.com/leksa" target="_blank" class="flex items-center gap-1.5 text-sm text-gray-600 hover:text-blue-500 transition-colors">
                  <Twitter class="w-4 h-4" />
                  @leksa
                </a>
              </div>
            </div>

            <!-- Atina -->
            <div class="bg-gray-50 rounded-lg p-4">
              <div class="font-semibold text-gray-900">Atina Handayani</div>
              <div class="text-sm text-gray-500 mb-2">Community Development Specialist (Yogyakarta)</div>
              <div class="flex gap-4">
                <a href="https://instagram.com/lawasawal" target="_blank" class="flex items-center gap-1.5 text-sm text-gray-600 hover:text-pink-600 transition-colors">
                  <Instagram class="w-4 h-4" />
                  @lawasawal
                </a>
                <a href="https://x.com/tinaberkata" target="_blank" class="flex items-center gap-1.5 text-sm text-gray-600 hover:text-blue-500 transition-colors">
                  <Twitter class="w-4 h-4" />
                  @tinaberkata
                </a>
              </div>
            </div>

            <!-- Ali -->
            <div class="bg-gray-50 rounded-lg p-4">
              <div class="font-semibold text-gray-900">Ali Mulyagusdin</div>
              <div class="text-sm text-gray-500 mb-2">Senior Advisor (Aceh)</div>
              <div class="flex gap-4">
                <a href="https://instagram.com/mulyagusdin" target="_blank" class="flex items-center gap-1.5 text-sm text-gray-600 hover:text-pink-600 transition-colors">
                  <Instagram class="w-4 h-4" />
                  @mulyagusdin
                </a>
              </div>
            </div>
          </div>
        </section>

        <!-- Kolaborasi -->
        <section class="mb-12">
          <h2 class="text-2xl font-bold text-gray-900 mb-6">Kolaborasi Relawan & Komunitas</h2>
          <p class="text-gray-600 mb-6">
            Dayawarga berjalan berkat dukungan dan keterlibatan banyak pihak. Kami berkolaborasi dengan teman-teman relawan dan organisasi, antara lain:
          </p>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <a href="https://instagram.com/karstaceh" target="_blank" class="flex items-center gap-3 bg-white border border-gray-200 rounded-lg p-4 hover:border-blue-300 hover:shadow-sm transition-all">
              <div class="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
                <Users class="w-5 h-5 text-blue-600" />
              </div>
              <div>
                <div class="font-medium text-gray-900">Karst Aceh</div>
                <div class="text-sm text-gray-500 flex items-center gap-1">
                  <Instagram class="w-3 h-3" /> @karstaceh
                </div>
              </div>
            </a>

            <a href="https://instagram.com/leuser.coffee" target="_blank" class="flex items-center gap-3 bg-white border border-gray-200 rounded-lg p-4 hover:border-blue-300 hover:shadow-sm transition-all">
              <div class="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">
                <Users class="w-5 h-5 text-green-600" />
              </div>
              <div>
                <div class="font-medium text-gray-900">Leuser Coffee Community</div>
                <div class="text-sm text-gray-500 flex items-center gap-1">
                  <Instagram class="w-3 h-3" /> @leuser.coffee
                </div>
              </div>
            </a>

            <a href="https://instagram.com/p2lh.aceh" target="_blank" class="flex items-center gap-3 bg-white border border-gray-200 rounded-lg p-4 hover:border-blue-300 hover:shadow-sm transition-all">
              <div class="w-10 h-10 bg-amber-100 rounded-full flex items-center justify-center">
                <Users class="w-5 h-5 text-amber-600" />
              </div>
              <div>
                <div class="font-medium text-gray-900">P2LH-Aceh</div>
                <div class="text-sm text-gray-500 flex items-center gap-1">
                  <Instagram class="w-3 h-3" /> @p2lh.aceh
                </div>
              </div>
            </a>
          </div>

          <p class="text-gray-500 text-sm mt-4">
            Kolaborasi ini memperkuat kerja lapangan, validasi data, serta perspektif lokal yang sangat penting dalam konteks kebencanaan.
          </p>
        </section>

        <!-- Prinsip Kami -->
        <section class="mb-12">
          <h2 class="text-2xl font-bold text-gray-900 mb-6">Prinsip Kami</h2>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="bg-gradient-to-br from-blue-50 to-blue-100 rounded-xl p-5">
              <div class="font-semibold text-blue-900 mb-1">Kolaboratif</div>
              <div class="text-sm text-blue-700">Dibangun bersama, digunakan bersama</div>
            </div>
            <div class="bg-gradient-to-br from-green-50 to-green-100 rounded-xl p-5">
              <div class="font-semibold text-green-900 mb-1">Berbasis data</div>
              <div class="text-sm text-green-700">Peta dan informasi yang dapat dipertanggungjawabkan</div>
            </div>
            <div class="bg-gradient-to-br from-purple-50 to-purple-100 rounded-xl p-5">
              <div class="font-semibold text-purple-900 mb-1">Berpusat pada warga</div>
              <div class="text-sm text-purple-700">Suara dan pengalaman warga adalah inti</div>
            </div>
            <div class="bg-gradient-to-br from-amber-50 to-amber-100 rounded-xl p-5">
              <div class="font-semibold text-amber-900 mb-1">Terbuka & inklusif</div>
              <div class="text-sm text-amber-700">Siapa pun bisa berkontribusi sesuai kapasitasnya</div>
            </div>
          </div>

          <p class="text-gray-600 mt-6">
            Dayawarga percaya bahwa menghadapi bencana bukan hanya soal bertahan, tetapi juga tentang membangun kembali dengan cara yang lebih adil, sadar, dan berdaya.
          </p>
        </section>

        <!-- CTA -->
        <section class="bg-gradient-to-r from-blue-600 to-blue-700 rounded-2xl p-8 text-center text-white">
          <Heart class="w-10 h-10 mx-auto mb-4 opacity-80" />
          <p class="text-lg mb-2">
            Jika kamu warga, relawan, peneliti, jurnalis, atau bagian dari institusi yang ingin berkontribusi—
          </p>
          <p class="text-2xl font-bold">
            Dayawarga adalah ruangmu juga.
          </p>
        </section>

        <!-- Footer spacing -->
        <div class="h-12"></div>
      </div>
    </main>
  </div>
</template>
