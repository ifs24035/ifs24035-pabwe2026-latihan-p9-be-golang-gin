# Abdullah Ubaid - PABWE 2026 Praktikum 9 — Todolist Backend (Golang Gin)

Backend RESTful API untuk aplikasi Todolist dengan arsitektur **Feature-First Approach & TDD (100% Code Coverage)**, dibangun menggunakan **Go (v1.27.4)**, **Gin Web Framework**, basis data **PostgreSQL (`pabwe_p9_todolist`)**, dokumentasi interaktif **Swagger UI**, dan integrasi kecerdasan buatan (**OpenAI Todo Generator**).

---

## 🚀 Fitur Utama Backend

1. **Feature-First Architecture**:
   - Struktur modul dikelompokkan berdasarkan domain bisnis: `auth`, `todo`, `ai`, `admin`.
   - Menerapkan Dependency Inversion dengan repository interface dan service interface.
   - Pemisahan tanggung jawab yang jelas antara Handler, Service, Repository, DTO, dan Middleware.

2. **Kelompok Todo (`TodoGroup`) & Template Publik**:
   - CRUD Kelompok Todo (`/api/todo-groups`).
   - Publikasi template (`is_public = true`).
   - Akses template publik tanpa login (`/api/templates` dan `/api/templates/:id`).
   - Kloning template publik menjadi daftar tugas pribadi (`POST /api/templates/:id/clone`).

3. **Autentikasi & Autorisasi (JWT)**:
   - Login menggunakan username atau email.
   - Atribut ID form login terstandar: `id="login-username"`, `id="login-password"`, `id="login-submit"`.
   - Role-Based Access Control: `admin` (monitoring seluruh user) dan `member` (pengguna biasa).

4. **AI Todo Generator (OpenAI)**:
   - Menghasilkan daftar rekomendasi todolist otomatis melalui endpoint OpenAI kompatibel (`POST /api/todos/ai-generate`).
   - Konfigurasi via `.env`: `OPENAI_ENDPOINT`, `OPENAI_KEY`, `OPENAI_MODEL`.

5. **Dokumentasi OpenAPI / Swagger UI**:
   - Tersedia dokumentasi interaktif di `/swagger/index.html` dengan spesifikasi OpenAPI `api.json` yang di-embed langsung ke binary executable (dapat diakses di `/swagger/api.json` atau `/api.json`).
   - Kontak pengembang: **Delcom - Labs**.

6. **Pure RESTful API & CORS**:
   - Backend murni mengekspos endpoint RESTful API dengan header CORS aktif, siap di-consume oleh Frontend ReactJS yang dideploy terpisah.

7. **Format Standar Respons API**:
   - Seluruh endpoint mengembalikan format JSON standar:
     ```json
     {
       "status": "success|fail|error",
       "message": "Pesan deskripsi",
       "data": { ... }
     }
     ```
   - Field `"data"` bersifat opsional (hanya disertakan saat endpoint mengembalikan data entitas).

8. **100% Test Coverage**:
   - Seluruh package `internal/config`, `internal/database`, `internal/middleware`, `internal/response`, `internal/features/...`, dan `internal/server` lulus pengujian dengan **100.0% Statement Coverage**.

---

## ⚙️ Konfigurasi Environment (`.env`)

```env
APP_PORT=7071
APP_ENV=development
APP_KEY=pabwe-secret-key-2026-todolist-golang-gin

# Default Admin Credentials
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin

# Database Configuration (Laravel-Style)
DB_CONNECTION=pgsql
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=pabwe_p9_todolist
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable

# OpenAI Configuration
OPENAI_ENDPOINT=https://delcom.org/api/v1
OPENAI_KEY=key
OPENAI_MODEL=nama-model
```

---

## 🧪 Menjalankan Pengujian (TDD 100% Coverage)

```bash
# Menjalankan seluruh pengujian unit & integrasi dengan laporan coverage:
go test -v -cover ./internal/...
```

---

## 🏃 Menjalankan Server (Mode Development)

```bash
# Menjalankan langsung melalui Go compiler:
go run main.go
```
- API Server aktif pada: `http://localhost:7071`
- Dokumentasi Swagger UI: `http://localhost:7071/swagger/index.html` (Spesifikasi JSON: `/swagger/api.json` atau `/api.json`)

---

## 📦 Mengompilasi (*Build*) & Menjalankan Hasil Kompilasi

Kompilasi kode sumber menjadi berkas binary mandiri (*standalone executable*) untuk keperluan produksi atau *deployment*:

### 1. Kompilasi (*Build*)
```bash
# Kompilasi untuk sistem operasi lokal (Windows):
go build -o server.exe main.go
```

> **Opsi Cross-Compile (misal untuk Server Linux/VPS):**
> ```bash
> # Di Windows PowerShell:
> $env:GOOS="linux"; $env:GOARCH="amd64"; go build -o server main.go
>
> # Di Windows Command Prompt (CMD):
> set GOOS=linux
> set GOARCH=amd64
> go build -o server main.go
> ```

### 2. Menjalankan Hasil *Build*
```bash
# Jalankan biner yang sudah dikompilasi (Windows):
.\server.exe

# Jika di Linux:
./server
```

> **Catatan Penting:**
> Pastikan berkas `.env` berada di direktori yang sama dengan berkas `server.exe` saat dijalankan agar konfigurasi basis data dan port aplikasi terbaca dengan benar.
