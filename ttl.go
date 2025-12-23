package littlecache

import (
	"sync"
	"time"
)

type TTLEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

func (e *TTLEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

type TTLCache struct {
	cache        LittleCache
	ttlEntries   map[string]*TTLEntry
	defaultTTL   time.Duration
	cleanupTimer *time.Timer
	mu           sync.RWMutex
	stopCleanup  chan bool
}

type TTLConfig struct {
	UnderlyingCache LittleCache
	DefaultTTL      time.Duration
	CleanupInterval time.Duration
}

func NewTTLCache(config TTLConfig) *TTLCache {
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 5 * time.Minute // default 5 minutes
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 1 * time.Minute // cleanup every minute
	}

	ttlCache := &TTLCache{
		cache:       config.UnderlyingCache,
		ttlEntries:  make(map[string]*TTLEntry),
		defaultTTL:  config.DefaultTTL,
		stopCleanup: make(chan bool, 1),
	}

	ttlCache.startCleanup(config.CleanupInterval)

	return ttlCache
}

func NewTTLCacheFromConfig(config Config, defaultTTL time.Duration) (*TTLCache, error) {
	underlyingCache, err := NewLittleCache(config)
	if err != nil {
		return nil, err
	}

	ttlConfig := TTLConfig{
		UnderlyingCache: underlyingCache,
		DefaultTTL:      defaultTTL,
		CleanupInterval: 1 * time.Minute,
	}

	return NewTTLCache(ttlConfig), nil
}

func (t *TTLCache) Set(key string, value interface{}) {
	t.SetWithTTL(key, value, t.defaultTTL)
}

func (t *TTLCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	expiresAt := time.Now().Add(ttl)
	ttlEntry := &TTLEntry{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	t.ttlEntries[key] = ttlEntry
	t.cache.Set(key, value)
}

func (t *TTLCache) Get(key string) (interface{}, bool) {
	t.mu.RLock()
	ttlEntry, exists := t.ttlEntries[key]
	if !exists {
		t.mu.RUnlock()
		return nil, false
	}

	if ttlEntry.IsExpired() {
		t.mu.RUnlock()
		t.Delete(key)
		return nil, false
	}
	t.mu.RUnlock()

	return t.cache.Get(key)
}

func (t *TTLCache) Delete(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.ttlEntries, key)
	t.cache.Delete(key)
}

func (t *TTLCache) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.ttlEntries = make(map[string]*TTLEntry)
	t.cache.Clear()
}

func (t *TTLCache) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	count := 0
	for _, entry := range t.ttlEntries {
		if !entry.IsExpired() {
			count++
		}
	}
	return count
}

func (t *TTLCache) Resize(newSize int) error {
	return t.cache.Resize(newSize)
}

func (t *TTLCache) GetTTL(key string) (time.Duration, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	entry, exists := t.ttlEntries[key]
	if !exists {
		return 0, false
	}

	if entry.IsExpired() {
		return 0, false
	}

	remaining := time.Until(entry.ExpiresAt)
	return remaining, true
}

func (t *TTLCache) ExtendTTL(key string, additionalTime time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	entry, exists := t.ttlEntries[key]
	if !exists || entry.IsExpired() {
		return false
	}

	entry.ExpiresAt = entry.ExpiresAt.Add(additionalTime)
	return true
}

func (t *TTLCache) startCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.cleanup()
			case <-t.stopCleanup:
				return
			}
		}
	}()
}

func (t *TTLCache) cleanup() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, entry := range t.ttlEntries {
		if now.After(entry.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(t.ttlEntries, key)
		t.cache.Delete(key)
	}
}

func (t *TTLCache) Stop() {
	select {
	case t.stopCleanup <- true:
	default:
	}
}
