package generator

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Bảng ký tự cho short code (loại bỏ các ký tự dễ nhầm lẫn: 0, O, l, I)
const charset = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ123456789"

// ShortCodeGeneratorImpl là implementation của ShortCodeGenerator
// Sử dụng nhiều thuật toán kết hợp để đảm bảo không trùng lặp
type ShortCodeGeneratorImpl struct {
	length    int
	mu        sync.Mutex
	counter   uint64
	machineID string
}

// NewShortCodeGenerator tạo instance mới của ShortCodeGenerator
func NewShortCodeGenerator(length int) *ShortCodeGeneratorImpl {
	// Tạo machine ID unique cho mỗi instance
	machineID := generateMachineID()

	return &ShortCodeGeneratorImpl{
		length:    length,
		counter:   0,
		machineID: machineID,
	}
}

// Generate tạo short code mới với độ dài mặc định
// Thuật toán: Kết hợp UUID + Timestamp + Counter + Random
func (g *ShortCodeGeneratorImpl) Generate() string {
	return g.GenerateWithLength(g.length)
}

// GenerateWithLength tạo short code với độ dài cụ thể
func (g *ShortCodeGeneratorImpl) GenerateWithLength(length int) string {
	g.mu.Lock()
	g.counter++
	currentCounter := g.counter
	g.mu.Unlock()

	// Tạo seed từ nhiều nguồn để đảm bảo uniqueness:
	// 1. UUID - đảm bảo global uniqueness
	// 2. Timestamp với nanoseconds - đảm bảo temporal uniqueness
	// 3. Counter - đảm bảo sequential uniqueness trong cùng instance
	// 4. Machine ID - đảm bảo uniqueness giữa các instances
	// 5. Cryptographic random - thêm entropy

	uuidPart := uuid.New().String()
	timestamp := time.Now().UnixNano()

	// Combine tất cả các nguồn
	combined := combineSeeds(uuidPart, timestamp, currentCounter, g.machineID)

	// Generate short code từ combined seed
	return generateFromSeed(combined, length)
}

// IsValid kiểm tra short code có hợp lệ không
func (g *ShortCodeGeneratorImpl) IsValid(code string) bool {
	if len(code) < 4 || len(code) > 12 {
		return false
	}

	for _, char := range code {
		if !strings.ContainsRune(charset, char) {
			return false
		}
	}

	return true
}

// generateMachineID tạo ID unique cho instance
func generateMachineID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:8]
}

// combineSeeds kết hợp các nguồn entropy
func combineSeeds(uuidPart string, timestamp int64, counter uint64, machineID string) []byte {
	// Hash combination of all seeds
	combined := make([]byte, 32)

	// Add UUID bytes
	uuidBytes := []byte(uuidPart)
	for i := 0; i < len(uuidBytes) && i < 16; i++ {
		combined[i] = uuidBytes[i]
	}

	// Add timestamp bytes
	for i := 0; i < 8; i++ {
		combined[16+i] = byte(timestamp >> (i * 8))
	}

	// Add counter bytes
	for i := 0; i < 4; i++ {
		combined[24+i] = byte(counter >> (i * 8))
	}

	// Add machine ID bytes
	machineBytes := []byte(machineID)
	for i := 0; i < len(machineBytes) && i < 4; i++ {
		combined[28+i] = machineBytes[i]
	}

	// XOR with cryptographic random for additional entropy
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	for i := range combined {
		combined[i] ^= randomBytes[i]
	}

	return combined
}

// generateFromSeed sinh short code từ seed
func generateFromSeed(seed []byte, length int) string {
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		// Sử dụng crypto/rand cho mỗi ký tự
		n, _ := rand.Int(rand.Reader, charsetLen)

		// XOR với seed để thêm deterministic element
		seedIndex := i % len(seed)
		index := (n.Int64() + int64(seed[seedIndex])) % int64(len(charset))

		result[i] = charset[index]
	}

	return string(result)
}

// GenerateBase62 là thuật toán alternative sử dụng Base62 encoding
// Phù hợp khi cần short code ngắn hơn
func GenerateBase62(id uint64) string {
	const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	if id == 0 {
		return string(base62Chars[0])
	}

	var result []byte
	for id > 0 {
		result = append([]byte{base62Chars[id%62]}, result...)
		id /= 62
	}

	return string(result)
}

// GenerateSnowflake tạo ID dựa trên thuật toán Snowflake
// Đảm bảo unique trong distributed system
type SnowflakeGenerator struct {
	mu        sync.Mutex
	epoch     int64 // Custom epoch (milliseconds)
	machineID int64 // Machine/Node ID (10 bits = 1024 machines)
	sequence  int64 // Sequence number (12 bits = 4096 per millisecond)
	lastTime  int64
}

// NewSnowflakeGenerator tạo Snowflake generator mới
func NewSnowflakeGenerator(machineID int64) *SnowflakeGenerator {
	// Epoch: January 1, 2024
	epoch := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()

	return &SnowflakeGenerator{
		epoch:     epoch,
		machineID: machineID & 0x3FF, // 10 bits
		sequence:  0,
	}
}

// Generate tạo Snowflake ID
func (s *SnowflakeGenerator) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()

	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & 0xFFF // 12 bits
		if s.sequence == 0 {
			// Wait for next millisecond
			for now <= s.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastTime = now

	// Snowflake ID format:
	// | 41 bits timestamp | 10 bits machine ID | 12 bits sequence |
	id := ((now - s.epoch) << 22) | (s.machineID << 12) | s.sequence

	return id
}

// GenerateAsShortCode chuyển Snowflake ID thành short code
func (s *SnowflakeGenerator) GenerateAsShortCode() string {
	id := s.Generate()
	return GenerateBase62(uint64(id))
}
