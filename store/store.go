package store

import (
	"fmt"
	"sync"
	"time"
)

type DataStore struct {
	mu          sync.RWMutex
	stringStore *HashTable
	listStore   *HashTable
	setStore    *HashTable
	hashStore   *HashTable
}

func NewDataStore() *DataStore {
	return &DataStore{
		stringStore: NewHashTable(1024),
		listStore:   NewHashTable(512),
		setStore:    NewHashTable(512),
		hashStore:   NewHashTable(512),
	}
}

// Set String operations
func (ds *DataStore) Set(key string, value interface{}, ttl time.Duration) {
	ds.stringStore.Set(key, value, ttl)
}

func (ds *DataStore) Get(key string) (interface{}, bool) {
	return ds.stringStore.Get(key)
}

func (ds *DataStore) Delete(key string) bool {
	deleted := ds.stringStore.Delete(key)
	deleted = ds.listStore.Delete(key) || deleted
	deleted = ds.setStore.Delete(key) || deleted
	deleted = ds.hashStore.Delete(key) || deleted
	return deleted
}

func (ds *DataStore) Exists(key string) bool {
	return ds.stringStore.Exists(key) ||
		ds.listStore.Exists(key) ||
		ds.setStore.Exists(key) ||
		ds.hashStore.Exists(key)
}

func (ds *DataStore) Keys(pattern string) []string {
	keysMap := make(map[string]bool)

	for _, key := range ds.stringStore.Keys(pattern) {
		keysMap[key] = true
	}
	for _, key := range ds.listStore.Keys(pattern) {
		keysMap[key] = true
	}
	for _, key := range ds.setStore.Keys(pattern) {
		keysMap[key] = true
	}
	for _, key := range ds.hashStore.Keys(pattern) {
		keysMap[key] = true
	}

	keys := make([]string, 0, len(keysMap))
	for key := range keysMap {
		keys = append(keys, key)
	}

	return keys
}

func (ds *DataStore) TTL(key string) int64 {
	// Check all stores
	if ttl := ds.stringStore.TTL(key); ttl != -2 {
		return ttl
	}
	if ttl := ds.listStore.TTL(key); ttl != -2 {
		return ttl
	}
	if ttl := ds.setStore.TTL(key); ttl != -2 {
		return ttl
	}
	if ttl := ds.hashStore.TTL(key); ttl != -2 {
		return ttl
	}
	return -2
}

func (ds *DataStore) Expire(key string, ttl time.Duration) bool {
	expired := ds.stringStore.Expire(key, ttl)
	expired = ds.listStore.Expire(key, ttl) || expired
	expired = ds.setStore.Expire(key, ttl) || expired
	expired = ds.hashStore.Expire(key, ttl) || expired
	return expired
}

// LPush List operations
func (ds *DataStore) LPush(key string, values ...interface{}) int {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	var list []interface{}
	if existing, ok := ds.listStore.Get(key); ok {
		list = existing.([]interface{})
	}

	list = append(values, list...)
	ds.listStore.Set(key, list, 0)
	return len(list)
}

func (ds *DataStore) RPush(key string, values ...interface{}) int {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	var list []interface{}
	if existing, ok := ds.listStore.Get(key); ok {
		list = existing.([]interface{})
	}

	list = append(list, values...)
	ds.listStore.Set(key, list, 0)
	return len(list)
}

func (ds *DataStore) LPop(key string) interface{} {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	existing, ok := ds.listStore.Get(key)
	if !ok {
		return nil
	}

	list := existing.([]interface{})
	if len(list) == 0 {
		return nil
	}

	value := list[0]
	list = list[1:]
	ds.listStore.Set(key, list, 0)
	return value
}

func (ds *DataStore) RPop(key string) interface{} {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	existing, ok := ds.listStore.Get(key)
	if !ok {
		return nil
	}

	list := existing.([]interface{})
	if len(list) == 0 {
		return nil
	}

	value := list[len(list)-1]
	list = list[:len(list)-1]
	ds.listStore.Set(key, list, 0)
	return value
}

func (ds *DataStore) LLen(key string) int {
	existing, ok := ds.listStore.Get(key)
	if !ok {
		return 0
	}

	list := existing.([]interface{})
	return len(list)
}

// SAdd Set operations
func (ds *DataStore) SAdd(key string, members ...interface{}) int {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	set := make(map[interface{}]bool)
	if existing, ok := ds.setStore.Get(key); ok {
		set = existing.(map[interface{}]bool)
	}

	added := 0
	for _, member := range members {
		if !set[member] {
			set[member] = true
			added++
		}
	}

	ds.setStore.Set(key, set, 0)
	return added
}

func (ds *DataStore) SRem(key string, members ...interface{}) int {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	existing, ok := ds.setStore.Get(key)
	if !ok {
		return 0
	}

	set := existing.(map[interface{}]bool)
	removed := 0
	for _, member := range members {
		if set[member] {
			delete(set, member)
			removed++
		}
	}

	ds.setStore.Set(key, set, 0)
	return removed
}

func (ds *DataStore) SMembers(key string) []interface{} {
	existing, ok := ds.setStore.Get(key)
	if !ok {
		fmt.Printf("DEBUG: Set '%s' not found\n", key) // Debug log
		return nil
	}

	set, ok := existing.(map[interface{}]bool)
	if !ok {
		fmt.Printf("DEBUG: Invalid set type for key '%s': %T\n", key, existing) // Debug log
		return nil
	}

	members := make([]interface{}, 0, len(set))
	for member := range set {
		members = append(members, member)
		fmt.Printf("DEBUG: Found member '%v' in set '%s'\n", member, key) // Debug log
	}

	return members
}

func (ds *DataStore) SIsMember(key string, member interface{}) bool {
	existing, ok := ds.setStore.Get(key)
	if !ok {
		return false
	}

	set := existing.(map[interface{}]bool)
	return set[member]
}

// HSet Hash operations
func (ds *DataStore) HSet(key string, field string, value interface{}) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	hash := make(map[string]interface{})
	if existing, ok := ds.hashStore.Get(key); ok {
		hash = existing.(map[string]interface{})
	}

	exists := hash[field] != nil
	hash[field] = value
	ds.hashStore.Set(key, hash, 0)
	return !exists
}

func (ds *DataStore) HGet(key, field string) interface{} {
	existing, ok := ds.hashStore.Get(key)
	if !ok {
		return nil
	}

	hash := existing.(map[string]interface{})
	return hash[field]
}

func (ds *DataStore) HDel(key string, fields ...string) int {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	existing, ok := ds.hashStore.Get(key)
	if !ok {
		return 0
	}

	hash := existing.(map[string]interface{})
	deleted := 0
	for _, field := range fields {
		if hash[field] != nil {
			delete(hash, field)
			deleted++
		}
	}

	ds.hashStore.Set(key, hash, 0)
	return deleted
}

func (ds *DataStore) HGetAll(key string) map[string]interface{} {
	existing, ok := ds.hashStore.Get(key)
	if !ok {
		return nil
	}

	return existing.(map[string]interface{})
}

func (ds *DataStore) RemoveExpired() int {
	removed := ds.stringStore.RemoveExpired()
	removed += ds.listStore.RemoveExpired()
	removed += ds.setStore.RemoveExpired()
	removed += ds.hashStore.RemoveExpired()
	return removed
}

func (ds *DataStore) FlushAll() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.stringStore = NewHashTable(1024)
	ds.listStore = NewHashTable(512)
	ds.setStore = NewHashTable(512)
	ds.hashStore = NewHashTable(512)
}
