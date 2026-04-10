package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// VocabEntry represents a single vocabulary entry
type VocabEntry struct {
	Name    string
	Pattern string
}

// Tombstone represents a pruned vocabulary entry
type Tombstone struct {
	Name           string    `json:"name"`
	Pattern        string    `json:"pattern"`
	Hash           string    `json:"hash"`
	PrunedAt       time.Time `json:"pruned_at"`
	SemanticVector []float32 `json:"semantic_vector"`
}

// TombstoneStore manages tombstones for pruned vocabulary entries
type TombstoneStore struct {
	tombstones map[string]Tombstone
	mu         struct{}
}

// NewTombstoneStore creates a new tombstone store
func NewTombstoneStore() *TombstoneStore {
	return &TombstoneStore{
		tombstones: make(map[string]Tombstone),
	}
}

// Add creates a tombstone from a pruned vocabulary entry
func (s *TombstoneStore) Add(entry VocabEntry) Tombstone {
	// Compute SHA256 hash of the original entry content
	hashInput := entry.Name + ":" + entry.Pattern
	hashBytes := sha256.Sum256([]byte(hashInput))
	hash := hex.EncodeToString(hashBytes[:])

	tombstone := Tombstone{
		Name:           entry.Name,
		Pattern:        entry.Pattern,
		Hash:           hash,
		PrunedAt:       time.Now().UTC(),
		SemanticVector: make([]float32, 128), // Placeholder: 128-dimensional semantic vector
	}

	s.tombstones[tombstone.Hash] = tombstone
	return tombstone
}

// Has checks if the agent once knew a concept by name
func (s *TombstoneStore) Has(name string) bool {
	for _, tombstone := range s.tombstones {
		if tombstone.Name == name {
			return true
		}
	}
	return false
}

// HasByHash checks if the agent once knew a concept by hash
func (s *TombstoneStore) HasByHash(hash string) bool {
	_, exists := s.tombstones[hash]
	return exists
}

// Serialize converts the tombstone store to JSON for repo signaling
func (s *TombstoneStore) Serialize() []byte {
	data, err := json.MarshalIndent(s.tombstones, "", "  ")
	if err != nil {
		return nil
	}
	return data
}

// Deserialize loads tombstones from JSON data
func (s *TombstoneStore) Deserialize(data []byte) error {
	tombstones := make(map[string]Tombstone)
	err := json.Unmarshal(data, &tombstones)
	if err != nil {
		return err
	}
	s.tombstones = tombstones
	return nil
}

// Count returns the number of tombstones
func (s *TombstoneStore) Count() int {
	return len(s.tombstones)
}

// GetAll returns all tombstones
func (s *TombstoneStore) GetAll() []Tombstone {
	result := make([]Tombstone, 0, len(s.tombstones))
	for _, tombstone := range s.tombstones {
		result = append(result, tombstone)
	}
	return result
}
