package littlecache

import (
	"fmt"
	"sync"
	"testing"
)

func TestDefCache_BasicOperations(t *testing.T) {
	config := Config{MaxSize: 3, EvictionPolicy: NoEviction}
	cache, err := NewDefCache(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test Set and Get
	cache.Set("key1", "value1")
	value, found := cache.Get("key1")
	if !found {
		t.Errorf("Expected to find key1")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test Size
	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}
}

func TestDefCache_NoEviction(t *testing.T) {
	config := Config{MaxSize: 2, EvictionPolicy: NoEviction}
	cache, err := NewDefCache(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Fill cache to capacity
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}

	// Try to add one more item - should be ignored due to NoEviction
	cache.Set("key3", "value3")

	if cache.Size() != 2 {
		t.Errorf("Expected size to remain 2, got %d", cache.Size())
	}

	// Check that new item was not added
	_, found := cache.Get("key3")
	if found {
		t.Errorf("Expected key3 not to be found")
	}

	// Check that original items are still there
	value1, found1 := cache.Get("key1")
	value2, found2 := cache.Get("key2")

	if !found1 || value1 != "value1" {
		t.Errorf("Expected to find key1 with value1")
	}
	if !found2 || value2 != "value2" {
		t.Errorf("Expected to find key2 with value2")
	}
}

func TestDefCache_UpdateExistingKey(t *testing.T) {
	config := Config{MaxSize: 2, EvictionPolicy: NoEviction}
	cache, err := NewDefCache(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Fill cache to capacity
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Update existing key - should work even when cache is full
	cache.Set("key1", "newvalue1")

	if cache.Size() != 2 {
		t.Errorf("Expected size to remain 2, got %d", cache.Size())
	}

	value, found := cache.Get("key1")
	if !found || value != "newvalue1" {
		t.Errorf("Expected key1 to be updated to newvalue1, got %v", value)
	}
}

func TestDefCache_Delete(t *testing.T) {
	config := Config{MaxSize: 3, EvictionPolicy: NoEviction}
	cache, err := NewDefCache(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	cache.Delete("key1")

	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}

	_, found := cache.Get("key1")
	if found {
		t.Errorf("Expected key1 to be deleted")
	}

	value, found := cache.Get("key2")
	if !found || value != "value2" {
		t.Errorf("Expected key2 to remain")
	}
}

func TestDefCache_Clear(t *testing.T) {
	config := Config{MaxSize: 3, EvictionPolicy: NoEviction}
	cache, err := NewDefCache(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	if found1 || found2 {
		t.Errorf("Expected all keys to be cleared")
	}
}

func TestDefCache_Resize(t *testing.T) {
	config := Config{MaxSize: 2, EvictionPolicy: NoEviction}
	cache, err := NewDefCache(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Resize to larger capacity
	err = cache.Resize(4)
	if err != nil {
		t.Errorf("Unexpected error during resize: %v", err)
	}

	// Should now be able to add more items
	cache.Set("key3", "value3")
	cache.Set("key4", "value4")

	if cache.Size() != 4 {
		t.Errorf("Expected size 4 after resize, got %d", cache.Size())
	}

	// Test invalid resize
	err = cache.Resize(0)
	if err == nil {
		t.Errorf("Expected error for invalid resize")
	}
}

func TestDefCache_Concurrency(t *testing.T) {
	config := Config{MaxSize: 100, EvictionPolicy: NoEviction}
	cache, err := NewDefCache(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 10

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				cache.Set(key, value)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()
	// Just check that we don't panic during concurrent operations
}
