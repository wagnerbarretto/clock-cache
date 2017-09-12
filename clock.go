package gomemc3

import "container/ring"
import "sync"

type ClockReplacer struct {
	maxEntries int
	hand       *ring.Ring
	cache      map[interface{}]*ring.Ring
	mutex      sync.RWMutex
	onEvict    func(interface{}, interface{})
}

// An entry in the cache is represented by a key/value pair and a Recency bit (R)
type entry struct {
	R     bool
	Key   interface{}
	Value interface{}
}

// New creates a new cache with a maximum size limit. If maxEntries is 0 the cache has no limit
func New(maxEntries int, onEvict func(interface{}, interface{})) *ClockReplacer {
	return &ClockReplacer{
		maxEntries: maxEntries,
		cache:      make(map[interface{}]*ring.Ring),
		onEvict:    onEvict}
}

// Put inserts a new entry into the cache
func (c *ClockReplacer) Put(key interface{}, value interface{}) {
	c.mutex.Lock()
	if e, ok := c.cache[key]; ok {
		e.Value.(*entry).R = false
		e.Value.(*entry).Value = value
	} else if c.maxEntries > 0 && c.maxEntries == len(c.cache) {
		c.evictAndAllocate(key, value)
	} else {
		c.allocateNewEntry(key, value)
	}
	c.mutex.Unlock()
}

// Get look up a key in the cache
func (c *ClockReplacer) Get(key interface{}) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if e, ok := c.cache[key]; ok {
		e.Value.(*entry).R = true
		return e.Value.(*entry).Value, true
	}
	return nil, false
}

// Delete an entry with a given key
func (c *ClockReplacer) Delete(key interface{}) {
	c.mutex.Lock()
	if e, ok := c.cache[key]; ok {
		e.Prev().Unlink(1)
		delete(c.cache, key)
	}
	c.mutex.Unlock()
}

func (c *ClockReplacer) allocateNewEntry(key interface{}, value interface{}) {
	e := entry{false, key, value}
	r := &ring.Ring{Value: &e}
	if c.hand != nil {
		c.hand.Link(r)
	}
	c.hand = r
	c.cache[key] = r
}

func (c *ClockReplacer) evictAndAllocate(key interface{}, value interface{}) {
	c.hand = c.hand.Next()
	for c.hand.Value.(*entry).R {
		c.hand.Value.(*entry).R = false
		c.hand = c.hand.Next()
	}
	oldK := c.hand.Value.(*entry).Key
	oldV := c.hand.Value.(*entry).Value
	delete(c.cache, oldK)
	c.cache[key] = c.hand
	c.hand.Value.(*entry).R = false
	c.hand.Value.(*entry).Key = key
	c.hand.Value.(*entry).Value = value
	if c.onEvict != nil {
		go c.onEvict(oldK, oldV)
	}
}
