package littlecache

import (
	"strconv"
	"sync"
	"testing"
)

func TestLFUCache_BasicOperations(t *testing.T) {
	config := Config{MaxSize: 3, EvictionPolicy: LFU}
	cache, err := NewLFUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LFU cache: %v", err)
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

func TestLFUCache_Eviction(t *testing.T) {
	config := Config{MaxSize: 3, EvictionPolicy: LFU}
	cache, err := NewLFUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LFU cache: %v", err)
	}

	// Fill cache to capacity
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}

	// Access key1 multiple times to increase its frequency
	cache.Get("key1")
	cache.Get("key1")
	cache.Get("key1")

	// Access key2 once
	cache.Get("key2")

	// key3 has frequency 1 (only set, never accessed)
	// key2 has frequency 2 (set + 1 access)
	// key1 has frequency 4 (set + 3 accesses)

	// Add new key, should evict key3 (least frequently used)
	cache.Set("key4", "value4")
	if cache.Size() != 3 {
		t.Errorf("Expected size to remain 3, got %d", cache.Size())
	}

	// key3 should be evicted
	_, exists := cache.Get("key3")
	if exists {
		t.Errorf("Expected key3 to be evicted")
	}

	// key1, key2, and key4 should exist
	_, exists = cache.Get("key1")
	if !exists {
		t.Errorf("Expected key1 to exist")
	}
	_, exists = cache.Get("key2")
	if !exists {
		t.Errorf("Expected key2 to exist")
	}
	_, exists = cache.Get("key4")
	if !exists {
		t.Errorf("Expected key4 to exist")
	}
}

func TestLFUCache_LFUOrder(t *testing.T) {
	config := Config{MaxSize: 3, EvictionPolicy: LFU}
	cache, err := NewLFUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LFU cache: %v", err)
	}

	// Add items
	cache.Set("first", 1)
	cache.Set("second", 2)
	cache.Set("third", 3)

	// Access first item multiple times
	cache.Get("first")
	cache.Get("first")

	// Access second item once
	cache.Get("second")

	// third has frequency 1
	// second has frequency 2
	// first has frequency 3

	// Add fourth item, should evict "third" (least frequently used)
	cache.Set("fourth", 4)

	// Check that "third" was evicted
	_, exists := cache.Get("third")
	if exists {
		t.Errorf("Expected 'third' to be evicted")
	}

	// Check that others exist
	_, exists = cache.Get("first")
	if !exists {
		t.Errorf("Expected 'first' to exist")
	}
	_, exists = cache.Get("second")
	if !exists {
		t.Errorf("Expected 'second' to exist")
	}
	_, exists = cache.Get("fourth")
	if !exists {
		t.Errorf("Expected 'fourth' to exist")
	}
}

func TestLFUCache_Delete(t *testing.T) {
	config := Config{MaxSize: 3, EvictionPolicy: LFU}
	cache, err := NewLFUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LFU cache: %v", err)
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

func TestLFUCache_Clear(t *testing.T) {
	config := Config{MaxSize: 3, EvictionPolicy: LFU}
	cache, err := NewLFUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LFU cache: %v", err)
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

func TestLFUCache_Resize(t *testing.T) {
	config := Config{MaxSize: 4, EvictionPolicy: LFU}
	cache, err := NewLFUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LFU cache: %v", err)
	}

	// Fill cache and create frequency differences
	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)
	cache.Set("d", 4)

	// Create frequency differences
	cache.Get("a") // freq: 2
	cache.Get("a") // freq: 3
	cache.Get("b") // freq: 2
	// c and d have freq: 1

	// Resize to smaller capacity (should keep highest frequency items)
	err = cache.Resize(2)
	if err != nil {
		t.Errorf("Unexpected error during resize: %v", err)
	}

	if cache.Size() != 2 {
		t.Errorf("Expected size 2 after resize, got %d", cache.Size())
	}

	// 'a' should remain (highest frequency: 3)
	_, exists := cache.Get("a")
	if !exists {
		t.Errorf("Expected 'a' to remain")
	}

	// 'b' should remain (frequency: 2)
	_, exists = cache.Get("b")
	if !exists {
		t.Errorf("Expected 'b' to remain")
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

func TestLFUCache_Concurrency(t *testing.T) {
	config := Config{MaxSize: 100, EvictionPolicy: LFU}
	cache, err := NewLFUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LFU cache: %v", err)
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

func TestLFUCache_FrequencyTracking(t *testing.T) {
	config := Config{MaxSize: 2, EvictionPolicy: LFU}
	cache, err := NewLFUCache(config)
	if err != nil {
		t.Fatalf("Failed to create LFU cache: %v", err)
	}

	// Add two items
	cache.Set("item1", "value1")
	cache.Set("item2", "value2")

	// Access item1 multiple times
	cache.Get("item1")
	cache.Get("item1")
	cache.Get("item1")

	// item1 frequency: 4 (1 set + 3 gets)
	// item2 frequency: 1 (1 set)

	// Add third item, should evict item2
	cache.Set("item3", "value3")

	// item2 should be evicted
	_, exists := cache.Get("item2")
	if exists {
		t.Errorf("Expected item2 to be evicted due to low frequency")
	}

	// item1 and item3 should exist
	_, exists = cache.Get("item1")
	if !exists {
		t.Errorf("Expected item1 to exist")
	}
	_, exists = cache.Get("item3")
	if !exists {
		t.Errorf("Expected item3 to exist")
	}
}
