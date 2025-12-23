package littlecache

import (
	"sync"
	"testing"
	"time"
)

func TestTTLCache_BasicOperations(t *testing.T) {
	config := Config{MaxSize: 10, EvictionPolicy: LRU}
	ttlCache, err := NewTTLCacheFromConfig(config, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create TTL cache: %v", err)
	}
	defer ttlCache.Stop()

	// Test Set and Get
	ttlCache.Set("key1", "value1")
	value, exists := ttlCache.Get("key1")
	if !exists {
		t.Errorf("Expected key1 to exist")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test Size
	if ttlCache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", ttlCache.Size())
	}

	// Test Get non-existent key
	_, exists = ttlCache.Get("nonexistent")
	if exists {
		t.Errorf("Expected nonexistent key to not exist")
	}
}

func TestTTLCache_Expiration(t *testing.T) {
	config := Config{MaxSize: 10, EvictionPolicy: LRU}
	ttlCache, err := NewTTLCacheFromConfig(config, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create TTL cache: %v", err)
	}
	defer ttlCache.Stop()

	// Set a value with short TTL
	ttlCache.Set("key1", "value1")

	// Should exist immediately
	value, exists := ttlCache.Get("key1")
	if !exists || value != "value1" {
		t.Errorf("Expected key1 to exist immediately after set")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not exist after expiration
	_, exists = ttlCache.Get("key1")
	if exists {
		t.Errorf("Expected key1 to be expired")
	}

	// Size should be 0 after expiration
	if ttlCache.Size() != 0 {
		t.Errorf("Expected size 0 after expiration, got %d", ttlCache.Size())
	}
}

func TestTTLCache_SetWithTTL(t *testing.T) {
	config := Config{MaxSize: 10, EvictionPolicy: LRU}
	ttlCache, err := NewTTLCacheFromConfig(config, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create TTL cache: %v", err)
	}
	defer ttlCache.Stop()

	// Set with custom TTL
	ttlCache.SetWithTTL("key1", "value1", 100*time.Millisecond)

	// Should exist immediately
	value, exists := ttlCache.Get("key1")
	if !exists || value != "value1" {
		t.Errorf("Expected key1 to exist immediately")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, exists = ttlCache.Get("key1")
	if exists {
		t.Errorf("Expected key1 to be expired")
	}
}

func TestTTLCache_GetTTL(t *testing.T) {
	config := Config{MaxSize: 10, EvictionPolicy: LRU}
	ttlCache, err := NewTTLCacheFromConfig(config, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to create TTL cache: %v", err)
	}
	defer ttlCache.Stop()

	// Set a value
	ttlCache.Set("key1", "value1")

	// Get TTL immediately
	ttl, exists := ttlCache.GetTTL("key1")
	if !exists {
		t.Errorf("Expected TTL to exist for key1")
	}

	// Should be close to 1 second (allowing for some processing time)
	if ttl < 900*time.Millisecond || ttl > 1*time.Second {
		t.Errorf("Expected TTL around 1 second, got %v", ttl)
	}

	// Wait a bit
	time.Sleep(200 * time.Millisecond)

	// Get TTL again
	ttl, exists = ttlCache.GetTTL("key1")
	if !exists {
		t.Errorf("Expected TTL to still exist for key1")
	}

	// Should be less than before
	if ttl > 900*time.Millisecond {
		t.Errorf("Expected TTL to decrease, got %v", ttl)
	}

	// Test non-existent key
	_, exists = ttlCache.GetTTL("nonexistent")
	if exists {
		t.Errorf("Expected no TTL for nonexistent key")
	}
}

func TestTTLCache_ExtendTTL(t *testing.T) {
	config := Config{MaxSize: 10, EvictionPolicy: LRU}
	ttlCache, err := NewTTLCacheFromConfig(config, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create TTL cache: %v", err)
	}
	defer ttlCache.Stop()

	// Set a value
	ttlCache.Set("key1", "value1")

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Extend TTL
	success := ttlCache.ExtendTTL("key1", 300*time.Millisecond)
	if !success {
		t.Errorf("Expected ExtendTTL to succeed")
	}

	// Wait beyond original expiration time
	time.Sleep(150 * time.Millisecond)

	// Should still exist because we extended it
	_, exists := ttlCache.Get("key1")
	if !exists {
		t.Errorf("Expected key1 to still exist after TTL extension")
	}

	// Test extending non-existent key
	success = ttlCache.ExtendTTL("nonexistent", 1*time.Second)
	if success {
		t.Errorf("Expected ExtendTTL to fail for nonexistent key")
	}
}

func TestTTLCache_Delete(t *testing.T) {
	config := Config{MaxSize: 10, EvictionPolicy: LRU}
	ttlCache, err := NewTTLCacheFromConfig(config, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create TTL cache: %v", err)
	}
	defer ttlCache.Stop()

	ttlCache.Set("key1", "value1")
	ttlCache.Set("key2", "value2")

	if ttlCache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", ttlCache.Size())
	}

	ttlCache.Delete("key1")
	if ttlCache.Size() != 1 {
		t.Errorf("Expected size 1 after delete, got %d", ttlCache.Size())
	}

	_, exists := ttlCache.Get("key1")
	if exists {
		t.Errorf("Expected key1 to be deleted")
	}

	// TTL should also be removed
	_, exists = ttlCache.GetTTL("key1")
	if exists {
		t.Errorf("Expected TTL for key1 to be removed")
	}
}

func TestTTLCache_Clear(t *testing.T) {
	config := Config{MaxSize: 10, EvictionPolicy: LRU}
	ttlCache, err := NewTTLCacheFromConfig(config, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create TTL cache: %v", err)
	}
	defer ttlCache.Stop()

	ttlCache.Set("key1", "value1")
	ttlCache.Set("key2", "value2")
	ttlCache.Set("key3", "value3")

	ttlCache.Clear()
	if ttlCache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", ttlCache.Size())
	}

	// All keys should be gone
	_, exists := ttlCache.Get("key1")
	if exists {
		t.Errorf("Expected key1 to be cleared")
	}

	// TTLs should also be cleared
	_, exists = ttlCache.GetTTL("key1")
	if exists {
		t.Errorf("Expected TTL for key1 to be cleared")
	}
}

