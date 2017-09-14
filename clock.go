package gomemc3

import "container/ring"
import "sync"

type ClockReplacer struct {
	maxEntries uint
	clockCount uint8
	hand       *ring.Ring
	cache      map[interface{}]*ring.Ring
	mutex      sync.RWMutex
	onEvict    func(interface{}, interface{})
}

// An entry in the cache is represented by a key/value pair and an access counter
type entry struct {
	AccessCounter uint8
	Key           interface{}
	Value         interface{}
}

// New creates a new cache with a maximum size limit.
func New(maxEntries uint, clockCount uint8, onEvict func(interface{}, interface{})) *ClockReplacer {
	r := ring.New(int(maxEntries))
	r.Value = &entry{}
	for p := r.Next(); p != r; p = p.Next() {
		p.Value = &entry{}
	}
	return &ClockReplacer{
		maxEntries: maxEntries,
		clockCount: clockCount,
		hand:       r,
		cache:      make(map[interface{}]*ring.Ring),
		onEvict:    onEvict}
}

// Put inserts a new entry into the cache
func (c *ClockReplacer) Put(key interface{}, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if e, ok := c.cache[key]; ok {
		e.Value.(*entry).AccessCounter = c.clockCount
		e.Value.(*entry).Value = value
	} else {
		c.evict()
		c.cache[key] = c.hand
		c.hand.Value.(*entry).AccessCounter = c.clockCount
		c.hand.Value.(*entry).Key = key
		c.hand.Value.(*entry).Value = value
		c.hand = c.hand.Next()
	}
}

// Get look up a key in the cache
func (c *ClockReplacer) Get(key interface{}) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if e, ok := c.cache[key]; ok {
		if e.Value.(*entry).AccessCounter != c.clockCount {
			e.Value.(*entry).AccessCounter = c.clockCount
		}
		return e.Value.(*entry).Value, true
	}
	return nil, false
}

// Delete an entry with a given key
func (c *ClockReplacer) Delete(key interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if e, ok := c.cache[key]; ok {
		delete(c.cache, key)
		e.Value.(*entry).AccessCounter = 0
	}
}

func (c *ClockReplacer) evict() {
	for c.hand.Value.(*entry).AccessCounter > 0 {
		c.hand.Value.(*entry).AccessCounter--
		c.hand = c.hand.Next()
	}
	if c.hand.Value.(*entry).Key != nil {
		delete(c.cache, c.hand.Value.(*entry).Key)
		if c.onEvict != nil {
			go c.onEvict(c.hand.Value.(*entry).Key, c.hand.Value.(*entry).Value)
		}
	}
}
