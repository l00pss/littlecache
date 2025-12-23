package littlecache

import (
	"errors"
	"time"
)

type LittleCacheError error

var (
	// ErrInvalidMaxSize is returned when the MaxSize in the config is invalid.
	ErrInvalidMaxSize = errors.New("invalid MaxSize: must be greater than 0")
	// ErrInvalidEvictionPolicy is returned when the EvictionPolicy in the config is invalid.
	ErrInvalidEvictionPolicy = errors.New("invalid EvictionPolicy")
)

type EvictionPolicy int

const (
	// NoEviction indicates that no eviction policy is applied.
	NoEviction EvictionPolicy = iota
	// LRU indicates that the Least Recently Used eviction policy is applied.
	LRU
	// LFU indicates that the Least Frequently Used eviction policy is applied.
	LFU
	// TTL indicates that the Time To Live eviction policy is applied.
	TTL
)

type LittleCache interface {
	// Set adds a key-value pair to the cache.
	Set(key string, value interface{})
	// Get retrieves a value from the cache by key.
	Get(key string) (interface{}, bool)
	// Delete removes a key-value pair from the cache by key.
	Delete(key string)
	// Clear removes all key-value pairs from the cache.
	Clear()
	// Size returns the number of key-value pairs in the cache.
	Size() int
	// Resize changes the capacity of the cache.
	Resize(newSize int) error
}

type Config struct {
	// MaxSize defines the maximum number of items the cache can hold.
	MaxSize int
	// EvictionPolicy defines the eviction policy to use when the cache is full.
	EvictionPolicy EvictionPolicy
}

func DefaultConfig() Config {
	return Config{
		MaxSize:        2048,
		EvictionPolicy: LRU,
	}
}

func (c *Config) Validate() error {
	if c.MaxSize <= 0 {
		return ErrInvalidMaxSize
	}
	if c.EvictionPolicy < NoEviction || c.EvictionPolicy > TTL {
		return ErrInvalidEvictionPolicy
	}
	return nil
}

func NewLittleCache(config Config) (LittleCache, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	switch config.EvictionPolicy {
	case NoEviction:
		return NewDefCache(config)
	case LRU:
		return NewLRUCache(config)
	case LFU:
		return NewLFUCache(config)
	case TTL:
		return NewTTLCacheFromConfig(config, time.Duration(5*time.Minute))
	default:
		return nil, ErrInvalidEvictionPolicy
	}
}
