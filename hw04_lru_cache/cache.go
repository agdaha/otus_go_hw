package hw04lrucache

import "sync"

type Key string

type cacheValue struct {
	key   Key
	value interface{}
}
type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	mu       sync.Mutex
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func (c *lruCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	clear(c.items) // = make(map[Key]*ListItem, c.capacity)
	c.queue = NewList()
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if item, ok := c.items[key]; ok {
		c.queue.MoveToFront(item)
		return item.Value.(cacheValue).value, true
	}
	return nil, false
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	v := cacheValue{
		key:   key,
		value: value,
	}

	if item, ok := c.items[key]; ok {
		item.Value = v
		c.queue.MoveToFront(item)
		return true
	}

	c.queue.PushFront(v)
	c.items[key] = c.queue.Front()

	if c.queue.Len() > c.capacity {
		item := c.queue.Back()
		c.queue.Remove(item)
		delete(c.items, item.Value.(cacheValue).key)
	}
	return false
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
