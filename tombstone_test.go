package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestNewTombstoneStore(t *testing.T) {
	store := NewTombstoneStore()
	if store == nil {
		t.Fatal("NewTombstoneStore returned nil")
	}
	if store.Count() != 0 {
		t.Errorf("Expected empty store, got %d tombstones", store.Count())
	}
}

func TestCreateTombstone(t *testing.T) {
	store := NewTombstoneStore()
	entry := VocabEntry{
		Name:    "compute",
		Pattern: "arithmetic computation",
	}

	tombstone := store.Add(entry)

	// Verify tombstone fields
	if tombstone.Name != "compute" {
		t.Errorf("Expected Name 'compute', got '%s'", tombstone.Name)
	}
	if tombstone.Pattern != "arithmetic computation" {
		t.Errorf("Expected Pattern 'arithmetic computation', got '%s'", tombstone.Pattern)
	}
	if tombstone.Hash == "" {
		t.Error("Expected non-empty Hash")
	}
	if tombstone.PrunedAt.IsZero() {
		t.Error("Expected non-zero PrunedAt timestamp")
	}
	if len(tombstone.SemanticVector) != 128 {
		t.Errorf("Expected SemanticVector length 128, got %d", len(tombstone.SemanticVector))
	}

	// Verify tombstone was added to store
	if store.Count() != 1 {
		t.Errorf("Expected store count 1, got %d", store.Count())
	}
}

func TestHashConsistency(t *testing.T) {
	store := NewTombstoneStore()
	entry1 := VocabEntry{
		Name:    "factorial",
		Pattern: "compute factorial",
	}
	entry2 := VocabEntry{
		Name:    "factorial",
		Pattern: "compute factorial",
	}

	tombstone1 := store.Add(entry1)
	tombstone2 := store.Add(entry2)

	// Same entries should produce same hash
	if tombstone1.Hash != tombstone2.Hash {
		t.Errorf("Same entries produced different hashes: %s vs %s", tombstone1.Hash, tombstone2.Hash)
	}
}

func TestHashUniqueness(t *testing.T) {
	store := NewTombstoneStore()
	entry1 := VocabEntry{
		Name:    "add",
		Pattern: "IADD",
	}
	entry2 := VocabEntry{
		Name:    "multiply",
		Pattern: "IMUL",
	}

	tombstone1 := store.Add(entry1)
	tombstone2 := store.Add(entry2)

	// Different entries should produce different hashes
	if tombstone1.Hash == tombstone2.Hash {
		t.Errorf("Different entries produced same hash: %s", tombstone1.Hash)
	}
}

func TestHas(t *testing.T) {
	store := NewTombstoneStore()

	// Test with empty store
	if store.Has("compute") {
		t.Error("Has() returned true for empty store")
	}

	// Add a tombstone
	entry := VocabEntry{
		Name:    "factorial",
		Pattern: "compute factorial",
	}
	store.Add(entry)

	// Test Has for present name
	if !store.Has("factorial") {
		t.Error("Has() returned false for existing name 'factorial'")
	}

	// Test Has for absent name
	if store.Has("nonexistent") {
		t.Error("Has() returned true for non-existent name")
	}
}

func TestHasByHash(t *testing.T) {
	store := NewTombstoneStore()
	entry := VocabEntry{
		Name:    "hello",
		Pattern: "greeting",
	}

	tombstone := store.Add(entry)

	// Test HasByHash for present hash
	if !store.HasByHash(tombstone.Hash) {
		t.Error("HasByHash() returned false for existing hash")
	}

	// Test HasByHash for absent hash
	fakeHash := strings.Repeat("a", 64)
	if store.HasByHash(fakeHash) {
		t.Error("HasByHash() returned true for non-existent hash")
	}
}

