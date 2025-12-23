# LittleCache

A simple and fast, thread-safe cache library for Go with support for multiple eviction policies.

## Features

- **Thread-safe**: Safe for concurrent use with multiple goroutines
- **No Eviction**: Simple cache that doesn't evict items when full
- **LRU Eviction**: Least Recently Used eviction policy
- **LFU Eviction**: Least Frequently Used eviction policy
- **TTL Support**: Time-To-Live expiration with automatic cleanup
- **Dynamic Resizing**: Change cache capacity at runtime
- **Simple API**: Easy to use interface
- **Zero Dependencies**: No external dependencies

## Installation

```bash
go get github.com/l00pss/littlecache
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "time"
    "github.com/l00pss/littlecache"
)

func main() {
    // Create a new cache with default configuration (LRU, capacity: 2048)
    config := littlecache.DefaultConfig()
    cache, err := littlecache.NewLittleCache(config)
    if err != nil {
        panic(err)
    }

    // Set values
    cache.Set("key1", "value1")
    cache.Set("key2", 42)

    // Get values
    value, exists := cache.Get("key1")
    if exists {
        fmt.Println("Found:", value)
    }

    // Check cache size
    fmt.Println("Cache size:", cache.Size())
}
```

### Custom Configuration

```go
// NoEviction Cache (doesn't evict items when full)
config := littlecache.Config{
    MaxSize:        100,
    EvictionPolicy: littlecache.NoEviction,
}

cache, err := littlecache.NewLittleCache(config)
if err != nil {
    panic(err)
}

// LRU Cache
lru_config := littlecache.Config{
    MaxSize:        100,
    EvictionPolicy: littlecache.LRU,
}

lru_cache, err := littlecache.NewLittleCache(lru_config)
if err != nil {
    panic(err)
}

// LFU Cache
lfu_config := littlecache.Config{
    MaxSize:        100,
    EvictionPolicy: littlecache.LFU,
}

lfu_cache, err := littlecache.NewLittleCache(lfu_config)
if err != nil {
    panic(err)
}
```

### TTL (Time-To-Live) Cache

```go
// Create TTL cache with 5 minute default TTL
config := littlecache.Config{
    MaxSize:        100,
    EvictionPolicy: littlecache.LRU,
}

ttlCache, err := littlecache.NewTTLCacheFromConfig(config, 5*time.Minute)
if err != nil {
    panic(err)
}
defer ttlCache.Stop() // Stop cleanup goroutine

// Set with default TTL (5 minutes)
ttlCache.Set("key1", "value1")

// Set with custom TTL
ttlCache.SetWithTTL("key2", "value2", 1*time.Minute)

// Check remaining TTL
if remaining, exists := ttlCache.GetTTL("key1"); exists {
    fmt.Printf("Key1 expires in: %v\n", remaining)
}

// Extend TTL
ttlCache.ExtendTTL("key1", 2*time.Minute)
```

### Advanced TTL Configuration

```go
// Custom TTL configuration
underlyingCache, err := littlecache.NewLittleCache(littlecache.Config{
    MaxSize:        100,
    EvictionPolicy: littlecache.LRU,
})
if err != nil {
    panic(err)
}

ttlConfig := littlecache.TTLConfig{
    UnderlyingCache: underlyingCache,
    DefaultTTL:      10 * time.Minute,
    CleanupInterval: 30 * time.Second, // How often to run cleanup
}

ttlCache := littlecache.NewTTLCache(ttlConfig)
defer ttlCache.Stop()
```

### Dynamic Resizing

```go
// Resize cache to new capacity
err := cache.Resize(200)
if err != nil {
    panic(err)
}
```

### Eviction Policies

#### NoEviction
Doesn't evict items when cache reaches capacity. New items are simply not added if the cache is full, but existing items can still be updated.

```go
config := littlecache.Config{
    MaxSize:        3,
    EvictionPolicy: littlecache.NoEviction,
}

cache, err := littlecache.NewLittleCache(config)
if err != nil {
    panic(err)
}

// Fill cache to capacity
cache.Set("key1", "value1")
cache.Set("key2", "value2") 
cache.Set("key3", "value3")

// Try to add 4th item - will be ignored since cache is full
cache.Set("key4", "value4")

fmt.Println("Cache size:", cache.Size()) // Output: 3

// key4 won't be found
if _, found := cache.Get("key4"); !found {
    fmt.Println("key4 not found - cache was full")
}

// But you can still update existing keys
cache.Set("key1", "updated_value1") // This works
```

#### LRU (Least Recently Used)
Evicts the least recently accessed item when cache reaches capacity.

```go
config := littlecache.Config{
    MaxSize:        100,
    EvictionPolicy: littlecache.LRU,
}
```

#### LFU (Least Frequently Used)
Evicts the least frequently accessed item when cache reaches capacity. If multiple items have the same frequency, the oldest one is evicted.

```go
config := littlecache.Config{
    MaxSize:        100,
    EvictionPolicy: littlecache.LFU,
}
```

#### TTL (Time-To-Live)
Automatically expires items after a specified duration. Can be combined with any underlying cache type (LRU or LFU).

```go
// Create TTL cache with LRU as underlying cache
ttlCache, err := littlecache.NewTTLCacheFromConfig(
    littlecache.Config{MaxSize: 100, EvictionPolicy: littlecache.LRU},
    5*time.Minute, // default TTL
)
```

## API Reference

### Basic Cache Interface Methods

- `Set(key string, value interface{})` - Add or update a key-value pair
- `Get(key string) (interface{}, bool)` - Retrieve a value by key
- `Delete(key string)` - Remove a key-value pair
- `Clear()` - Remove all key-value pairs
- `Size() int` - Get the number of items in cache
- `Resize(newSize int) error` - Change cache capacity

### TTL Cache Additional Methods

- `SetWithTTL(key string, value interface{}, ttl time.Duration)` - Set with custom TTL
- `GetTTL(key string) (time.Duration, bool)` - Get remaining time until expiration
- `ExtendTTL(key string, additionalTime time.Duration) bool` - Extend expiration time
- `Stop()` - Stop the cleanup goroutine (important for graceful shutdown)

### Configuration

#### Basic Cache Configuration
```go
type Config struct {
    MaxSize        int            // Maximum number of items
    EvictionPolicy EvictionPolicy // Eviction policy (NoEviction, LRU, LFU)
}
```

**Available Eviction Policies:**
- `NoEviction`: No items are evicted when cache is full
- `LRU`: Least Recently Used eviction
- `LFU`: Least Frequently Used eviction
- `TTL`: Time-To-Live expiration (used with TTL cache wrapper)

#### TTL Cache Configuration
```go
type TTLConfig struct {
    UnderlyingCache LittleCache   // The cache implementation to wrap
    DefaultTTL      time.Duration // Default expiration time for items
    CleanupInterval time.Duration // How often to run expired item cleanup
}

type TTLEntry struct {
    Value     interface{}
    ExpiresAt time.Time
}
```

## Thread Safety

LittleCache is designed for concurrent use. All operations are protected by read-write mutexes, allowing multiple concurrent reads while ensuring exclusive access for writes.

## Testing

Run the test suite:

```bash
go test
go test -v  # verbose output
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.