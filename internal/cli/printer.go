package cli

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/labs"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/progress"
)

const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	dim       = "\033[2m"
	italic    = "\033[3m"
	underline = "\033[4m"
	red       = "\033[31m"
	green     = "\033[32m"
	yellow    = "\033[33m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	white     = "\033[37m"
	brRed     = "\033[91m"
	brGreen   = "\033[92m"
	brYellow  = "\033[93m"
	brBlue    = "\033[94m"
	brMagenta = "\033[95m"
	brCyan    = "\033[96m"
	brWhite   = "\033[97m"
	dimW      = "\033[90m"
	bgRed     = "\033[41m"
	bgGreen   = "\033[42m"
	bgYellow  = "\033[43m"
)

func padRight(s string, width int) string {
	w := utf8.RuneCountInString(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func diffTag(d string) string {
	switch strings.ToLower(d) {
	case "easy":
		return bgGreen + bold + " EASY " + reset
	case "medium":
		return bgYellow + bold + " MEDIUM " + reset
	case "hard":
		return bgRed + bold + " HARD " + reset
	default:
		return d
	}
}

func diffColor(d string) string {
	switch strings.ToLower(d) {
	case "easy":
		return brGreen
	case "medium":
		return brYellow
	case "hard":
		return brRed
	default:
		return white
	}
}

func catColor(c string) string {
	switch strings.ToLower(c) {
	case "control-plane":
		return brCyan
	case "networking":
		return brBlue
	case "scheduling":
		return brMagenta
	case "dns":
		return brYellow
	case "storage":
		return brWhite
	case "security":
		return brRed
	case "rbac":
		return brGreen
	case "workloads":
		return cyan
	default:
		return white
	}
}

func catIcon(c string) string {
	switch strings.ToLower(c) {
	case "control-plane":
		return "⚙"
	case "networking":
		return "⛓"
	case "scheduling":
		return "📅"
	case "dns":
		return "🔍"
	case "storage":
		return "💾"
	case "security":
		return "🔒"
	case "rbac":
		return "🔑"
	case "workloads":
		return "📦"
	default:
		return "◆"
	}
}

func progressBar(completed, total, width int) string {
	pct := 0
	if total > 0 {
		pct = completed * 100 / total
	}
	filled := pct * width / 100
	empty := width - filled
	bar := brGreen + strings.Repeat("━", filled) + dimW + strings.Repeat("─", empty) + reset
	return fmt.Sprintf("%s %s%d/%d%s", bar, brWhite, completed, total, reset)
}

func PrintBanner() {
	fmt.Println()
	lines := []string{
		"  █████   ████  ████████    █████████             █████                 █████                ██████████                                            █████    █████       ███",
		"  ▒███  ███   ▒███   ▒███ ▒███    ▒▒▒             ▒███         ██████   ▒███████             ▒███  █ ▒  █████ █████  ██████  ████████  █████ ████ ███████   ▒███████   ████",
		"  ▒███████    ▒▒████████  ▒▒█████████  ██████████ ▒███        ▒▒▒▒▒███  ▒███▒▒███ ██████████ ▒██████   ▒▒███ ▒▒███  ███▒▒███▒▒███▒▒███▒▒███ ▒███ ▒▒▒███▒    ▒███▒▒███ ▒▒███",
		"  ▒███▒▒███    ███▒▒▒▒███  ▒▒▒▒▒▒▒▒███▒▒▒▒▒▒▒▒▒▒  ▒███         ███████  ▒███ ▒███▒▒▒▒▒▒▒▒▒▒  ▒███▒▒█    ▒███  ▒███ ▒███████  ▒███ ▒▒▒  ▒███ ▒███   ▒███     ▒███ ▒███  ▒███",
		"  ▒███ ▒▒███  ▒███   ▒███  ███    ▒███            ▒███      █ ███▒▒███  ▒███ ▒███            ▒███ ▒   █ ▒▒███ ███  ▒███▒▒▒   ▒███      ▒███ ▒███   ▒███ ███ ▒███ ▒███  ▒███",
		"  █████ ▒▒████▒▒████████  ▒▒█████████             ███████████▒▒████████ ████████             ██████████  ▒▒█████   ▒▒██████  █████     ▒▒███████   ▒▒█████  ████ █████ █████",
		"  ▒▒▒▒▒   ▒▒▒▒  ▒▒▒▒▒▒▒▒    ▒▒▒▒▒▒▒▒▒             ▒▒▒▒▒▒▒▒▒▒▒  ▒▒▒▒▒▒▒▒ ▒▒▒▒▒▒▒▒             ▒▒▒▒▒▒▒▒▒▒    ▒▒▒▒▒     ▒▒▒▒▒▒  ▒▒▒▒▒       ▒▒▒▒▒███    ▒▒▒▒▒  ▒▒▒▒ ▒▒▒▒▒",
	}

	for _, line := range lines {
		fmt.Printf("  %s%s%s%s\n", dim, cyan, line, reset)
	}

	fmt.Println()
	fmt.Printf("  %s%s   ██ K8S-Lab-Everything ██%s\n", bold, brGreen, reset)
	fmt.Println()
	fmt.Printf("  %s143 → 203 Hands-On Labs  │  Kubernetes Troubleshooting  │  Interactive TUI%s\n", dimW, reset)
	fmt.Println()
	fmt.Printf("  %s▸%s Run %scka-lab-runner lab pick%s to select a lab interactively\n", brCyan, reset, bold, reset)
	fmt.Printf("  %s▸%s Run %scka-lab-runner lab list%s to see all available labs\n\n", brCyan, reset, bold, reset)
}

func PrintLabList(labList []labs.Lab) {
	PrintLabListWithProgress(labList, false)
}

func PrintLabListWithProgress(labList []labs.Lab, showProgress bool) {
	if len(labList) == 0 {
		fmt.Printf("\n  %sNo labs available.%s\n\n", dim, reset)
		return
	}

	fmt.Println()

	if showProgress {
		completed := progress.CompletedCount()
		total := len(labList)
		fmt.Printf("  %s%sProgress%s\n", bold, reset, reset)
		fmt.Printf("  %s\n\n", progressBar(completed, total, 40))
	}

	grouped := make(map[string][]labs.Lab)
	order := []string{"control-plane", "workloads", "networking", "scheduling", "dns", "storage", "security", "rbac"}
	seen := make(map[string]bool)

	for _, lab := range labList {
		cat := string(lab.Category())
		grouped[cat] = append(grouped[cat], lab)
		if !seen[cat] {
			if !contains(order, cat) {
				order = append(order, cat)
			}
			seen[cat] = true
		}
	}

	for _, cat := range order {
		catLabs, ok := grouped[cat]
		if !ok || len(catLabs) == 0 {
			continue
		}

		color := catColor(cat)
		icon := catIcon(cat)
		fmt.Printf("  %s%s  %s %s%s  %s%s%d labs%s\n", color, icon, bold, strings.ToUpper(cat), reset, dimW, reset, len(catLabs), reset)
		fmt.Printf("  %s%s\n", dim, strings.Repeat("┄", 66))

		for _, lab := range catLabs {
			info := labs.GetInfo(lab)
			check := "  "
			if showProgress && progress.IsCompleted(info.ID) {
				check = fmt.Sprintf("%s✔%s", brGreen, reset)
			}

			title := padRight(truncate(info.Title, 34), 36)
			id := padRight(info.ID, 30)

			fmt.Printf("  %s  %s%s%s  %s%s  %s%s%s\n",
				check,
				dimW, id, reset,
				title, reset,
				"", diffTag(string(info.Difficulty)), reset,
			)
		}
		fmt.Println()
	}

	diffCounts := make(map[string]int)
	for _, lab := range labList {
		diffCounts[string(lab.Difficulty())]++
	}

	fmt.Printf("  %s%s──────────────────────────────────────────────────────%s\n", dim, white, reset)
	fmt.Printf("  %sTotal:%s %s%d labs%s", bold, reset, brWhite, len(labList), reset)
	if c, ok := diffCounts["easy"]; ok {
		fmt.Printf("  %s● %d easy%s", brGreen, c, reset)
	}
	if c, ok := diffCounts["medium"]; ok {
		fmt.Printf("  %s● %d medium%s", brYellow, c, reset)
	}
	if c, ok := diffCounts["hard"]; ok {
		fmt.Printf("  %s● %d hard%s", brRed, c, reset)
	}
	fmt.Println()
}

func PrintLabDetails(lab labs.Lab) {
	w := 60

	fmt.Println()
	fmt.Printf("  %s┌%s%s┐%s\n", cyan, strings.Repeat("─", w), cyan, reset)
	fmt.Printf("  %s│%s  %s%-*s%s  %s│%s\n", cyan, reset, bold, w-4, lab.Title(), reset, cyan, reset)
	fmt.Printf("  %s└%s%s┘%s\n", cyan, strings.Repeat("─", w), cyan, reset)
	fmt.Println()

	fmt.Printf("  %s%s▸ Details%s\n", bold, cyan, reset)
	fmt.Println()
	fmt.Printf("  %s  ID %s│%s  %s%s%s\n", dimW, dim, reset, bold, lab.ID(), reset)
	fmt.Printf("  %s  Category %s│%s  %s%s%s\n", dimW, dim, reset, bold, catColor(string(lab.Category()))+strings.ToUpper(string(lab.Category())), reset)
	fmt.Printf("  %s  Difficulty %s│%s  %s%s\n", dimW, dim, reset, diffTag(string(lab.Difficulty())), reset)
	fmt.Printf("  %s  Est. Time %s│%s  %s%d min%s\n", dimW, dim, reset, brWhite, lab.EstimatedTime(), reset)

	if domain := labs.GetDomain(lab); domain != "" {
		fmt.Printf("  %s  CKA Domain %s│%s  %s%s%s\n", dimW, dim, reset, magenta, domain, reset)
	}

	if prereqs := labs.GetPrerequisites(lab); len(prereqs) > 0 {
		fmt.Printf("  %s  Prereqs %s│%s  %s%s%s\n", dimW, dim, reset, yellow, strings.Join(prereqs, ", "), reset)
	}

	tags := lab.Tags()
	if len(tags) > 0 {
		fmt.Printf("  %s  Tags %s│%s  %s%s%s\n", dimW, dim, reset, dim, strings.Join(tags, "  "), reset)
	}

	if progress.IsCompleted(lab.ID()) {
		fmt.Printf("  %s  Status %s│%s  %s%s✔ COMPLETED%s\n", dimW, dim, reset, brGreen, bold, reset)
	}

	fmt.Println()
	fmt.Printf("  %s%s▸ Description%s\n", bold, cyan, reset)
	fmt.Println()
	for _, line := range strings.Split(lab.Description(), "\n") {
		if strings.TrimSpace(line) == "" {
			fmt.Println()
		} else {
			fmt.Printf("  %s  │%s %s\n", dim, reset, line)
		}
	}
	fmt.Println()

	hints := lab.Hints()
	if len(hints) > 0 {
		fmt.Printf("  %s%s▸ Hints%s\n", bold, cyan, reset)
		fmt.Println()
		for i, hint := range hints {
			fmt.Printf("  %s  %d.%s  %s▸%s %s\n", dim, i+1, reset, yellow, reset, hint)
		}
		fmt.Println()
	}

	fmt.Printf("  %s%s▸ Commands%s\n", bold, cyan, reset)
	fmt.Println()
	fmt.Printf("  %s  cka-lab-runner lab run %s       %s# Apply the broken scenario%s\n", dimW, lab.ID(), dim, reset)
	fmt.Printf("  %s  cka-lab-runner lab verify %s   %s# Check your fix%s\n", dimW, lab.ID(), dim, reset)
	fmt.Printf("  %s  cka-lab-runner lab hint %s     %s# Get help%s\n", dimW, lab.ID(), dim, reset)
	fmt.Printf("  %s  cka-lab-runner lab solution %s %s# Show full solution%s\n", dimW, lab.ID(), dim, reset)
	fmt.Println()
}

func Success(message string) {
	fmt.Printf("\n  %s%s✔%s  %s%s%s\n", bold, brGreen, reset, bold, message, reset)
}

func Error(message string) {
	fmt.Printf("\n  %s%s✖%s  %s%s%s\n", bold, brRed, reset, bold, message, reset)
}

func Info(message string) {
	fmt.Printf("\n  %s▸%s  %s\n", brCyan, reset, message)
}

func Warning(message string) {
	fmt.Printf("\n  %s%s⚠%s  %s%s%s\n", bold, brYellow, reset, dim, message, reset)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
