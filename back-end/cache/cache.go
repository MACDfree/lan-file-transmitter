package cache

import (
	"container/list"
)

// Cache 是一个LRU的缓存实现
type Cache struct {
	maxCount  int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value interface{})
}

type Entry struct {
	Key   string
	Value interface{}
}

// New 实例化一个Cache实例
func New(maxCount int64, onEvicted func(key string, value interface{})) *Cache {
	return &Cache{
		maxCount:  maxCount,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get 从Cache中根据key获取value
func (c *Cache) Get(key string) (value interface{}, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*Entry)
		return kv.Value, true
	}
	return
}

// Add 增加一个key-value到缓存中
func (c *Cache) Add(key string, value interface{}) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*Entry)
		kv.Value = value
		return
	}
	ele := c.ll.PushFront(&Entry{key, value})
	c.cache[key] = ele
	if int64(c.ll.Len()) > c.maxCount {
		c.RemoveOldest()
	}
}

// Remove 删除指定key的缓存
func (c *Cache) Remove(key string) {
	if ele, ok := c.cache[key]; ok {
		c.ll.Remove(ele)
		delete(c.cache, key)
	}
}

// RemoveOldest 删除最老的key-value对
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*Entry)
		delete(c.cache, kv.Key)
		if c.OnEvicted != nil {
			c.OnEvicted(kv.Key, kv.Value)
		}
	}
}

// List 列出所有键名
func (c *Cache) List() []*Entry {
	list := make([]*Entry, 0)

	element := c.ll.Front()
	if element == nil {
		return list
	}
	list = append(list, element.Value.(*Entry))

	for {
		element = element.Next()
		if element != nil {
			list = append(list, element.Value.(*Entry))
		} else {
			break
		}
	}

	return list
}
