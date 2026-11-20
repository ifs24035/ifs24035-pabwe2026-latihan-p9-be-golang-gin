package database

import (
	"fmt"
	"time"

	"app/internal/config"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User model for database package migrations
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string    `gorm:"not null;default:'member'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TodoGroup struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	IsPublic    bool      `gorm:"default:false" json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Todos       []Todo    `gorm:"foreignKey:TodoGroupID" json:"todos,omitempty"`
}

type Todo struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	TodoGroupID    *uint      `gorm:"index" json:"todo_group_id,omitempty"`
	Title          string     `gorm:"not null" json:"title"`
	Description    string     `json:"description"`
	Status         string     `gorm:"not null;default:'pending'" json:"status"` // pending, in_progress, completed
	DueDate        *time.Time `json:"due_date,omitempty"`
	CoverImagePath string     `json:"cover_image_path,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	User           *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func Connect(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	if cfg.DBConnection == "sqlite" || cfg.DBConnection == "sqlite3" {
		dialector = sqlite.Open(cfg.DBDatabase)
	} else {
		dialector = postgres.Open(cfg.DSN())
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

func ConnectTest() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	_ = Migrate(db)
	return db, err
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &TodoGroup{}, &Todo{})
}

func Seed(db *gorm.DB, cfgs ...*config.Config) error {
	cfg := config.LoadConfig()
	if len(cfgs) > 0 && cfgs[0] != nil {
		cfg = cfgs[0]
	}
	adminUsername := cfg.AdminUsername
	adminPassword := cfg.AdminPassword

	// Seed Admin
	var admin User
	err := db.Where("username = ?", adminUsername).First(&admin).Error
	if err != nil {
		hashedAdmin, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		admin = User{
			Username: adminUsername,
			Email:    adminUsername + "@delcom.org",
			Password: string(hashedAdmin),
			Role:     "admin",
		}
		_ = db.Create(&admin).Error
	}

	// Seed Member
	var member User
	err = db.Where("username = ?", "member").First(&member).Error
	if err != nil {
		hashedMember, _ := bcrypt.GenerateFromPassword([]byte("member"), bcrypt.DefaultCost)
		member = User{
			Username: "member",
			Email:    "member@delcom.org",
			Password: string(hashedMember),
			Role:     "member",
		}
		_ = db.Create(&member).Error
	}

	// Check total existing todos (seed 100 default todos: 50 in templates, 50 personal)
	var todoCount int64
	db.Model(&Todo{}).Count(&todoCount)
	if todoCount == 0 {
		now := time.Now()
		due1 := now.Add(24 * time.Hour)
		due2 := now.Add(48 * time.Hour)
		due3 := now.Add(72 * time.Hour)
		due7 := now.Add(7 * 24 * time.Hour)

		// 1. Seed 5 Public Templates with 10 Todos each (Total = 50 Template Todos)
		templateDefs := []struct {
			Title       string
			Description string
			Todos       []struct {
				Title       string
				Description string
				Status      string
				DueDate     *time.Time
			}
		}{
			{
				Title:       "Onboarding Software Engineer Baru",
				Description: "Checklist langkah awal dan pengenalan lingkungan pengembangan software engineer baru.",
				Todos: []struct {
					Title       string
					Description string
					Status      string
					DueDate     *time.Time
				}{
					{"Setup Go 1.27 environment dan Go workspace", "Pasang Go compiler, gopls, dan atur GOPATH/GOROOT secara tepat.", "completed", &due1},
					{"Konfigurasi PostgreSQL dan SQLite driver lokal", "Pastikan database lokal berjalan dan kredensial teruji di .env.", "completed", &due1},
					{"Clone repositori proyek dan pelajari struktur Feature-First", "Pelajari pemisahan domain pada internal/features.", "completed", &due2},
					{"Setup Docker dan containerized test environment", "Jalankan docker-compose untuk service database dan mocking.", "in_progress", &due2},
					{"Pelajari format standar API Response", "Pahami struktur JSON status (success/fail/error), message, dan data.", "in_progress", &due2},
					{"Pelajari standard middleware CORS, JWT Auth, dan Rate Limiting", "Review alur verifikasi token dan proteksi rute.", "in_progress", &due3},
					{"Setup mock server dan Postman / Swagger collection", "Impor openapi spec ke Postman untuk pengujian endpoint.", "pending", &due3},
					{"Tulis unit testing dasar dengan Testify", "Implementasikan test assertion untuk repository dan handler.", "pending", &due7},
					{"Ikuti sesi code review pertama dengan Tech Lead", "Diskusikan konvensi penamaan kode dan git branching strategy.", "pending", &due7},
					{"Selesaikan tiket bug onboarding perdana", "Perbaiki issue kecil pada backlog untuk adaptasi alur kerja.", "pending", &due7},
				},
			},
			{
				Title:       "Sprint Scrum & Delivery Produk",
				Description: "Alur kerja siklus sprint agile scrum dari perencanaan hingga retrospective.",
				Todos: []struct {
					Title       string
					Description string
					Status      string
					DueDate     *time.Time
				}{
					{"Sprint Planning: Breakdown backlog item menjadi task teknis", "Pecah user stories menjadi subtugas spesifik dan terukur.", "completed", &due1},
					{"Estimasi story points bersama tim pengembang", "Gunakan planning poker untuk menentukan bobot kesulitan task.", "completed", &due1},
					{"Daily Standup: Sinkronisasi blocker dan progres sprint harian", "Update status pekerjaan: apa yang dikerjakan dan kendala yang dihadapi.", "completed", &due2},
					{"Implementasi user story prioritas tinggi", "Selesaikan fitur utama sesuai kriteria penerimaan (acceptance criteria).", "in_progress", &due2},
					{"Setup Unit Test & Integration Test dengan 100% coverage", "Pastikan seluruh skenario branch dan error handling teruji.", "in_progress", &due3},
					{"Buat Pull Request dan assign reviewer tim", "Sertakan deskripsi perubahan, checklist pengujian, dan screenshot.", "pending", &due3},
					{"Selesaikan catatan review dan merge ke branch staging", "Lakukan revisi kode sesuai masukan reviewer sebelum squash merge.", "pending", &due7},
					{"Smoke test & regression test pada environment staging", "Verifikasi fitur berjalan normal tanpa merusak fungsionalitas lain.", "pending", &due7},
					{"Sprint Review & Demo aplikasi ke stakeholders", "Demonstrasikan fungsionalitas yang telah selesai di hadapan klien.", "pending", &due7},
					{"Sprint Retrospective: Evaluasi proses kerja dan action item", "Diskusikan apa yang berjalan baik dan hal yang perlu ditingkatkan.", "pending", &due7},
				},
			},
			{
				Title:       "Belajar Fullstack React & Go Gin",
				Description: "Panduan terstruktur belajar pengembangan web modern menggunakan Golang Gin dan React.",
				Todos: []struct {
					Title       string
					Description string
					Status      string
					DueDate     *time.Time
				}{
					{"Pahami arsitektur Clean Architecture & Feature-First di Golang", "Pelajari struktur folder berlapis dan pemisahan concerns.", "completed", &due1},
					{"Pelajari GORM routing, preloading, dan automigration", "Kuasai relasi antar tabel One-to-Many dan Many-to-Many di GORM.", "completed", &due1},
					{"Implementasi JWT Authentication dan Password Hashing Bcrypt", "Bangun autentikasi aman dengan stateless JSON Web Token.", "completed", &due2},
					{"Bangun RESTful API dengan Gin engine", "Buat endpoint CRUD dengan validasi struct Gin binding.", "in_progress", &due2},
					{"Setup React 19 dengan Vite dan Tailwind CSS v4", "Inisialisasi frontend dengan Tailwind CSS v4 custom variant.", "in_progress", &due3},
					{"Buat custom hook useTheme untuk Dark/Light mode dan aksen warna", "Kelola mode tema dan warna primer berbasis CSS variable.", "in_progress", &due3},
					{"Integrasikan React frontend dengan Gin backend menggunakan Fetch API", "Konsumsi endpoint REST API dan simpan state secara reaktif.", "pending", &due3},
					{"Tangani error handling dan standard response di frontend", "Tampilkan feedback informatif saat request berhasil atau gagal.", "pending", &due7},
					{"Implementasi Vitest testing dengan React Testing Library", "Tulis tes komponen dan simulasi aksi pengguna dengan userEvent.", "pending", &due7},
					{"Build production bundle dan deploy aplikasi fullstack", "Uji hasil build statis di production mode dengan performa optimal.", "pending", &due7},
				},
			},
			{
				Title:       "Persiapan Ujian & Praktikum PABWE",
				Description: "Checklist belajar intensif dan latihan mandiri persiapan evaluasi praktikum PABWE.",
				Todos: []struct {
					Title       string
					Description string
					Status      string
					DueDate     *time.Time
				}{
					{"Pelajari slide materi HTML5, CSS3, dan Responsive Design", "Pahami layout flexbox, grid, dan prinsip responsivitas mobile-first.", "completed", &due1},
					{"Review implementasi React Hooks: useState, useEffect, useCallback", "Pahami siklus hidup komponen dan manajemen efek samping.", "completed", &due1},
					{"Latihan membuat dynamic form dengan validasi input", "Bangun validasi client-side untuk email, password, dan field wajib.", "completed", &due2},
					{"Pahami konsep asynchronous JavaScript: Promise dan async/await", "Pelajari alur eksekusi asinkron dan error handling try-catch.", "in_progress", &due2},
					{"Review materi Golang syntax, struct, interface, dan goroutine", "Pahami keunggulan concurrency dan type safety di Golang.", "in_progress", &due3},
					{"Latihan routing dan middleware pada framework Gin", "Pahami penanganan request context dan pipeline middleware.", "in_progress", &due3},
					{"Praktikkan integrasi database relasional dengan ORM", "Latihan membuat query filter, sorting, dan pagination dengan GORM.", "pending", &due3},
					{"Pelajari teknik penulisan unit test untuk mencapai 100% code coverage", "Pastikan coverage statements, branch, functions, dan lines 100%.", "pending", &due7},
					{"Simulasi pengerjaan studi kasus praktikum dalam batas waktu", "Uji kecepatan koding dan penyelesaian fitur dari awal hingga akhir.", "pending", &due7},
					{"Review checklist penilaian dan standarisasi koding", "Pastikan penulisan kode rapi, bebas warning linter, dan terstruktur.", "pending", &due7},
				},
			},
			{
				Title:       "Launch & Deployment Checklist Production",
				Description: "Daftar periksa kesiapan peluncuran sistem ke lingkungan produksi server.",
				Todos: []struct {
					Title       string
					Description string
					Status      string
					DueDate     *time.Time
				}{
					{"Audit keamanan: Verifikasi proteksi CORS dan JWT secret key", "Pastikan secret key menggunakan string acak dengan entropi tinggi.", "completed", &due1},
					{"Konfigurasi environment variables (.env) production", "Sesuaikan host, port, kredensial database, dan mode release.", "completed", &due1},
					{"Setup automated database migration dan backup otomatis", "Konfigurasi script migrasi skema dan jadwal dump basis data berkala.", "completed", &due2},
					{"Konfigurasi HTTPS / SSL certificate menggunakan Let's Encrypt", "Pasang sertifikat enkripsi TLS untuk mengamankan data pengguna.", "in_progress", &due2},
					{"Setup reverse proxy Nginx untuk routing frontend dan backend", "Atur upstream server, caching static file, dan gzip compression.", "in_progress", &due3},
					{"Optimasi asset frontend: Minifikasi CSS, JS, dan kompresi gambar", "Pastikan ukuran bundle kecil agar waktu muat halaman cepat.", "in_progress", &due3},
					{"Setup Health Check endpoint (/health) untuk monitoring uptime", "Sediakan probe endpoint untuk pengecekan status server berkala.", "pending", &due3},
					{"Konfigurasi centralized logging dan alert error", "Integrasikan error reporting untuk deteksi masalah secara dini.", "pending", &due7},
					{"Lakukan stress test & load testing dengan k6 atau Apache Bench", "Uji daya tahan backend pada beban ratusan request serentak.", "pending", &due7},
					{"Final verification checklist dan rilis versi v1.0.0 ke publik", "Pastikan seluruh fungsi bekerja sempurna sebelum pengumuman rilis.", "pending", &due7},
				},
			},
		}

		for _, tDef := range templateDefs {
			grp := TodoGroup{
				UserID:      member.ID,
				Title:       tDef.Title,
				Description: tDef.Description,
				IsPublic:    true,
			}
			_ = db.Create(&grp).Error
			for _, tItem := range tDef.Todos {
				todo := Todo{
					UserID:      member.ID,
					TodoGroupID: &grp.ID,
					Title:       tItem.Title,
					Description: tItem.Description,
					Status:      tItem.Status,
					DueDate:     tItem.DueDate,
				}
				_ = db.Create(&todo).Error
			}
		}

		// 2. Seed 50 Personal Todos (Total = 50 Personal Todos, TodoGroupID = nil)
		personalTodos := []struct {
			Title       string
			Description string
			Status      string
			DueDate     *time.Time
		}{
			{"Review PR arsitektur modular backend Gin", "Periksa pemisahan handler, service, dan repository pada modul todo.", "completed", &due1},
			{"Riset model OpenAI untuk generator todolist otomatis", "Uji prompt engineering JSON output dengan model gpt-4o-mini.", "completed", &due1},
			{"Perbarui dokumentasi API pada Swagger/OpenAPI", "Lengkapi schema request body dan contoh response status 200.", "completed", &due1},
			{"Desain wireframe antarmuka mode gelap dan tema aksen", "Eksplorasi kontras palet warna gelap dan variasi 5 warna tema.", "completed", &due1},
			{"Optimasi performa query database dengan indexing", "Tambahkan index pada user_id dan todo_group_id di tabel todos.", "completed", &due1},
			{"Perbaiki format response API agar memiliki status, message, dan data", "Standarisasi helper jsonResponse untuk format terpadu.", "completed", &due2},
			{"Buat favicon resolusi multi-device untuk frontend", "Generate favicon.png, apple-touch-icon.png, dan logo SVG.", "completed", &due2},
			{"Tambahkan variabel ADMIN_USERNAME dan ADMIN_PASSWORD di .env", "Pastikan default admin dan password tersimpan di config.", "completed", &due2},
			{"Tulis unit test useTheme hook di React", "Uji inisialisasi state, toggle dark mode, dan pergantian aksen.", "completed", &due2},
			{"Uji coba responsivitas layout modal di layar tablet dan mobile", "Pastikan modal tidak terpotong dan nyaman diakses di layar sentuh.", "completed", &due2},
			{"Refactor komponen StickyNavbar dengan tema warna dinamis", "Integrasikan warna aksen dan mode gelap pada sticky header.", "completed", &due2},
			{"Verifikasi code coverage Vitest mencapai 100%", "Jalankan vitest run --coverage untuk memastikan semua baris teruji.", "completed", &due2},
			{"Verifikasi code coverage Go test mencapai 100%", "Jalankan go test -cover ./... pada seluruh internal package.", "completed", &due3},
			{"Tambahkan rate limiting middleware pada rute autentikasi", "Cegah brute-force login dengan pembatasan 10 request per menit.", "completed", &due3},
			{"Uji integritas database SQLite dan PostgreSQL", "Validasi keselarasan tipe data antara SQLite lokal dan PostgreSQL.", "completed", &due3},
			{"Buat panduan instalasi dan build pada README.md", "Tuliskan instruksi langkah demi langkah menjalankan aplikasi.", "in_progress", &due2},
			{"Implementasi token expiration check pada frontend", "Tangani token kedaluwarsa dengan auto-logout yang mulus.", "in_progress", &due2},
			{"Review security checklist password hashing bcrypt", "Pastikan cost factor bcrypt aman dan salt digenerate secara acak.", "in_progress", &due2},
			{"Sempurnakan navigasi breadcrumb dan pagination", "Bantu pengguna bernavigasi antar kelompok todolist dengan jelas.", "in_progress", &due3},
			{"Uji performa caching HTTP header pada static assets", "Atur Cache-Control header untuk file JS, CSS, dan gambar.", "in_progress", &due3},
			{"Analisis log performa query SQL berulang", "Cari potensi N+1 query problem dan selesaikan dengan preloading.", "in_progress", &due3},
			{"Uji ketahanan aplikasi terhadap serangan SQL Injection", "Pastikan seluruh query menggunakan parameter binding GORM.", "in_progress", &due3},
			{"Tambahkan validasi batas ukuran file unggahan", "Batasi upload gambar cover todo maksimal 2MB per berkas.", "in_progress", &due3},
			{"Buat script sinkronisasi data antar environment", "Otomasi seeding data dummy untuk staging dan development.", "in_progress", &due7},
			{"Implementasi graceful shutdown pada server HTTP Gin", "Tangani SIGINT dan SIGTERM agar request aktif selesai terlebih dahulu.", "in_progress", &due7},
			{"Siapkan materi presentasi asistensi praktikum minggu ke-9", "Buat ringkasan arsitektur fullstack dan best practice koding.", "pending", &due3},
			{"Buat rubrik penilaian tugas mahasiswa untuk modul React dan Go", "Tentukan kriteria kelengkapan fitur, testing, dan kerapian kode.", "pending", &due3},
			{"Lakukan backup database mingguan ke cloud storage", "Jadwalkan script cron backup basis data otomatis tiap akhir pekan.", "pending", &due3},
			{"Evaluasi feedback pengguna terkait tampilan antarmuka modal", "Kumpulkan masukan seputar kenyamanan navigasi dan kontras teks.", "pending", &due3},
			{"Rencanakan fitur export todolist ke format PDF dan CSV", "Sediakan opsi unduh daftar tugas untuk keperluan cetak laporan.", "pending", &due3},
			{"Riset implementasi WebSocket untuk kolaborasi todolist realtime", "Pelajari channel message broadcast antar pengguna terhubung.", "pending", &due7},
			{"Tambahkan filter pencarian berdasarkan prioritas dan due date", "Bantu pengguna memfilter tugas yang mendekati batas tenggat.", "pending", &due7},
			{"Setup CI/CD pipeline dengan GitHub Actions", "Otomasi linting, build, dan testing setiap ada Pull Request baru.", "pending", &due7},
			{"Konfigurasi alert webhook Discord untuk notifikasi error server", "Kirim notifikasi instan saat terjadi panic atau server 500.", "pending", &due7},
			{"Audit dependensi npm dan Go module dari vulnerability", "Jalankan npm audit dan govulncheck secara teratur.", "pending", &due7},
			{"Buat tutorial interaktif penggunaan AI todolist generator", "Pandu pengguna baru cara menyusun prompt agar hasil todolist akurat.", "pending", &due7},
			{"Refactor styling button agar konsisten di semua modal", "Gunakan class utility seragam untuk varian tombol aksi.", "pending", &due7},
			{"Riset integrasi kalender Google untuk sync jadwal tugas", "Eksplorasi OAuth2 Google Calendar API untuk sinkronisasi due date.", "pending", &due7},
			{"Sederhanakan alur onboarding pengguna baru di aplikasi", "Buat tur panduan interaktif pertama kali pengguna masuk dashboard.", "pending", &due7},
			{"Implementasi debounce pada input pencarian frontend", "Tunda pemanggilan API pencarian selama 300ms setelah user mengetik.", "pending", &due7},
			{"Tambahkan fitur arsip untuk tugas yang sudah selesai lama", "Pindahkan tugas tuntas lebih dari 30 hari ke arsip riwayat.", "pending", &due7},
			{"Review penggunaan memori aplikasi pada traffic tinggi", "Profil heap dan goroutine menggunakan pprof untuk cegah memory leak.", "pending", &due7},
			{"Update dependensi Vite ke versi LTS terbaru", "Manfaatkan perbaikan bug dan peningkatan kecepatan bundler Vite.", "pending", &due7},
			{"Rancang skema tagging multi-kategori untuk tugas", "Izinkan satu tugas memiliki lebih dari satu label penanda.", "pending", &due7},
			{"Buat script otomasi data dummy untuk load testing", "Generate ribuan data todolist sintetis untuk stress test API.", "pending", &due7},
			{"Pelajari best practice arsitektur micro frontends", "Riset modularitas komponen frontend berskala enterprise.", "pending", &due7},
			{"Sempurnakan accessibility (a11y) pada elemen dialog dan tombol", "Tambahkan atribut aria, focus trap, dan navigasi keyboard lengkap.", "pending", &due7},
			{"Eksplorasi penggunaan Redis untuk distributed caching", "Simpan hasil query public templates pada in-memory cache Redis.", "pending", &due7},
			{"Uji kompatibilitas cross-browser pada Safari, Chrome, dan Firefox", "Pastikan tampilan CSS dan fungsionalitas konsisten di semua browser.", "pending", &due7},
			{"Kumpulkan laporan akhir implementasi praktikum 9 PABWE", "Finalisasi dokumentasi pengujian, tangkapan layar, dan link repositori.", "pending", &due7},
		}

		for _, pItem := range personalTodos {
			todo := Todo{
				UserID:      member.ID,
				TodoGroupID: nil,
				Title:       pItem.Title,
				Description: pItem.Description,
				Status:      pItem.Status,
				DueDate:     pItem.DueDate,
			}
			_ = db.Create(&todo).Error
		}
	}

	return nil
}