func TestTTLCache_AutomaticCleanup(t *testing.T) {
	// Create config with fast cleanup interval
	config := Config{MaxSize: 10, EvictionPolicy: LRU}
	underlyingCache, err := NewLittleCache(config)
	if err != nil {
		t.Fatalf("Failed to create underlying cache: %v", err)
	}

	ttlConfig := TTLConfig{
		UnderlyingCache: underlyingCache,
		DefaultTTL:      100 * time.Millisecond,
		CleanupInterval: 50 * time.Millisecond, // Fast cleanup for testing
	}

	ttlCache := NewTTLCache(ttlConfig)
	defer ttlCache.Stop()

	// Set multiple values
	ttlCache.Set("key1", "value1")
	ttlCache.Set("key2", "value2")
	ttlCache.Set("key3", "value3")

	// Initially should have 3 items
	if ttlCache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", ttlCache.Size())
	}

	// Wait for expiration and cleanup
	time.Sleep(200 * time.Millisecond)

	// After cleanup, should have 0 items
	if ttlCache.Size() != 0 {
		t.Errorf("Expected size 0 after cleanup, got %d", ttlCache.Size())
	}

	// Keys should not be accessible
	_, exists := ttlCache.Get("key1")
	if exists {
		t.Errorf("Expected key1 to be cleaned up")
	}
}

func TestTTLCache_Resize(t *testing.T) {
	config := Config{MaxSize: 10, EvictionPolicy: LRU}
	ttlCache, err := NewTTLCacheFromConfig(config, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create TTL cache: %v", err)
	}
	defer ttlCache.Stop()

	// Add a single item
	ttlCache.Set("key1", "value1")

	// Resize to smaller capacity
	err = ttlCache.Resize(3)
	if err != nil {
		t.Errorf("Unexpected error during resize: %v", err)
	}

	// The item should still be accessible
	_, exists := ttlCache.Get("key1")
	if !exists {
		t.Errorf("Expected key1 to still exist after resize")
	}

	// Test invalid resize
	err = ttlCache.Resize(0)
	if err == nil {
		t.Errorf("Expected error for invalid resize")
	}
}

func TestTTLCache_Concurrency(t *testing.T) {
	config := Config{MaxSize: 100, EvictionPolicy: LRU}
	ttlCache, err := NewTTLCacheFromConfig(config, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to create TTL cache: %v", err)
	}
	defer ttlCache.Stop()

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "key_" + string(rune('0'+goroutineID)) + "_" + string(rune('0'+j%10))
				value := "value_" + string(rune('0'+goroutineID)) + "_" + string(rune('0'+j%10))
				ttlCache.Set(key, value)
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "key_" + string(rune('0'+goroutineID)) + "_" + string(rune('0'+j%10))
				ttlCache.Get(key)
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "key_" + string(rune('0'+goroutineID)) + "_" + string(rune('0'+j%10))
				ttlCache.GetTTL(key)
				ttlCache.ExtendTTL(key, 1*time.Second)
			}
		}(i)
	}

	wg.Wait()

	size := ttlCache.Size()
	if size > 100 {
		t.Errorf("Cache size exceeded capacity: %d", size)
	}
}

func TestTTLCache_MixedTTLs(t *testing.T) {
	config := Config{MaxSize: 10, EvictionPolicy: LRU}
	ttlCache, err := NewTTLCacheFromConfig(config, 1*time.Second) // Default TTL
	if err != nil {
		t.Fatalf("Failed to create TTL cache: %v", err)
	}
	defer ttlCache.Stop()

	// Set items with different TTLs
	ttlCache.Set("short", "value1")                                  // Uses default TTL (1 second)
	ttlCache.SetWithTTL("long", "value2", 5*time.Second)             // Custom long TTL
	ttlCache.SetWithTTL("veryshort", "value3", 100*time.Millisecond) // Custom short TTL

	// Initially all should exist
	if ttlCache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", ttlCache.Size())
	}

	// Wait for very short to expire
	time.Sleep(150 * time.Millisecond)

	// Very short should be gone, others should remain
	_, exists := ttlCache.Get("veryshort")
	if exists {
		t.Errorf("Expected veryshort to be expired")
	}

	_, exists = ttlCache.Get("short")
	if !exists {
		t.Errorf("Expected short to still exist")
	}

	_, exists = ttlCache.Get("long")
	if !exists {
		t.Errorf("Expected long to still exist")
	}

	// Wait for short to expire
	time.Sleep(1 * time.Second)

	// Short should be gone, long should remain
	_, exists = ttlCache.Get("short")
	if exists {
		t.Errorf("Expected short to be expired")
	}

	_, exists = ttlCache.Get("long")
	if !exists {
		t.Errorf("Expected long to still exist")
	}
}
