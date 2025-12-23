# LittleCache

A simple and fast, thread-safe cache library for Go with support for multiple eviction policies.

## Features

- **Thread-safe**: Safe for concurrent use with multiple goroutines
- **LRU Eviction**: Least Recently Used eviction policy
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
    "littlecache"
)

func main() {
    // Create a new cache with default configuration (LRU, capacity: 1024)
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
config := littlecache.Config{
    MaxSize:        100,
    EvictionPolicy: littlecache.LRU,
}

cache, err := littlecache.NewLittleCache(config)
if err != nil {
    panic(err)
}
```

### Dynamic Resizing

```go
// Resize cache to new capacity
err := cache.Resize(200)
if err != nil {
    panic(err)
}
```

## API Reference

### Interface Methods

- `Set(key string, value interface{})` - Add or update a key-value pair
- `Get(key string) (interface{}, bool)` - Retrieve a value by key
- `Delete(key string)` - Remove a key-value pair
- `Clear()` - Remove all key-value pairs
- `Size() int` - Get the number of items in cache
- `Resize(newSize int) error` - Change cache capacity

### Configuration

```go
type Config struct {
    MaxSize        int            // Maximum number of items
    EvictionPolicy EvictionPolicy // Eviction policy (LRU, LFU, NoEviction)
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