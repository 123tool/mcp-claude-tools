## MCP Claude Tools Server (Remote HTTP)

`mcp-claude-tools` adalah server **Model Context Protocol (MCP)** berperforma tinggi yang ditulis dalam bahasa Go (Golang). Server ini mengekspos perkakas (*tools*) manipulasi berkas dan eksekusi shell bawaan Claude Code melalui protokol HTTP berbasis JSON-RPC 2.0. 

Dengan server ini, klien MCP mana pun dapat memanggil fungsi sistem, mengelola berkas, dan mengeksekusi perintah shell dari jarak jauh secara aman dan efisien.

---

## ✨ Fitur

## 💻 Manajemen Shell & Bash
* **`bash`**: Mengekspos eksekusi perintah shell lokal dengan dukungan batasan waktu (*timeout*) dan opsi berjalan di latar belakang (*background process*).
* **`bash_output`**: Mengambil aliran keluaran (`stdout`/`stderr`) dari proses latar belakang yang sedang berjalan berdasarkan ID tertentu.
* **`kill_shell`**: Menghentikan paksa proses shell latar belakang yang tidak responsif atau sudah selesai.

## 📁 Berkas (File Tools)

* **`baca`**: Membaca isi berkas teks dengan dukungan parameter `offset_lines` dan `limit_lines` untuk efisiensi memori token.
* **`tulis`**: Menulis ulang atau membuat berkas baru langsung ke dalam penyimpanan disk.
* **`sunting`**: Melakukan penggantian string/teks secara presisi (*exact match replacement*) di dalam berkas tanpa merusak struktur lain.
* **`glob`**: Mencari jalur file atau folder menggunakan pola pencarian wildcards (globbing pattern).
* **`grep`**: Pencarian teks regex super cepat di dalam direktori kerja menggunakan utilitas pihak ketiga `ripgrep (rg)`.

---

## 🔒 Fitur

* **Validasi Jalur (Anti-Directory Traversal):** Server menolak keras jalur relatif (`../`). Semua interaksi berkas divalidasi ketat agar tetap berada di bawah kendali direktori kerja utama (`BaseDir`).
* **Proteksi Slowloris & Timeout HTTP:** Menggunakan `ReadHeaderTimeout` dan `IdleTimeout` pada server HTTP native Go untuk mencegah serangan penolakan layanan (DoS).
* **Batasan Ukuran Berkas:** Membatasi pembacaan berkas maksimal **10MB** untuk mencegah kelebihan beban penggunaan RAM dan memori server.
* **Graceful Shutdown:** Menangkap sinyal OS (`SIGINT`, `SIGTERM`) untuk menyelesaikan sisa proses HTTP yang menggantung sebelum mematikan server sepenuhnya.

---

## 🚀 Panduan

## Prasyarat Sistem
* **Go** versi 1.21 ke atas.
* **Ripgrep (`rg`)** terinstal di sistem Anda jika ingin menggunakan fitur pencarian teks mendalam (`grep`).

## Instalasi

1. Klon repositori ini ke mesin lokal atau server remote Anda:
   ```
   git clone https://github.com/123tool/mcp-claude-tools.git
   cd mcp-claude-tools
2. Unduh dependency dan rapikan modul proyek :
   ```
   go mod tidy
3. Jalankan server MCP :
   ```
   go run main.go
4. server akan berjalan pada alamat `http://localhost:8080/mcp` dan menjadikan direktori saat ini sebagai basis folder aman.

## Panduan Integrasi (Spesifikasi JSON-RPC 2.0)

​Klien MCP dapat berinteraksi dengan mengirimkan permintaan `HTTP POST` ke endpoint `/mcp`.

**​Contoh 1 :**
Menjalankan Perintah Bash (Sinkron)
​Request Payload :
```
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "bash",
    "arguments": {
      "command": "echo 'Halo Dunia Pro!' && uname -a",
      "timeout": 10,
      "run_in_bg": false
    }
  },
  "id": 101
}
```
**Response Sukses :**
```
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Halo Dunia Pro!\nLinux server-pro 5.15.0-x86_64...\n"
      }
    ],
    "isError": false
  },
  "id": 101
}
```
**Contoh 2 :**
Menulis Berkas Baru (tulis) ​Request Payload :
```
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "tulis",
    "arguments": {
      "path": "/absolute/path/to/project/config.json",
      "content": "{\n  \"status\": \"active\"\n}"
    }
  },
  "id": 102
}
