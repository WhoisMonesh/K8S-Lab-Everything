package progress

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const ProgressFile = ".lab-progress.json"

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

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
		return fmt.Sprintf("\n  %s%s%s\n\n  Run %scka-lab-runner lab run <id>%s to start your first lab!\n",
			colorDim, "No labs completed yet.", colorReset, colorCyan, colorReset)
	}

	var totalDuration time.Duration
	byCategory := make(map[string]int)
	byDifficulty := make(map[string]int)

	for _, r := range p.Labs {
		totalDuration += r.Duration
		byCategory[r.Category]++
		byDifficulty[r.Difficulty]++
	}

	result := fmt.Sprintf("\n  %s╔══════════════════════════════════════════════════╗%s\n", colorCyan, colorReset)
	result += fmt.Sprintf("  %s║%s  %s%-48s%s  %s║%s\n", colorCyan, colorReset, colorBold, "Lab Progress", colorReset, colorCyan, colorReset)
	result += fmt.Sprintf("  %s╚══════════════════════════════════════════════════╝%s\n", colorCyan, colorReset)
	result += "\n"

	result += fmt.Sprintf("  %s%sCompleted:%s  %s%d lab(s)%s\n", colorBold, colorReset, colorReset, colorGreen, total, colorReset)
	result += "\n"

	result += fmt.Sprintf("  %sBy Difficulty:%s\n", colorBold, colorReset)
	for _, d := range []string{"easy", "medium", "hard"} {
		if c, ok := byDifficulty[d]; ok {
			color := colorGreen
			if d == "medium" {
				color = colorYellow
			} else if d == "hard" {
				color = colorRed
			}
			bar := strings.Repeat("█", c)
			result += fmt.Sprintf("    %s%-8s%s %s%s%s  %d\n", colorDim, d, colorReset, color, bar, colorReset, c)
		}
	}
	result += "\n"

	result += fmt.Sprintf("  %sBy Category:%s\n", colorBold, colorReset)
	cats := make([]string, 0, len(byCategory))
	for c := range byCategory {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	for _, c := range cats {
		bar := strings.Repeat("█", byCategory[c])
		result += fmt.Sprintf("    %-16s %s%s%s  %d\n", c, colorCyan, bar, colorReset, byCategory[c])
	}

	result += fmt.Sprintf("\n  %sTotal time:%s  %s%s%s\n", colorBold, colorReset, colorWhite, totalDuration.Round(time.Second), colorReset)
	result += "\n"
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
