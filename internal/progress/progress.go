package progress

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const ProgressFile = ".lab-progress.json"

const (
	reset      = "\033[0m"
	bold       = "\033[1m"
	dim        = "\033[2m"
	italic     = "\033[3m"
	red        = "\033[31m"
	green      = "\033[32m"
	yellow     = "\033[33m"
	blue       = "\033[34m"
	magenta    = "\033[35m"
	cyan       = "\033[36m"
	white      = "\033[37m"
	brightCyan = "\033[96m"
	dimWhite   = "\033[37;2m"
	dimCyan    = "\033[36;2m"
	dimGreen   = "\033[32;2m"
	dimYellow  = "\033[33;2m"
	dimRed     = "\033[31;2m"
	bgGreen    = "\033[42m"
	bgYellow   = "\033[43m"
	bgRed      = "\033[41m"
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
	Rating      int           `json:"rating,omitempty"`
	Attempts    []Attempt     `json:"attempts,omitempty"`
}

// Attempt records a single verify attempt for benchmark tracking.
type Attempt struct {
	At      time.Time     `json:"at"`
	Passed  bool          `json:"passed"`
	Elapsed time.Duration `json:"elapsed_seconds"`
}

type Progress struct {
	Labs          map[string]*LabResult `json:"labs"`
	StartedAt     time.Time             `json:"started_at"`
	LastActiveLab string                `json:"last_active_lab,omitempty"`
	mu            sync.RWMutex
}

var (
	current    *Progress
	progressMu sync.Mutex
)

func filePath() string {
	if home, err := os.UserHomeDir(); err == nil {
		configDir := filepath.Join(home, ".config", "cka-lab-runner")
		os.MkdirAll(configDir, 0o755)
		return filepath.Join(configDir, ProgressFile)
	}
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

// CurrentStreak returns the number of consecutive days with at least one completed lab.
func CurrentStreak() int {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.Labs) == 0 {
		return 0
	}

	days := make(map[string]bool)
	for _, r := range p.Labs {
		days[r.CompletedAt.Format("2006-01-02")] = true
	}

	streak := 0
	d := time.Now()
	for {
		key := d.Format("2006-01-02")
		if !days[key] {
			break
		}
		streak++
		d = d.Add(-24 * time.Hour)
	}
	return streak
}

func progressBar(completed, total, width int) string {
	pct := 0
	if total > 0 {
		pct = completed * 100 / total
	}
	filled := pct * width / 100
	empty := width - filled
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("%s%s%s %d/%d (%d%%)",
		green, bar, reset, completed, total, pct)
}

func miniBar(count, max, width int) string {
	filled := 0
	if max > 0 {
		filled = count * width / max
	}
	empty := width - filled
	return fmt.Sprintf("%s%s%s", green, strings.Repeat("█", filled), dim+strings.Repeat("░", empty)+reset)
}

