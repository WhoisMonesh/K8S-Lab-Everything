package labs

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
)

var (
	registry = make(map[string]Lab)
	mu       sync.RWMutex
)

// Register registers a lab in the global registry
func Register(lab Lab) {
	mu.Lock()
	defer mu.Unlock()

	id := lab.ID()
	if _, exists := registry[id]; exists {
		panic(fmt.Sprintf("lab with ID %q already registered", id))
	}

	registry[id] = lab
}

// Get retrieves a lab by ID
func Get(id string) (Lab, error) {
	mu.RLock()
	defer mu.RUnlock()

	lab, exists := registry[id]
	if !exists {
		return nil, fmt.Errorf("lab with ID %q not found", id)
	}

	return lab, nil
}

// List returns all registered labs
func List() []Lab {
	mu.RLock()
	defer mu.RUnlock()

	labs := make([]Lab, 0, len(registry))
	for _, lab := range registry {
		labs = append(labs, lab)
	}

	// Sort by ID for consistent ordering
	sort.Slice(labs, func(i, j int) bool {
		return labs[i].ID() < labs[j].ID()
	})

	return labs
}

// ListByCategory returns labs filtered by category
func ListByCategory(category Category) []Lab {
	all := List()
	filtered := make([]Lab, 0)

	for _, lab := range all {
		if lab.Category() == category {
			filtered = append(filtered, lab)
		}
	}

	return filtered
}

// ListByDifficulty returns labs filtered by difficulty
func ListByDifficulty(difficulty Difficulty) []Lab {
	all := List()
	filtered := make([]Lab, 0)

	for _, lab := range all {
		if lab.Difficulty() == difficulty {
			filtered = append(filtered, lab)
		}
	}

	return filtered
}

// Random returns a random lab, optionally filtered by category and difficulty
func Random(seed int64, category Category, difficulty Difficulty) (Lab, error) {
	all := List()
	filtered := make([]Lab, 0)

	for _, lab := range all {
		matches := true
		if category != "" && lab.Category() != category {
			matches = false
		}
		if difficulty != "" && lab.Difficulty() != difficulty {
			matches = false
		}
		if matches {
			filtered = append(filtered, lab)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no labs found matching criteria")
	}

	rng := rand.New(rand.NewSource(seed))
	idx := rng.Intn(len(filtered))
	return filtered[idx], nil
}

// IDs returns a sorted list of all lab IDs
func IDs() []string {
	mu.RLock()
	defer mu.RUnlock()

	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}

	sort.Strings(ids)
	return ids
}