func TestSerializeDeserialize(t *testing.T) {
	store := NewTombstoneStore()

	// Add multiple tombstones
	entries := []VocabEntry{
		{Name: "compute", Pattern: "arithmetic computation"},
		{Name: "factorial", Pattern: "compute factorial"},
		{Name: "hello", Pattern: "greeting"},
	}

	var hashes []string
	for _, entry := range entries {
		tombstone := store.Add(entry)
		hashes = append(hashes, tombstone.Hash)
	}

	// Serialize
	data := store.Serialize()
	if data == nil {
		t.Fatal("Serialize() returned nil")
	}
	if len(data) == 0 {
		t.Fatal("Serialize() returned empty data")
	}

	// Deserialize into new store
	newStore := NewTombstoneStore()
	err := newStore.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize() failed: %v", err)
	}

	// Verify all tombstones were restored
	if newStore.Count() != store.Count() {
		t.Errorf("Expected %d tombstones, got %d", store.Count(), newStore.Count())
	}

	// Verify specific tombstones by hash
	for _, hash := range hashes {
		if !newStore.HasByHash(hash) {
			t.Errorf("Tombstone with hash %s not found after deserialization", hash)
		}
	}

	// Verify name presence
	for _, entry := range entries {
		if !newStore.Has(entry.Name) {
			t.Errorf("Tombstone with name %s not found after deserialization", entry.Name)
		}
	}
}

func TestDeserializeInvalid(t *testing.T) {
	store := NewTombstoneStore()

	// Test with invalid JSON
	err := store.Deserialize([]byte("{invalid json}"))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	// Test with empty data
	err = store.Deserialize([]byte{})
	if err == nil {
		t.Error("Expected error for empty data, got nil")
	}
}

func TestProveOnceKnew(t *testing.T) {
	// Scenario: An agent learns a concept, then prunes it
	// Later, it should be able to prove it once knew it

	store := NewTombstoneStore()

	// Agent learns a concept
	originalEntry := VocabEntry{
		Name:    "quicksort",
		Pattern: "sorting algorithm",
	}

	// Agent prunes the concept, creating a tombstone
	tombstone := store.Add(originalEntry)

	// Verify agent can prove it once knew the concept
	if !store.Has("quicksort") {
		t.Error("Agent cannot prove it once knew 'quicksort'")
	}

	// Verify the hash matches what we expect
	hashInput := "quicksort:sorting algorithm"
	hashBytes := sha256.Sum256([]byte(hashInput))
	expectedHash := hex.EncodeToString(hashBytes[:])

	if tombstone.Hash != expectedHash {
		t.Errorf("Hash mismatch: got %s, expected %s", tombstone.Hash, expectedHash)
	}

	// Verify other agents cannot find concepts the agent never knew
	if store.Has("mergesort") {
		t.Error("Agent claims to know 'mergesort' which it never learned")
	}
}

func TestGetAll(t *testing.T) {
	store := NewTombstoneStore()

	// Add multiple tombstones
	entries := []VocabEntry{
		{Name: "a", Pattern: "pattern_a"},
		{Name: "b", Pattern: "pattern_b"},
		{Name: "c", Pattern: "pattern_c"},
	}

	for _, entry := range entries {
		store.Add(entry)
	}

	all := store.GetAll()

	if len(all) != 3 {
		t.Errorf("Expected 3 tombstones, got %d", len(all))
	}

	// Verify all entries are present
	found := make(map[string]bool)
	for _, tombstone := range all {
		found[tombstone.Name] = true
	}

	for _, entry := range entries {
		if !found[entry.Name] {
			t.Errorf("Tombstone for %s not found in GetAll()", entry.Name)
		}
	}
}

func TestMultiplePrunesSameConcept(t *testing.T) {
	store := NewTombstoneStore()

	// Prune the same concept multiple times
	entry := VocabEntry{
		Name:    "duplicate",
		Pattern: "test pattern",
	}

	store.Add(entry)
	store.Add(entry)
	store.Add(entry)

	// Should only have one tombstone
	if store.Count() != 1 {
		t.Errorf("Expected 1 tombstone after multiple adds, got %d", store.Count())
	}
}
