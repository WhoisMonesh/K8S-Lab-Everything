package progress

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const ProgressFile = ".lab-progress.json"

type LabResult struct {
	LabID       string        `json:"lab_id"`
	Title       string        `json:"title"`
	Category    string        `json:"category"`
	Difficulty  string        `json:"difficulty"`
	CompletedAt time.Time     `json:"completed_at"`
	Duration    time.Duration `json:"duration_seconds"`
	Estimated   int           `json:"estimated_minutes"`
	Timed       bool          `json:"timed,omitempty"`
	TimedOut    bool          `json:"timed_out,omitempty"`
	Namespace   string        `json:"namespace,omitempty"`
}

type Progress struct {
	Labs      map[string]*LabResult `json:"labs"`
	StartedAt time.Time             `json:"started_at"`
	mu        sync.RWMutex
}

var (
	current    *Progress
	progressMu sync.Mutex
)

func filePath() string {
	if dir, err := os.Getwd(); err == nil {
		return filepath.Join(dir, ProgressFile)
	}
	return ProgressFile
}

func Load() *Progress {
	progressMu.Lock()
	defer progressMu.Unlock()

	if current != nil {
		return current
	}

	data, err := os.ReadFile(filePath())
	if err != nil {
		current = &Progress{
			Labs:      make(map[string]*LabResult),
			StartedAt: time.Now(),
		}
		return current
	}

	var p Progress
	if err := json.Unmarshal(data, &p); err != nil {
		current = &Progress{
			Labs:      make(map[string]*LabResult),
			StartedAt: time.Now(),
		}
		return current
	}

	if p.Labs == nil {
		p.Labs = make(map[string]*LabResult)
	}
	current = &p
	return current
}

func Save() error {
	progressMu.Lock()
	defer progressMu.Unlock()

	if current == nil {
		return nil
	}

	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling progress: %w", err)
	}

	return os.WriteFile(filePath(), data, 0644)
}

func RecordCompletion(labID, title, category, difficulty string, duration time.Duration, estimated int, timed, timedOut bool, namespace string) {
	p := Load()
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Labs[labID] = &LabResult{
		LabID:       labID,
		Title:       title,
		Category:    category,
		Difficulty:  difficulty,
		CompletedAt: time.Now(),
		Duration:    duration,
		Estimated:   estimated,
		Timed:       timed,
		TimedOut:    timedOut,
		Namespace:   namespace,
	}

	Save()
}

func IsCompleted(labID string) bool {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, ok := p.Labs[labID]
	return ok
}

func GetResult(labID string) *LabResult {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	r, _ := p.Labs[labID]
	return r
}

func CompletedCount() int {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.Labs)
}

func Summary() string {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := len(p.Labs)
	if total == 0 {
		return "No labs completed yet. Run 'cka-lab-runner lab run <id>' to start!"
	}

	var totalDuration time.Duration
	byCategory := make(map[string]int)
	byDifficulty := make(map[string]int)

	for _, r := range p.Labs {
		totalDuration += r.Duration
		byCategory[r.Category]++
		byDifficulty[r.Difficulty]++
	}

	result := fmt.Sprintf("Progress: %d labs completed\n\n", total)
	result += "By Difficulty:\n"
	for _, d := range []string{"easy", "medium", "hard"} {
		if c, ok := byDifficulty[d]; ok {
			result += fmt.Sprintf("  %-8s %d\n", d, c)
		}
	}
	result += "\nBy Category:\n"
	cats := make([]string, 0, len(byCategory))
	for c := range byCategory {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	for _, c := range cats {
		result += fmt.Sprintf("  %-18s %d\n", c, byCategory[c])
	}

	result += fmt.Sprintf("\nTotal time: %s\n", totalDuration.Round(time.Second))
	return result
}

func ExportJSON() ([]byte, error) {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	results := make([]*LabResult, 0, len(p.Labs))
	for _, r := range p.Labs {
		results = append(results, r)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CompletedAt.Before(results[j].CompletedAt)
	})

	return json.MarshalIndent(struct {
		StartedAt time.Time    `json:"started_at"`
		Total     int          `json:"total_completed"`
		Labs      []*LabResult `json:"labs"`
	}{
		StartedAt: p.StartedAt,
		Total:     len(results),
		Labs:      results,
	}, "", "  ")
}
