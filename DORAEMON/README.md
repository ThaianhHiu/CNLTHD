# URL Shortener - Hệ thống rút gọn link chịu tải cao

## 📋 Giới thiệu

Đồ án môn Công nghệ lập trình hướng đối tượng - Xây dựng hệ thống rút gọn link (URL Shortener) sử dụng Golang với Gin framework, Redis caching và PostgreSQL.

## 🚀 Tính năng chính

- **Rút gọn link nhanh chóng**: Thuật toán sinh mã ngắn unique, không trùng lặp
- **Redirect cực nhanh**: Sử dụng Redis cache để redirect tức thì
- **Analytics bất đồng bộ**: Đếm lượt click không làm chậm request chính
- **Thiết kế chịu tải cao**: Goroutines, Channels, Worker Pool Pattern

## 🏗️ Kiến trúc hệ thống

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Frontend   │────▶│  Gin Server  │────▶│   Redis      │
│   (HTML/JS)  │     │  (Handlers)  │     │   (Cache)    │
└──────────────┘     └──────┬───────┘     └──────────────┘
                            │
                     ┌──────▼───────┐
                     │   Services   │
                     │  (Business)  │
                     └──────┬───────┘
                            │
              ┌─────────────┼─────────────┐
              ▼             ▼             ▼
       ┌──────────┐  ┌──────────┐  ┌──────────────┐
       │ URL Repo │  │Cache Repo│  │Analytics Repo│
       └────┬─────┘  └────┬─────┘  └──────┬───────┘
            │             │               │
            ▼             ▼               ▼
       ┌──────────┐  ┌──────────┐  ┌──────────────┐
       │PostgreSQL│  │  Redis   │  │Click Workers │
       │    DB    │  │  Cache   │  │ (Goroutines) │
       └──────────┘  └──────────┘  └──────────────┘
```

## 📁 Cấu trúc thư mục

```
DORAEMON/
├── config/
│   └── config.go           # Cấu hình ứng dụng
├── database/
│   ├── postgres.go         # Kết nối PostgreSQL
│   └── redis.go            # Kết nối Redis
├── models/
│   ├── url.go              # Model URL và ClickEvent
│   └── dto.go              # Request/Response DTOs
├── interfaces/
│   └── interfaces.go       # Interface definitions
├── repository/
│   ├── url_repository.go   # CRUD operations
│   ├── cache_repository.go # Redis cache operations
│   └── analytics_repository.go
├── generator/
│   └── shortcode.go        # Thuật toán sinh mã ngắn
├── services/
│   └── url_service.go      # Business logic
├── workers/
│   └── click_worker.go     # Async click analytics
├── handlers/
│   └── url_handler.go      # HTTP handlers
├── routes/
│   └── routes.go           # Route definitions
├── static/
│   └── index.html          # Frontend đơn giản
├── main.go                 # Entry point
├── go.mod
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## 🛠️ Công nghệ sử dụng

| Công nghệ | Mục đích |
|-----------|----------|
| **Go 1.21** | Ngôn ngữ lập trình chính |
| **Gin** | Web framework hiệu năng cao |
| **PostgreSQL** | Database lưu trữ URL |
| **Redis** | Cache để redirect nhanh |
| **GORM** | ORM cho PostgreSQL |
| **Docker** | Containerization |

## 🔧 Cài đặt và chạy

### Yêu cầu

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+

### Cách 1: Chạy với Docker (Khuyến nghị)

```bash
# Clone project
cd DORAEMON

# Khởi động tất cả services
docker-compose up -d --build

# Xem logs
docker-compose logs -f app

# Truy cập: http://localhost:8080
```

### Cách 2: Chạy thủ công

```bash
# 1. Cài đặt dependencies
go mod download

# 2. Copy file env
cp .env.example .env
# Sửa thông tin kết nối DB và Redis

# 3. Chạy ứng dụng
go run main.go

# Hoặc build và chạy
go build -o url-shortener
./url-shortener
```

