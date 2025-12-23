package littlecache

import (
	"sync"
)

type DefCache struct {
	config Config
	data   map[string]interface{}
	mu     sync.RWMutex
}

func NewDefCache(config Config) (*DefCache, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &DefCache{
		config: config,
		data:   make(map[string]interface{}),
	}, nil
}

func (d *DefCache) Set(key string, value interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.data[key]; exists {
		d.data[key] = value
		return
	}

	if len(d.data) >= d.config.MaxSize {
		return
	}

	d.data[key] = value
}

func (d *DefCache) Get(key string) (interface{}, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	value, exists := d.data[key]
	return value, exists
}

func (d *DefCache) Delete(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.data, key)
}

func (d *DefCache) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.data = make(map[string]interface{})
}

func (d *DefCache) Size() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.data)
}

func (d *DefCache) Resize(newSize int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if newSize <= 0 {
		return ErrInvalidMaxSize
	}

	d.config.MaxSize = newSize

	// If new size is smaller than current data, we keep all data
	// since NoEviction policy doesn't remove items
	return nil
}
