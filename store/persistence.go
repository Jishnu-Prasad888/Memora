package store

import (
	"encoding/gob"
	"log"
	"os"
	"sync"
	"time"
)

type Persistence struct {
	store    *DataStore
	filename string
	mu       sync.RWMutex
}

type Snapshot struct {
	StringData map[string]Entry
	ListData   map[string]Entry
	SetData    map[string]Entry
	HashData   map[string]Entry
	Timestamp  time.Time
}

func NewPersistence(store *DataStore, filename string) *Persistence {
	return &Persistence{
		store:    store,
		filename: filename,
	}
}

func (p *Persistence) Save() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	file, err := os.Create(p.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)

	// Create snapshot
	snapshot := p.createSnapshot()

	err = encoder.Encode(snapshot)
	if err != nil {
		return err
	}

	log.Printf("Snapshot saved to %s", p.filename)
	return nil
}

func (p *Persistence) createSnapshot() Snapshot {
	// This is a simplified snapshot - in production you'd want to handle
	// the actual data structures more carefully
	snapshot := Snapshot{
		StringData: make(map[string]Entry),
		ListData:   make(map[string]Entry),
		SetData:    make(map[string]Entry),
		HashData:   make(map[string]Entry),
		Timestamp:  time.Now(),
	}

	// Note: This is a placeholder - you'd need to implement proper serialization
	// for the actual data structures in your store

	return snapshot
}

func (p *Persistence) Load() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	file, err := os.Open(p.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No snapshot exists yet
		}
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var snapshot Snapshot

	err = decoder.Decode(&snapshot)
	if err != nil {
		return err
	}

	// Restore from snapshot
	// Note: This is a placeholder - implement proper deserialization
	log.Printf("Loaded snapshot from %s (created at %v)", p.filename, snapshot.Timestamp)

	return nil
}

func (p *Persistence) StartAutoSave(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		err := p.Save()
		if err != nil {
			log.Printf("Error saving snapshot: %v", err)
		}
	}
}

// BackgroundSave starts a background save using goroutine (for CPU-intensive operations)
func (p *Persistence) BackgroundSave() error {
	go func() {
		err := p.Save()
		if err != nil {
			log.Printf("Background save failed: %v", err)
		} else {
			log.Println("Background save completed successfully")
		}
	}()
	return nil
}
