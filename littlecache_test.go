package littlecache

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config.MaxSize != 2048 {
		t.Errorf("Expected MaxSize 2048, got %d", config.MaxSize)
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

	t.Run("create LFU cache", func(t *testing.T) {
		config := Config{MaxSize: 10, EvictionPolicy: LFU}
		cache, err := NewLittleCache(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if _, ok := cache.(*LFUCache); !ok {
			t.Errorf("Expected LFUCache type")
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
