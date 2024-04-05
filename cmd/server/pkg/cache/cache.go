package cache

import (
	"container/list"
	"sync"
)

type Entry struct {
	key   string
	value interface{}
}

type Cache interface {
	Get(key string) (entry interface{}, exists bool)
	Set(key string, value interface{})
	Delete(key string)
}

type cache struct {
	cache    map[string]*list.Element
	mutexes  sync.Map
	lruList  *list.List
	capacity int
}

// NewCache creates a new cache instance with the specified capacity.
func NewCache(capacity int) Cache {
	return &cache{
		make(map[string]*list.Element),
		sync.Map{},
		list.New(),
		capacity,
	}
}

func (c *cache) Get(key string) (entry interface{}, ok bool) {
	// Acquire a lock for the key
	mutex, _ := c.mutexes.LoadOrStore(key, &sync.Mutex{})
	mutex.(*sync.Mutex).Lock()
	defer mutex.(*sync.Mutex).Unlock()

	var element *list.Element
	if element, ok = c.cache[key]; ok {
		c.lruList.MoveToFront(element)
		return element.Value.(*Entry).value, true
	}
	return nil, false
}

func (c *cache) Set(key string, value interface{}) {
	// Acquire a lock for the key
	mutex, _ := c.mutexes.LoadOrStore(key, &sync.Mutex{})
	mutex.(*sync.Mutex).Lock()
	defer mutex.(*sync.Mutex).Unlock()

	if elem, ok := c.cache[key]; ok {
		// Update existing entry
		elem.Value = value
		c.lruList.MoveToFront(elem) // Move to front (most recently used)
	} else {
		// Add new entry
		if len(c.cache) >= c.capacity {
			// Evict least recently used entry
			delete(c.cache, c.lruList.Back().Value.(*Entry).key)
			c.lruList.Remove(c.lruList.Back())
		}
		// Add new entry to the front of the list (most recently used)
		elem = c.lruList.PushFront(&Entry{key, value})
		c.cache[key] = elem
	}
}

func (c *cache) Delete(key string) {
	// Acquire a lock for the key
	mutex, _ := c.mutexes.LoadOrStore(key, &sync.Mutex{})
	mutex.(*sync.Mutex).Lock()
	defer mutex.(*sync.Mutex).Unlock()

	delete(c.cache, key)
	c.lruList.Remove(c.cache[key])
}