func Summary() string {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := len(p.Labs)
	if total == 0 {
		return fmt.Sprintf("\n  %s%s%s\n\n  %s▸%s Run %scka-lab-runner lab run <id>%s to start your first lab!\n\n",
			dim+italic, "No labs completed yet.", reset,
			cyan, reset, brightCyan, reset)
	}

	var totalDuration time.Duration
	byCategory := make(map[string]int)
	byDifficulty := make(map[string]int)

	for _, r := range p.Labs {
		totalDuration += r.Duration
		byCategory[r.Category]++
		byDifficulty[r.Difficulty]++
	}

	w := 52
	bar := func() string {
		return fmt.Sprintf("  %s┌%s%s┐%s\n", cyan, strings.Repeat("─", w), cyan, reset)
	}
	barEnd := func() string {
		return fmt.Sprintf("  %s└%s%s┘%s\n", cyan, strings.Repeat("─", w), cyan, reset)
	}

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(bar())
	b.WriteString(fmt.Sprintf("  %s│%s  %s%-*s%s  %s│%s\n", cyan, reset, bold, w-4, "Lab Progress Dashboard", reset, cyan, reset))
	b.WriteString(barEnd())
	b.WriteString("\n")

	// Stats row
	b.WriteString(fmt.Sprintf("  %s%sCompleted%s  %s%s%s\n", bold, reset, reset, green, bold, reset))
	b.WriteString(fmt.Sprintf("  %s  %d lab(s)%s\n\n", green, total, reset))

	// Difficulty breakdown
	b.WriteString(fmt.Sprintf("  %sBy Difficulty%s\n", bold, reset))
	maxDiff := 0
	for _, c := range byDifficulty {
		if c > maxDiff {
			maxDiff = c
		}
	}
	for _, d := range []string{"easy", "medium", "hard"} {
		if c, ok := byDifficulty[d]; ok {
			color := green
			labelColor := dimGreen
			if d == "medium" {
				color = yellow
				labelColor = dimYellow
			} else if d == "hard" {
				color = red
				labelColor = dimRed
			}
			b.WriteString(fmt.Sprintf("    %s%-8s%s  %s %s%d%s\n",
				labelColor, d, reset, miniBar(c, maxDiff, 20), color, c, reset))
		}
	}
	b.WriteString("\n")

	// Category breakdown
	b.WriteString(fmt.Sprintf("  %sBy Category%s\n", bold, reset))
	maxCat := 0
	for _, c := range byCategory {
		if c > maxCat {
			maxCat = c
		}
	}
	cats := make([]string, 0, len(byCategory))
	for c := range byCategory {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	for _, c := range cats {
		b.WriteString(fmt.Sprintf("    %-16s  %s %s%d%s\n",
			c, miniBar(byCategory[c], maxCat, 20), cyan, byCategory[c], reset))
	}
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("  %sTotal time:%s  %s%s%s\n", bold, reset, dimWhite, totalDuration.Round(time.Second), reset))
	b.WriteString("\n")

	return b.String()
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

func RateLab(labID string, rating int) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	p := Load()
	p.mu.Lock()
	defer p.mu.Unlock()

	r, ok := p.Labs[labID]
	if !ok {
		return fmt.Errorf("lab %s not completed yet", labID)
	}

	r.Rating = rating
	return Save()
}

func (p *Progress) ResetProgress(labIDs []string, all bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if all {
		p.Labs = make(map[string]*LabResult)
	} else {
		for _, id := range labIDs {
			delete(p.Labs, id)
		}
	}
	Save()
}

func GetRating(labID string) int {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	r, ok := p.Labs[labID]
	if !ok {
		return 0
	}
	return r.Rating
}

func AverageRating() float64 {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	total, count := 0, 0
	for _, r := range p.Labs {
		if r.Rating > 0 {
			total += r.Rating
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return float64(total) / float64(count)
}

func Streak() (current int, longest int, lastPracticeDate string) {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.Labs) == 0 {
		return 0, 0, ""
	}

	dates := make(map[time.Time]bool)
	for _, r := range p.Labs {
		d := r.CompletedAt.Truncate(24 * time.Hour)
		dates[d] = true
	}

	sortedDates := make([]time.Time, 0, len(dates))
	for d := range dates {
		sortedDates = append(sortedDates, d)
	}
	sort.Slice(sortedDates, func(i, j int) bool {
		return sortedDates[i].Before(sortedDates[j])
	})

	lastDate := sortedDates[len(sortedDates)-1]
	today := time.Now().Truncate(24 * time.Hour)

	if lastDate.Before(today) {
		return 0, len(sortedDates), lastDate.Format("2006-01-02")
	}

	longest = 1
	current = 1
	for i := len(sortedDates) - 1; i > 0; i-- {
		diff := sortedDates[i].Sub(sortedDates[i-1])
		if diff == 24*time.Hour {
			current++
			if current > longest {
				longest = current
			}
		} else {
			current = 1
		}
	}

	return current, longest, lastDate.Format("2006-01-02")
}

func StreakInfo() string {
	current, longest, lastDate := Streak()

	if current == 0 && longest == 0 {
		return "  No practice streak yet. Complete a lab to start one!\n"
	}

	var b strings.Builder
	if current > 0 {
		b.WriteString(fmt.Sprintf("  %s🔥 Current Streak:%s %s%d day(s)%s\n", bold, reset, green, current, reset))
	} else {
		b.WriteString(fmt.Sprintf("  %s🔥 Current Streak:%s %s0 day(s)%s (last practice: %s)\n", bold, reset, dimWhite, reset, lastDate))
	}
	b.WriteString(fmt.Sprintf("  %s🏆 Longest Streak:%s  %s%d day(s)%s\n", bold, reset, yellow, longest, reset))
	return b.String()
}

// Results returns the sorted list of all lab results (newest completion first).
func Results() []*LabResult {
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
	return results
}

// ImportFromFile merges progress from a JSON file, keeping the later completion
// per lab when both exist.
func ImportFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading import file: %w", err)
	}

	var imported struct {
		Labs []*LabResult `json:"labs"`
	}
	if err := json.Unmarshal(data, &imported); err != nil {
		return fmt.Errorf("parsing import file: %w", err)
	}

	p := Load()
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, r := range imported.Labs {
		if r == nil || r.LabID == "" {
			continue
		}
		if existing, ok := p.Labs[r.LabID]; ok && existing.CompletedAt.After(r.CompletedAt) {
			continue // keep the more recent one
		}
		if existing, ok := p.Labs[r.LabID]; ok && existing.CompletedAt.Equal(r.CompletedAt) {
			continue
		}
		p.Labs[r.LabID] = r
	}

	return Save()
}

