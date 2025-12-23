package littlecache

import (
	"strconv"
	"sync"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config.MaxSize != 1024 {
		t.Errorf("Expected MaxSize 1024, got %d", config.MaxSize)
	}
	if config.EvictionPolicy != LRU {
		t.Errorf("Expected EvictionPolicy LRU, got %d", config.EvictionPolicy)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config",
			config:      Config{MaxSize: 10, EvictionPolicy: LRU},
			expectError: false,
		},
		{
			name:        "invalid MaxSize zero",
			config:      Config{MaxSize: 0, EvictionPolicy: LRU},
			expectError: true,
			errorMsg:    "invalid MaxSize: must be greater than 0",
		},
		{
			name:        "invalid MaxSize negative",
			config:      Config{MaxSize: -1, EvictionPolicy: LRU},
			expectError: true,
			errorMsg:    "invalid MaxSize: must be greater than 0",
		},
		{
			name:        "invalid EvictionPolicy",
			config:      Config{MaxSize: 10, EvictionPolicy: 99},
			expectError: true,
			errorMsg:    "invalid EvictionPolicy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestNewLittleCache(t *testing.T) {
	t.Run("create LRU cache", func(t *testing.T) {
		config := Config{MaxSize: 10, EvictionPolicy: LRU}
		cache, err := NewLittleCache(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if _, ok := cache.(*LRUCache); !ok {
			t.Errorf("Expected LRUCache type")
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		config := Config{MaxSize: 0, EvictionPolicy: LRU}
		_, err := NewLittleCache(config)
		if err == nil {
			t.Errorf("Expected error for invalid config")
		}
	})

	t.Run("no eviction policy", func(t *testing.T) {
		config := Config{MaxSize: 10, EvictionPolicy: NoEviction}
		_, err := NewLittleCache(config)
		if err == nil {
			t.Errorf("Expected error for NoEviction policy")
		}
	})
}

func TestLRUCache_BasicOperations(t *testing.T) {
	config := Config{MaxSize: 3, EvictionPolicy: LRU}
	cache, err := NewLRUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	// Test Set and Get
	cache.Set("key1", "value1")
	value, exists := cache.Get("key1")
	if !exists {
		t.Errorf("Expected key1 to exist")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test Size
	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}

	// Test Get non-existent key
	_, exists = cache.Get("nonexistent")
	if exists {
		t.Errorf("Expected nonexistent key to not exist")
	}

	// Test Update existing key
	cache.Set("key1", "updated_value1")
	value, exists = cache.Get("key1")
	if !exists || value != "updated_value1" {
		t.Errorf("Expected updated_value1, got %v", value)
	}
	if cache.Size() != 1 {
		t.Errorf("Expected size to remain 1 after update, got %d", cache.Size())
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	config := Config{MaxSize: 2, EvictionPolicy: LRU}
	cache, err := NewLRUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	// Fill cache to capacity
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}

	// Access key1 to make it recently used
	cache.Get("key1")

	// Add new key, should evict key2 (least recently used)
	cache.Set("key3", "value3")
	if cache.Size() != 2 {
		t.Errorf("Expected size to remain 2, got %d", cache.Size())
	}

	// key2 should be evicted
	_, exists := cache.Get("key2")
	if exists {
		t.Errorf("Expected key2 to be evicted")
	}

	// key1 and key3 should exist
	_, exists = cache.Get("key1")
	if !exists {
		t.Errorf("Expected key1 to exist")
	}
	_, exists = cache.Get("key3")
	if !exists {
		t.Errorf("Expected key3 to exist")
	}
}

func TestLRUCache_Delete(t *testing.T) {
	config := Config{MaxSize: 3, EvictionPolicy: LRU}
	cache, err := NewLRUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}

	cache.Delete("key1")
	if cache.Size() != 1 {
		t.Errorf("Expected size 1 after delete, got %d", cache.Size())
	}

	_, exists := cache.Get("key1")
	if exists {
		t.Errorf("Expected key1 to be deleted")
	}

	// Delete non-existent key should not affect cache
	cache.Delete("nonexistent")
	if cache.Size() != 1 {
		t.Errorf("Expected size to remain 1, got %d", cache.Size())
	}
}

func TestLRUCache_Clear(t *testing.T) {
	config := Config{MaxSize: 3, EvictionPolicy: LRU}
	cache, err := NewLRUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	cache.Clear()
	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	// All keys should be gone
	_, exists := cache.Get("key1")
	if exists {
		t.Errorf("Expected key1 to be cleared")
	}
	_, exists = cache.Get("key2")
	if exists {
		t.Errorf("Expected key2 to be cleared")
	}
	_, exists = cache.Get("key3")
	if exists {
		t.Errorf("Expected key3 to be cleared")
	}
}

func TestLRUCache_Resize(t *testing.T) {
	config := Config{MaxSize: 4, EvictionPolicy: LRU}
	cache, err := NewLRUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	// Fill cache
	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)
	cache.Set("d", 4)

	// Resize to smaller capacity
	err = cache.Resize(2)
	if err != nil {
		t.Errorf("Unexpected error during resize: %v", err)
	}

	if cache.Size() != 2 {
		t.Errorf("Expected size 2 after resize, got %d", cache.Size())
	}

	// Only most recently used items should remain (c and d)
	_, exists := cache.Get("a")
	if exists {
		t.Errorf("Expected 'a' to be evicted")
	}
	_, exists = cache.Get("b")
	if exists {
		t.Errorf("Expected 'b' to be evicted")
	}
	_, exists = cache.Get("c")
	if !exists {
		t.Errorf("Expected 'c' to remain")
	}
	_, exists = cache.Get("d")
	if !exists {
		t.Errorf("Expected 'd' to remain")
	}

	// Resize to larger capacity
	err = cache.Resize(5)
	if err != nil {
		t.Errorf("Unexpected error during resize: %v", err)
	}

	// Add more items
	cache.Set("e", 5)
	cache.Set("f", 6)
	cache.Set("g", 7)

	if cache.Size() != 5 {
		t.Errorf("Expected size 5, got %d", cache.Size())
	}

	// Test invalid resize
	err = cache.Resize(0)
	if err == nil {
		t.Errorf("Expected error for invalid resize")
	}
	err = cache.Resize(-1)
	if err == nil {
		t.Errorf("Expected error for negative resize")
	}
}

func TestLRUCache_Concurrency(t *testing.T) {
	config := Config{MaxSize: 100, EvictionPolicy: LRU}
	cache, err := NewLRUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 50
	numOperations := 100

	// Concurrent Set operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "key_" + strconv.Itoa(goroutineID) + "_" + strconv.Itoa(j)
				value := "value_" + strconv.Itoa(goroutineID) + "_" + strconv.Itoa(j)
				cache.Set(key, value)
			}
		}(i)
	}

	// Concurrent Get operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "key_" + strconv.Itoa(goroutineID) + "_" + strconv.Itoa(j)
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()

	// Cache size should not exceed capacity
	if cache.Size() > 100 {
		t.Errorf("Cache size exceeded capacity: %d", cache.Size())
	}
}

func TestLRUCache_LRUOrder(t *testing.T) {
	config := Config{MaxSize: 3, EvictionPolicy: LRU}
	cache, err := NewLRUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	// Add items in order
	cache.Set("first", 1)
	cache.Set("second", 2)
	cache.Set("third", 3)

	// Access first item to make it most recently used
	cache.Get("first")

	// Add fourth item, should evict "second" (least recently used)
	cache.Set("fourth", 4)

	// Check that "second" was evicted
	_, exists := cache.Get("second")
	if exists {
		t.Errorf("Expected 'second' to be evicted")
	}

	// Check that "first", "third", and "fourth" exist
	_, exists = cache.Get("first")
	if !exists {
		t.Errorf("Expected 'first' to exist")
	}
	_, exists = cache.Get("third")
	if !exists {
		t.Errorf("Expected 'third' to exist")
	}
	_, exists = cache.Get("fourth")
	if !exists {
		t.Errorf("Expected 'fourth' to exist")
	}
}
