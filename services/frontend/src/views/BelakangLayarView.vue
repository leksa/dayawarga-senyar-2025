<script setup lang="ts">
import DataLayersSidebar from '@/components/DataLayersSidebar.vue'
import { Database, Smartphone, GitBranch, Unlock, ExternalLink } from 'lucide-vue-next'
</script>

<template>
  <div class="flex-1 flex overflow-hidden">
    <DataLayersSidebar />

    <!-- Belakang Layar Content -->
    <main class="flex-1 bg-gradient-to-b from-gray-50 to-white overflow-y-auto">
      <div class="max-w-3xl mx-auto px-8 py-12">
        <!-- Header -->
        <div class="mb-12">
          <h1 class="text-4xl font-bold text-gray-900 mb-4">Belakang Layar</h1>
          <p class="text-xl text-gray-600 leading-relaxed">
            Bagaimana Dayawarga mengumpulkan, mengolah, dan menyajikan data kebencanaan secara terbuka dan dapat dipertanggungjawabkan.
          </p>
        </div>

        <!-- 1. Sistem Data Collection -->
        <section class="mb-12">
          <div class="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center flex-shrink-0">
                <Database class="w-6 h-6 text-blue-600" />
              </div>
              <div>
                <h2 class="text-xl font-bold text-gray-900 mb-4">1. Sistem Data Collection Berstandar Humanitarian</h2>
                <p class="text-gray-600 mb-4">
                  Dayawarga menggunakan <strong>Open Data Kit (ODK)</strong> dari <a href="https://getodk.org" target="_blank" class="text-blue-600 hover:underline">getODK</a> sebagai tulang punggung sistem pengumpulan data. ODK adalah platform <em>open source</em> yang telah lama digunakan secara luas dalam proyek-proyek kemanusiaan, kesehatan publik, dan riset lapangan di berbagai negara.
                </p>
                <p class="text-gray-600 mb-4">
                  Standar koleksi data yang digunakan mengikuti protokol <strong>OpenRosa</strong> dan <strong>OData</strong>, yang dikembangkan oleh komunitas dan working group internasional yang berfokus pada interoperabilitas dan keterbukaan data.
                </p>
                <p class="text-gray-600 mb-4">
                  Pendekatan ini juga sejalan dengan ekosistem terbuka lain yang digunakan secara global, seperti <strong>Enketo, KoBoToolbox, CommCare, OpenSRP</strong>, hingga <strong>openFHIR</strong> (Fast Healthcare Interoperability Resources) di sektor kesehatan.
                </p>
                <div class="bg-gray-50 rounded-lg p-4">
                  <p class="text-gray-500 text-sm">
                    Artinya, data yang dikumpulkan tidak terjebak dalam sistem tertutup, dan tetap kompatibel untuk kebutuhan lintas platform dan lintas sektor.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 2. ODK Central & ODK Collect -->
        <section class="mb-12">
          <div class="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center flex-shrink-0">
                <Smartphone class="w-6 h-6 text-green-600" />
              </div>
              <div>
                <h2 class="text-xl font-bold text-gray-900 mb-4">2. ODK Central & ODK Collect untuk Kerja Lapangan</h2>
                <p class="text-gray-600 mb-4">
                  Untuk pengelolaan data lapangan, Dayawarga menggunakan:
                </p>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
                  <div class="bg-blue-50 rounded-lg p-4">
                    <div class="font-semibold text-blue-900 mb-1">ODK Central</div>
                    <div class="text-sm text-blue-700">Sistem manajemen survei dan data</div>
                  </div>
                  <div class="bg-green-50 rounded-lg p-4">
                    <div class="font-semibold text-green-900 mb-1">ODK Collect</div>
                    <div class="text-sm text-green-700">Aplikasi Android untuk relawan lapangan</div>
                  </div>
                </div>
                <p class="text-gray-600 mb-4">
                  Melalui <strong>ODK Central</strong>, tim dapat:
                </p>
                <ul class="list-disc list-inside text-gray-600 mb-4 space-y-1 ml-4">
                  <li>Menerbitkan formulir survei</li>
                  <li>Mengatur struktur pertanyaan dan validasi data</li>
                  <li>Mengelola hasil koleksi data secara terpusat</li>
                </ul>
                <p class="text-gray-600 mb-4">
                  Sementara itu, relawan menggunakan <strong>ODK Collect</strong> di perangkat Android untuk mengisi survei langsung di lapangan. Setiap survei dapat dilakukan secara <strong>online maupun offline</strong>, sehingga tetap dapat digunakan di wilayah terdampak bencana yang memiliki keterbatasan atau gangguan akses internet.
                </p>
                <div class="bg-amber-50 border-l-4 border-amber-400 p-4 rounded-r-lg">
                  <p class="text-amber-800 text-sm">
                    Data akan tersimpan aman di perangkat dan dapat dikirimkan saat koneksi tersedia.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 3. Integrasi Data & Pipeline Analitik -->
        <section class="mb-12">
          <div class="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center flex-shrink-0">
                <GitBranch class="w-6 h-6 text-purple-600" />
              </div>
              <div>
                <h2 class="text-xl font-bold text-gray-900 mb-4">3. Integrasi Data & Pipeline Analitik</h2>
                <p class="text-gray-600 mb-4">
                  Selain pengumpulan data primer melalui ODK, Dayawarga juga menyiapkan proses untuk:
                </p>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-6">
                  <div class="flex items-center gap-2 text-gray-600">
                    <span class="w-2 h-2 bg-purple-500 rounded-full"></span>
                    Mengintegrasikan data eksisting dari berbagai lembaga dan sumber terbuka
                  </div>
                  <div class="flex items-center gap-2 text-gray-600">
                    <span class="w-2 h-2 bg-purple-500 rounded-full"></span>
                    Melakukan pemrosesan data GIS untuk kebutuhan pemetaan
                  </div>
                  <div class="flex items-center gap-2 text-gray-600">
                    <span class="w-2 h-2 bg-purple-500 rounded-full"></span>
                    Menjalankan proses ETL (Extract, Transform, Load)
                  </div>
                  <div class="flex items-center gap-2 text-gray-600">
                    <span class="w-2 h-2 bg-purple-500 rounded-full"></span>
                    Menggunakan pendekatan OLAP untuk analisis dan visualisasi
                  </div>
                </div>
                <div class="bg-gray-50 rounded-lg p-4">
                  <p class="text-gray-500 text-sm">
                    Pendekatan ini memungkinkan peta pemulihan yang ditampilkan ke publik tidak hanya akurat secara spasial, tetapi juga kaya konteks dan mudah dipahami oleh berbagai kalangan.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 4. Teknologi Terbuka -->
        <section class="mb-12">
          <div class="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 bg-cyan-100 rounded-xl flex items-center justify-center flex-shrink-0">
                <Unlock class="w-6 h-6 text-cyan-600" />
              </div>
              <div>
                <h2 class="text-xl font-bold text-gray-900 mb-4">4. Teknologi Terbuka & Data yang Dapat Diverifikasi</h2>
                <p class="text-gray-600 mb-4">
                  Seluruh teknologi yang digunakan dalam Aplikasi Relawan Dayawarga berbasis <strong>open source</strong>. Tidak hanya perangkat lunaknya, tetapi juga pendekatan terhadap data.
                </p>
                <p class="text-gray-600 mb-4">
                  Kami membuka akses terhadap:
                </p>
                <div class="grid grid-cols-1 md:grid-cols-3 gap-3 mb-6">
                  <div class="bg-cyan-50 rounded-lg p-4 text-center">
                    <div class="font-semibold text-cyan-900">Data mentah</div>
                    <div class="text-sm text-cyan-700">Raw data</div>
                  </div>
                  <div class="bg-cyan-50 rounded-lg p-4 text-center">
                    <div class="font-semibold text-cyan-900">Data spasial</div>
                    <div class="text-sm text-cyan-700">GIS layers</div>
                  </div>
                  <div class="bg-cyan-50 rounded-lg p-4 text-center">
                    <div class="font-semibold text-cyan-900">Hasil olahan</div>
                    <div class="text-sm text-cyan-700">Processed data</div>
                  </div>
                </div>
                <p class="text-gray-600 mb-4">
                  Baik dalam bentuk data apa adanya (<em>verbatim</em>) maupun hasil pemrosesan. Dengan pendekatan ini, semua pihak dapat:
                </p>
                <ul class="list-disc list-inside text-gray-600 mb-4 space-y-1 ml-4">
                  <li>Melakukan verifikasi</li>
                  <li>Menggunakan kembali data</li>
                  <li>Mengembangkan atau mengekstensifikasi untuk kebutuhan lain</li>
                </ul>
              </div>
            </div>
          </div>
        </section>

        <!-- CTA -->
        <section class="bg-gradient-to-r from-cyan-600 to-blue-700 rounded-2xl p-8 text-center text-white">
          <Unlock class="w-10 h-10 mx-auto mb-4 opacity-80" />
          <p class="text-lg mb-2">
            Kami percaya keterbukaan adalah fondasi kepercayaan,
          </p>
          <p class="text-2xl font-bold">
            terutama dalam kerja kemanusiaan.
          </p>
          <a
            href="https://github.com/leksa/dayawarga-senyar-2025"
            target="_blank"
            class="inline-flex items-center gap-2 mt-6 px-6 py-3 bg-white/20 hover:bg-white/30 rounded-lg transition-colors"
          >
            <span>Lihat di GitHub</span>
            <ExternalLink class="w-4 h-4" />
          </a>
        </section>

        <!-- Footer spacing -->
        <div class="h-12"></div>
      </div>
    </main>
  </div>
</template>
