package generator

import (
	"sync"
	"testing"
)

// TestShortCodeGenerator_Generate tests short code generation
func TestShortCodeGenerator_Generate(t *testing.T) {
	gen := NewShortCodeGenerator(6)

	// Test basic generation
	code := gen.Generate()
	if len(code) != 6 {
		t.Errorf("Expected code length 6, got %d", len(code))
	}

	// Test validity
	if !gen.IsValid(code) {
		t.Errorf("Generated code %s should be valid", code)
	}
}

// TestShortCodeGenerator_Uniqueness tests uniqueness of generated codes
func TestShortCodeGenerator_Uniqueness(t *testing.T) {
	gen := NewShortCodeGenerator(6)
	codes := make(map[string]bool)
	count := 10000

	for i := 0; i < count; i++ {
		code := gen.Generate()
		if codes[code] {
			t.Errorf("Duplicate code generated: %s", code)
		}
		codes[code] = true
	}

	if len(codes) != count {
		t.Errorf("Expected %d unique codes, got %d", count, len(codes))
	}
}

// TestShortCodeGenerator_Concurrent tests concurrent generation
func TestShortCodeGenerator_Concurrent(t *testing.T) {
	gen := NewShortCodeGenerator(6)
	codes := sync.Map{}
	var wg sync.WaitGroup
	goroutines := 10
	codesPerGoroutine := 1000

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < codesPerGoroutine; j++ {
				code := gen.Generate()
				if _, loaded := codes.LoadOrStore(code, true); loaded {
					t.Errorf("Duplicate code in concurrent test: %s", code)
				}
			}
		}()
	}

	wg.Wait()

	count := 0
	codes.Range(func(_, _ interface{}) bool {
		count++
		return true
	})

	expected := goroutines * codesPerGoroutine
	if count != expected {
		t.Errorf("Expected %d unique codes, got %d", expected, count)
	}
}

// TestShortCodeGenerator_IsValid tests validation
func TestShortCodeGenerator_IsValid(t *testing.T) {
	gen := NewShortCodeGenerator(6)

	tests := []struct {
		code    string
		isValid bool
	}{
		{"abc123", true},
		{"ABCDEF", true},
		{"aB3xY9", true},
		{"ab", false},            // Too short
		{"abc", false},           // Too short
		{"abcdefghijklm", false}, // Too long
		{"abc@#$", false},        // Invalid characters
		{"", false},              // Empty
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := gen.IsValid(tt.code)
			if result != tt.isValid {
				t.Errorf("IsValid(%s) = %v, want %v", tt.code, result, tt.isValid)
			}
		})
	}
}

// TestBase62Encoding tests Base62 encoding
func TestBase62Encoding(t *testing.T) {
	tests := []struct {
		id       uint64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{61, "z"},
		{62, "10"},
		{1000, "g8"},
	}

	for _, tt := range tests {
		result := GenerateBase62(tt.id)
		if result != tt.expected {
			t.Errorf("GenerateBase62(%d) = %s, want %s", tt.id, result, tt.expected)
		}
	}
}

// TestSnowflakeGenerator tests Snowflake ID generation
func TestSnowflakeGenerator(t *testing.T) {
	snowflake := NewSnowflakeGenerator(1)

	// Generate multiple IDs
	ids := make(map[int64]bool)
	for i := 0; i < 1000; i++ {
		id := snowflake.Generate()
		if ids[id] {
			t.Errorf("Duplicate Snowflake ID: %d", id)
		}
		ids[id] = true
	}
}

// BenchmarkShortCodeGeneration benchmarks code generation
func BenchmarkShortCodeGeneration(b *testing.B) {
	gen := NewShortCodeGenerator(6)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gen.Generate()
	}
}

// BenchmarkBase62Encoding benchmarks Base62 encoding
func BenchmarkBase62Encoding(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GenerateBase62(uint64(i))
	}
}
