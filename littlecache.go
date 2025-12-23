package littlecache

import (
	"errors"
	"sync"
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
		MaxSize:        1024,
		EvictionPolicy: LRU,
	}
}

func (c *Config) Validate() error {
	if c.MaxSize <= 0 {
		return ErrInvalidMaxSize
	}
	if c.EvictionPolicy < NoEviction || c.EvictionPolicy > LFU {
		return ErrInvalidEvictionPolicy
	}
	return nil
}

func NewLittleCache(config Config) (LittleCache, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	switch config.EvictionPolicy {
	case LRU:
		return NewLRUCache(config)
	case LFU:
		return NewLFUCache(config)
	default:
		return nil, ErrInvalidEvictionPolicy
	}
}

type LRUNode struct {
	key   string
	value interface{}
	prev  *LRUNode
	next  *LRUNode
}

type LRUCache struct {
	config   Config
	capacity int
	size     int
	cache    map[string]*LRUNode
	head     *LRUNode
	tail     *LRUNode
	mu       sync.RWMutex
}

func NewLRUCache(config Config) (*LRUCache, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	head := &LRUNode{}
	tail := &LRUNode{}
	head.next = tail
	tail.prev = head

	return &LRUCache{
		config:   config,
		capacity: config.MaxSize,
		size:     0,
		cache:    make(map[string]*LRUNode),
		head:     head,
		tail:     tail,
	}, nil
}

func (lru *LRUCache) addNode(node *LRUNode) {
	node.prev = lru.head
	node.next = lru.head.next
	lru.head.next.prev = node
	lru.head.next = node
}

func (lru *LRUCache) removeNode(node *LRUNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (lru *LRUCache) moveToHead(node *LRUNode) {
	lru.removeNode(node)
	lru.addNode(node)
}

func (lru *LRUCache) popTail() *LRUNode {
	lastNode := lru.tail.prev
	lru.removeNode(lastNode)
	return lastNode
}

func (lru *LRUCache) Set(key string, value interface{}) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	node, exists := lru.cache[key]

	if !exists {
		newNode := &LRUNode{key: key, value: value}
		lru.cache[key] = newNode
		lru.addNode(newNode)
		lru.size++

		if lru.size > lru.capacity {
			tail := lru.popTail()
			delete(lru.cache, tail.key)
			lru.size--
		}
	} else {
		node.value = value
		lru.moveToHead(node)
	}
}

func (lru *LRUCache) Get(key string) (interface{}, bool) {
	lru.mu.RLock()
	defer lru.mu.RUnlock()

	if node, exists := lru.cache[key]; exists {
		lru.mu.RUnlock()
		lru.mu.Lock()
		lru.moveToHead(node)
		lru.mu.Unlock()
		lru.mu.RLock()
		return node.value, true
	}
	return nil, false
}

func (lru *LRUCache) Delete(key string) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if node, exists := lru.cache[key]; exists {
		lru.removeNode(node)
		delete(lru.cache, key)
		lru.size--
	}
}

func (lru *LRUCache) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.cache = make(map[string]*LRUNode)
	lru.size = 0
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
}

func (lru *LRUCache) Size() int {
	lru.mu.RLock()
	defer lru.mu.RUnlock()
	return lru.size
}

func (lru *LRUCache) Resize(newSize int) error {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if newSize <= 0 {
		return ErrInvalidMaxSize
	}

	lru.capacity = newSize
	for lru.size > lru.capacity {
		tail := lru.popTail()
		delete(lru.cache, tail.key)
		lru.size--
	}
	return nil
}

// LFUCache represents a Least Frequently Used cache.
type LFUCache struct {
	LittleCache
	config Config
}

func NewLFUCache(config Config) (*LFUCache, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &LFUCache{config: config}, nil
}

func (lfu *LFUCache) Set(key string, value interface{}) {
	// Placeholder implementation
}

func (lfu *LFUCache) Get(key string) (interface{}, bool) {
	// Placeholder implementation
	return nil, false
}

func (lfu *LFUCache) Delete(key string) {
	// Placeholder implementation
}

func (lfu *LFUCache) Clear() {
	// Placeholder implementation
}

func (lfu *LFUCache) Size() int {
	// Placeholder implementation
	return 0
}

func (lfu *LFUCache) Resize(newSize int) error {
	return errors.New("Resize is not implemented for LFUCache")
}
