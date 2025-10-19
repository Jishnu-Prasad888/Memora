package store

import (
	"fmt"
	"sync"
	"time"
)

type Entry struct {
	Value      interface{}
	Expiration int64 // Unix nano timestamp, 0 means no expiration
}

type HashTable struct {
	mu      sync.RWMutex
	buckets []map[string]*Entry
	size    int
	count   int
}

func NewHashTable(size int) *HashTable {
	if size <= 0 {
		size = 1024
	}

	buckets := make([]map[string]*Entry, size)
	for i := range buckets {
		buckets[i] = make(map[string]*Entry)
	}

	return &HashTable{
		buckets: buckets,
		size:    size,
	}
}

func (h *HashTable) hash(key string) int {
	hash := 0
	for i := 0; i < len(key); i++ {
		hash = 31*hash + int(key[i])
	}
	return hash % h.size
}

func (h *HashTable) Set(key string, value interface{}, ttl time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	bucketIndex := h.hash(key)
	bucket := h.buckets[bucketIndex]

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	if _, exists := bucket[key]; !exists {
		h.count++
	}

	bucket[key] = &Entry{
		Value:      value,
		Expiration: expiration,
	}

	fmt.Printf("DEBUG: HashTable Set - key: '%s', value type: %T, value: %v\n", key, value, value) // Debug log
}

func (h *HashTable) Get(key string) (interface{}, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	bucketIndex := h.hash(key)
	bucket := h.buckets[bucketIndex]

	entry, exists := bucket[key]
	if !exists {
		return nil, false
	}

	if entry.Expiration > 0 && time.Now().UnixNano() > entry.Expiration {
		return nil, false
	}

	return entry.Value, true
}

func (h *HashTable) Delete(key string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	bucketIndex := h.hash(key)
	bucket := h.buckets[bucketIndex]

	if _, exists := bucket[key]; exists {
		delete(bucket, key)
		h.count--
		return true
	}

	return false
}

func (h *HashTable) Exists(key string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	bucketIndex := h.hash(key)
	bucket := h.buckets[bucketIndex]

	entry, exists := bucket[key]
	if !exists {
		return false
	}

	if entry.Expiration > 0 && time.Now().UnixNano() > entry.Expiration {
		return false
	}

	return true
}

func (h *HashTable) Keys(pattern string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	keys := make([]string, 0)
	now := time.Now().UnixNano()

	for _, bucket := range h.buckets {
		for key, entry := range bucket {
			// Check expiration
			if entry.Expiration > 0 && now > entry.Expiration {
				continue
			}

			// Simple pattern matching (supports * wildcard)
			if matchesPattern(key, pattern) {
				keys = append(keys, key)
			}
		}
	}

	return keys
}

func (h *HashTable) TTL(key string) int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	bucketIndex := h.hash(key)
	bucket := h.buckets[bucketIndex]

	entry, exists := bucket[key]
	if !exists {
		return -2 // Key doesn't exist
	}

	if entry.Expiration == 0 {
		return -1 // No expiration
	}

	ttl := (entry.Expiration - time.Now().UnixNano()) / int64(time.Second)
	if ttl < 0 {
		return -2 // Key expired
	}

	return ttl
}

func (h *HashTable) Expire(key string, ttl time.Duration) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	bucketIndex := h.hash(key)
	bucket := h.buckets[bucketIndex]

	entry, exists := bucket[key]
	if !exists {
		return false
	}

	if ttl > 0 {
		entry.Expiration = time.Now().Add(ttl).UnixNano()
	} else {
		entry.Expiration = 0
	}

	return true
}

func (h *HashTable) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}

func (h *HashTable) RemoveExpired() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	removed := 0
	now := time.Now().UnixNano()

	for _, bucket := range h.buckets {
		for key, entry := range bucket {
			if entry.Expiration > 0 && now > entry.Expiration {
				delete(bucket, key)
				removed++
				h.count--
			}
		}
	}

	return removed
}

func matchesPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Simple wildcard matching
	patternIndex, keyIndex := 0, 0
	patternLen, keyLen := len(pattern), len(key)

	for patternIndex < patternLen {
		if pattern[patternIndex] == '*' {
			patternIndex++
			if patternIndex >= patternLen {
				return true
			}

			for keyIndex < keyLen {
				if matchesPattern(key[keyIndex:], pattern[patternIndex:]) {
					return true
				}
				keyIndex++
			}
			return false
		} else if keyIndex < keyLen && (pattern[patternIndex] == '?' || pattern[patternIndex] == key[keyIndex]) {
			patternIndex++
			keyIndex++
		} else {
			return false
		}
	}

	return keyIndex == keyLen
}
