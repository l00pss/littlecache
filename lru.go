package littlecache

import "sync"

type LRUNode struct {
	key   string
	value interface{}
	prev  *LRUNode
	next  *LRUNode
}

type LRUCache struct {
	config Config
	size   int
	cache  map[string]*LRUNode
	head   *LRUNode
	tail   *LRUNode
	mu     sync.RWMutex
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
		config: config,
		size:   0,
		cache:  make(map[string]*LRUNode),
		head:   head,
		tail:   tail,
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

		if lru.size > lru.config.MaxSize {
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

	lru.config.MaxSize = newSize
	for lru.size > lru.config.MaxSize {
		tail := lru.popTail()
		delete(lru.cache, tail.key)
		lru.size--
	}
	return nil
}