func ExportCSV() (string, error) {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	var b strings.Builder
	b.WriteString("lab_id,title,category,difficulty,completed_at,duration_seconds,estimated_minutes,timed,timed_out,rating\n")

	ids := make([]string, 0, len(p.Labs))
	for id := range p.Labs {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		r := p.Labs[id]
		b.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%d,%d,%t,%t,%d\n",
			escapeCSV(r.LabID),
			escapeCSV(r.Title),
			escapeCSV(r.Category),
			escapeCSV(r.Difficulty),
			r.CompletedAt.Format(time.RFC3339),
			int(r.Duration.Seconds()),
			r.Estimated,
			r.Timed,
			r.TimedOut,
			r.Rating,
		))
	}

	return b.String(), nil
}

func ExportHTML() (string, error) {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := len(p.Labs)
	var totalDuration time.Duration
	for _, r := range p.Labs {
		totalDuration += r.Duration
	}

	var rows strings.Builder
	ids := make([]string, 0, total)
	for id := range p.Labs {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		r := p.Labs[id]
		rows.WriteString(fmt.Sprintf(`<tr>
  <td>%s</td>
  <td>%s</td>
  <td>%s</td>
  <td>%s</td>
  <td>%s</td>
  <td>%s</td>
  <td>%d min</td>
  <td>%d</td>
</tr>`+"\n",
			htmlEscape(r.LabID),
			htmlEscape(r.Title),
			htmlEscape(r.Category),
			htmlEscape(r.Difficulty),
			r.CompletedAt.Format("2006-01-02 15:04"),
			r.Duration.Round(time.Second),
			r.Estimated,
			r.Rating,
		))
	}

	template := `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>K8S-Lab-Everything Progress Report</title>
<style>
  body { font-family: -apple-system, "Segoe UI", sans-serif; margin: 40px auto; max-width: 960px; color:#222; }
  h1 { color:#1a73e8; }
  .stats { display:flex; gap:24px; margin:20px 0; }
  .stat { background:#f5f7fa; padding:16px 24px; border-radius:8px; }
  .stat b { display:block; font-size:28px; color:#1a73e8; }
  table { border-collapse:collapse; width:100%; margin-top:20px; }
  th, td { border:1px solid #ddd; padding:8px 12px; text-align:left; font-size:14px; }
  th { background:#1a73e8; color:#fff; }
  tr:nth-child(even) { background:#f9f9f9; }
</style>
</head>
<body>
  <h1>K8S-Lab-Everything Progress Report</h1>
  <div class="stats">
    <div class="stat"><b>__LABS__</b>Labs Completed</div>
    <div class="stat"><b>__TIME__</b>Total Practice Time</div>
    <div class="stat"><b>__STARTED__</b>Started</div>
  </div>
  <table>
    <thead><tr><th>ID</th><th>Title</th><th>Category</th><th>Difficulty</th><th>Completed</th><th>Duration</th><th>Est.</th><th>Rating</th></tr></thead>
    <tbody>
__ROWS__  </tbody>
  </table>
</body>
</html>
`
	replacer := strings.NewReplacer(
		"__LABS__", fmt.Sprintf("%d", total),
		"__TIME__", totalDuration.Round(time.Second).String(),
		"__STARTED__", p.StartedAt.Format("2006-01-02"),
		"__ROWS__", rows.String(),
	)
	return replacer.Replace(template), nil
}