## 📡 API Endpoints

### 1. Tạo Short URL

```http
POST /api/shorten
Content-Type: application/json

{
    "original_url": "https://example.com/very-long-url",
    "custom_code": "mycode",    // Optional
    "expires_in": 24            // Optional: hours
}
```

**Response:**
```json
{
    "short_url": "http://localhost:8080/abc123",
    "short_code": "abc123",
    "original_url": "https://example.com/very-long-url",
    "expires_at": "2024-01-15T10:30:00Z"
}
```

### 2. Redirect

```http
GET /:shortCode
```

Tự động redirect (301) đến URL gốc.

### 3. Xem thống kê

```http
GET /api/stats/:shortCode
```

**Response:**
```json
{
    "short_code": "abc123",
    "original_url": "https://example.com",
    "total_clicks": 1500,
    "created_at": "2024-01-10T08:00:00Z",
    "clicks_by_date": {
        "2024-01-14": 200,
        "2024-01-13": 350
    },
    "top_referers": [
        {"referer": "https://facebook.com", "count": 500}
    ],
    "top_countries": [
        {"country": "Vietnam", "count": 1000}
    ]
}
```

### 4. Xóa URL

```http
DELETE /api/urls/:shortCode
```

## 💡 Điểm nổi bật về kỹ thuật

### 1. Thuật toán sinh mã ngắn (Short Code Generator)

```go
// Kết hợp nhiều nguồn entropy để đảm bảo unique:
// - UUID: Global uniqueness
// - Timestamp (nanoseconds): Temporal uniqueness
// - Counter: Sequential uniqueness
// - Machine ID: Instance uniqueness
// - Cryptographic random: Additional entropy

func (g *ShortCodeGeneratorImpl) Generate() string {
    // Xem chi tiết trong generator/shortcode.go
}
```

### 2. Redis Caching Strategy

```
┌─────────────────────────────────────────────────────────┐
│                    Redirect Flow                         │
├─────────────────────────────────────────────────────────┤
│  Request ──▶ Redis (Cache Hit?) ──▶ Return immediately  │
│                    │                                     │
│                    ▼ (Cache Miss)                        │
│              PostgreSQL ──▶ Cache result ──▶ Return     │
└─────────────────────────────────────────────────────────┘
```

### 3. Async Click Analytics (Goroutines & Channels)

```go
// Worker Pool Pattern
type ClickAnalyticsWorker struct {
    eventChannel chan *ClickEvent  // Buffered channel
    workerCount  int               // Số goroutines
    batchSize    int               // Batch processing
}

// Non-blocking enqueue
func (w *ClickAnalyticsWorker) Enqueue(event *ClickEvent) {
    select {
    case w.eventChannel <- event:
        // Success
    default:
        // Channel full, don't block
    }
}
```

## 📊 Hiệu năng

| Metric | Giá trị |
|--------|---------|
| Redirect latency (cache hit) | < 5ms |
| Redirect latency (cache miss) | < 20ms |
| Throughput | > 10,000 req/s |
| Memory usage | ~50MB |

## 🧪 Testing

```bash
# Chạy tests
go test -v ./...

# Chạy tests với coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📝 Kiến thức áp dụng

### Goroutines và Channels
- Worker Pool Pattern trong `workers/click_worker.go`
- Non-blocking channel operations
- Graceful shutdown với signals

### Struct và Interface
- Interface definitions trong `interfaces/interfaces.go`
- Dependency Injection pattern
- Repository pattern

### RESTful API hiệu năng cao
- Gin framework với middleware
- Redis caching layer
- Async processing

## 👤 Tác giả

- **Sinh viên**: [Tên sinh viên]
- **MSSV**: [Mã số sinh viên]
- **Môn học**: Công nghệ lập trình hướng đối tượng
- **Học kỳ**: HK2 2025-2026

## 📄 License

MIT License - Xem file [LICENSE](LICENSE) để biết thêm chi tiết.
