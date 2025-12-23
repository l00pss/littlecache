package littlecache

import (
	"sync"
)

type LFUNode struct {
	key   string
	value interface{}
	freq  int
	prev  *LFUNode
	next  *LFUNode
}

type LFUCache struct {
	config  Config
	size    int
	cache   map[string]*LFUNode
	freqMap map[int]*LFUNode // frequency -> head of doubly linked list
	minFreq int
	mu      sync.RWMutex
}

func NewLFUCache(config Config) (*LFUCache, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &LFUCache{
		config:  config,
		size:    0,
		cache:   make(map[string]*LFUNode),
		freqMap: make(map[int]*LFUNode),
		minFreq: 0,
	}, nil
}

func (lfu *LFUCache) addNode(node *LFUNode, freq int) {
	if lfu.freqMap[freq] == nil {
		lfu.freqMap[freq] = &LFUNode{}
		lfu.freqMap[freq].next = lfu.freqMap[freq]
		lfu.freqMap[freq].prev = lfu.freqMap[freq]
	}

	head := lfu.freqMap[freq]
	node.next = head.next
	node.prev = head
	head.next.prev = node
	head.next = node
}

func (lfu *LFUCache) removeNode(node *LFUNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (lfu *LFUCache) updateFreq(node *LFUNode) {
	freq := node.freq
	lfu.removeNode(node)

	if lfu.freqMap[freq].next == lfu.freqMap[freq] {
		delete(lfu.freqMap, freq)
		if lfu.minFreq == freq {
			lfu.minFreq++
		}
	}

	node.freq++
	lfu.addNode(node, node.freq)
}

func (lfu *LFUCache) removeLFU() *LFUNode {
	head := lfu.freqMap[lfu.minFreq]
	lastNode := head.prev
	lfu.removeNode(lastNode)

	if head.next == head {
		delete(lfu.freqMap, lfu.minFreq)
	}

	return lastNode
}

func (lfu *LFUCache) Set(key string, value interface{}) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	node, exists := lfu.cache[key]

	if !exists {
		newNode := &LFUNode{key: key, value: value, freq: 1}
		lfu.cache[key] = newNode
		lfu.addNode(newNode, 1)
		lfu.size++
		lfu.minFreq = 1

		if lfu.size > lfu.config.MaxSize {
			lru := lfu.removeLFU()
			delete(lfu.cache, lru.key)
			lfu.size--
		}
	} else {
		node.value = value
		lfu.updateFreq(node)
	}
}

func (lfu *LFUCache) Get(key string) (interface{}, bool) {
	lfu.mu.RLock()
	node, exists := lfu.cache[key]
	if !exists {
		lfu.mu.RUnlock()
		return nil, false
	}

	value := node.value
	lfu.mu.RUnlock()

	lfu.mu.Lock()
	lfu.updateFreq(node)
	lfu.mu.Unlock()

	return value, true
}

func (lfu *LFUCache) Delete(key string) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	node, exists := lfu.cache[key]
	if !exists {
		return
	}

	lfu.removeNode(node)
	delete(lfu.cache, key)
	lfu.size--

	if lfu.freqMap[node.freq].next == lfu.freqMap[node.freq] {
		delete(lfu.freqMap, node.freq)
		if lfu.minFreq == node.freq && lfu.size > 0 {
			lfu.minFreq = 1
			for freq := range lfu.freqMap {
				if freq < lfu.minFreq {
					lfu.minFreq = freq
				}
			}
		}
	}
}

func (lfu *LFUCache) Clear() {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	lfu.cache = make(map[string]*LFUNode)
	lfu.freqMap = make(map[int]*LFUNode)
	lfu.size = 0
	lfu.minFreq = 0
}

func (lfu *LFUCache) Size() int {
	lfu.mu.RLock()
	defer lfu.mu.RUnlock()
	return lfu.size
}

func (lfu *LFUCache) Resize(newSize int) error {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if newSize <= 0 {
		return ErrInvalidMaxSize
	}

	lfu.config.MaxSize = newSize
	for lfu.size > lfu.config.MaxSize {
		lru := lfu.removeLFU()
		delete(lfu.cache, lru.key)
		lfu.size--
	}
	return nil
}