func escapeCSV(s string) string {
	s = strings.ReplaceAll(s, `"`, `""`)
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + s + `"`
	}
	return s
}

func htmlEscape(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return replacer.Replace(s)
}

// SetActiveLab records the lab the user is currently working on (for `resume`).
func SetActiveLab(labID string) {
	p := Load()
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.LastActiveLab == labID {
		return
	}
	p.LastActiveLab = labID
	Save()
}

// ActiveLab returns the last lab the user was working on, or "" if none.
func ActiveLab() string {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.LastActiveLab
}

// RecordAttempt appends a verify attempt to a lab for benchmark tracking.
func RecordAttempt(labID string, passed bool, elapsed time.Duration) {
	p := Load()
	p.mu.Lock()
	defer p.mu.Unlock()

	r, ok := p.Labs[labID]
	if !ok {
		r = &LabResult{LabID: labID}
		p.Labs[labID] = r
	}
	r.Attempts = append(r.Attempts, Attempt{
		At:      time.Now(),
		Passed:  passed,
		Elapsed: elapsed,
	})
	_ = Save()
}

// BenchmarkSummary returns a human-readable summary of verify attempt times
// for completed labs (average vs best).
func BenchmarkSummary() string {
	p := Load()
	p.mu.RLock()
	defer p.mu.RUnlock()

	var done, passCount int
	var totalDur, worstDur time.Duration
	var worstID string

	for id, r := range p.Labs {
		if len(r.Attempts) == 0 {
			continue
		}
		var labTotal time.Duration
		for _, a := range r.Attempts {
			labTotal += a.Elapsed
			if a.Passed {
				passCount++
			}
			if a.Elapsed > worstDur {
				worstDur = a.Elapsed
				worstID = id
			}
		}
		totalDur += labTotal
		done++
	}

	if done == 0 {
		return "  No verify attempts recorded yet.\n"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("  %sBenchmarks%s (from %d lab(s) with attempt history)\n\n", bold, reset, done))
	b.WriteString(fmt.Sprintf("  Labs passed on a verify:        %s%d%s\n", green, passCount, reset))
	b.WriteString(fmt.Sprintf("  Total verify time tracked:       %s%s%s\n", dimWhite, totalDur.Round(time.Second), reset))
	b.WriteString(fmt.Sprintf("  Average verify time:             %s%s%s\n", yellow, (totalDur / time.Duration(done)).Round(time.Second), reset))
	b.WriteString(fmt.Sprintf("  Slowest single verify:           %s%s%s (%s)\n", red, worstDur.Round(time.Second), reset, worstID))
	b.WriteString("\n")
	return b.String()
}

// Notify sends a desktop notification (non-blocking). Falls back to console on
func Notify(title, message string) {
	switch {
	case isCommandAvailable("notify-send"):
		runSilent("notify-send", title, message)
	case isCommandAvailable("osascript"):
		script := fmt.Sprintf("display notification %q with title %q", message, title)
		runSilent("osascript", "-e", script)
	case isCommandAvailable("powershell"):
		runPowerShellNotify(title, message)
	default:
		fmt.Printf("\n  [%s✓%s] %s — %s\n", green, reset, title, message)
	}
}

func runPowerShellNotify(title, message string) {
	script := fmt.Sprintf(
		`Add-Type -AssemblyName System.Windows.Forms; `+
			`$n = New-Object System.Windows.Forms.NotifyIcon; `+
			`$n.Icon = [System.Drawing.SystemIcons]::Information; `+
			`$n.Visible = $true; `+
			`$n.ShowBalloonTip(3000, %q, %q, [System.Windows.Forms.ToolTipIcon]::Info); `+
			`Start-Sleep -Milliseconds 3500; $n.Dispose()`,
		title, message,
	)
	runSilent("powershell", "-NoProfile", "-Command", script)
}

func runSilent(name string, args ...string) {
	cmd := exec.Command(name, args...)
	_ = cmd.Start()
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
