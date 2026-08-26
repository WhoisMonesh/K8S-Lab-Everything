package cli

import (
	"fmt"
	"strings"

	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/labs"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/progress"
)

const (
	reset       = "\033[0m"
	bold        = "\033[1m"
	dim         = "\033[2m"
	italic      = "\033[3m"
	underline   = "\033[4m"
	red         = "\033[31m"
	green       = "\033[32m"
	yellow      = "\033[33m"
	blue        = "\033[34m"
	magenta     = "\033[35m"
	cyan        = "\033[36m"
	white       = "\033[37m"
	brightRed   = "\033[91m"
	brightGreen = "\033[92m"
	brightCyan  = "\033[96m"
	dimWhite    = "\033[37;2m"
	dimCyan     = "\033[36;2m"
	dimGreen    = "\033[32;2m"
	dimYellow   = "\033[33;2m"
	dimRed      = "\033[31;2m"
	bgGreen     = "\033[42m"
	bgYellow    = "\033[43m"
	bgRed       = "\033[41m"
)

func badge(text, bg string) string {
	return bg + bold + " " + text + " " + reset
}

func diffBadge(d string) string {
	switch strings.ToLower(d) {
	case "easy":
		return badge("easy", bgGreen)
	case "medium":
		return badge("medium", bgYellow)
	case "hard":
		return badge("hard", bgRed)
	default:
		return d
	}
}

func diffColor(d string) string {
	switch strings.ToLower(d) {
	case "easy":
		return green
	case "medium":
		return yellow
	case "hard":
		return red
	default:
		return white
	}
}

func catColor(c string) string {
	switch strings.ToLower(c) {
	case "control-plane":
		return cyan
	case "networking":
		return blue
	case "scheduling":
		return magenta
	case "dns":
		return yellow
	case "storage":
		return white
	case "security":
		return red
	case "rbac":
		return green
	case "workloads":
		return brightCyan
	default:
		return white
	}
}

func catIcon(c string) string {
	switch strings.ToLower(c) {
	case "control-plane":
		return "◆"
	case "networking":
		return "◆"
	case "scheduling":
		return "◆"
	case "dns":
		return "◆"
	case "storage":
		return "◆"
	case "security":
		return "◆"
	case "rbac":
		return "◆"
	case "workloads":
		return "◆"
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
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("%s%s%s %s%d/%d%s (%d%%)",
		green, bar, reset, dimWhite, completed, total, reset, pct)
}

func PrintBanner() {
	fmt.Println()
	fmt.Printf("  %s%s╔═══════════════════════════════════════════════════════════════════╗%s\n", dim, cyan, reset)
	fmt.Printf("  %s%s║%s                                                                       %s%s║%s\n", dim, cyan, reset, dim, cyan, reset)
	fmt.Printf("  %s%s║%s     %s██╗  ██╗██╗      ██████╗ ███╗   ██╗███████╗██╗    ██╗ %s    %s%s║%s\n", dim, cyan, reset, bold, reset, dim, cyan, reset)
	fmt.Printf("  %s%s║%s     %s██║  ██║██║     ██╔═══██╗████╗  ██║██╔════╝██║    ██║ %s    %s%s║%s\n", dim, cyan, reset, bold, reset, dim, cyan, reset)
	fmt.Printf("  %s%s║%s     %s███████║██║     ██║   ██║██╔██╗ ██║█████╗  ██║ █╗ ██║ %s    %s%s║%s\n", dim, cyan, reset, bold, reset, dim, cyan, reset)
	fmt.Printf("  %s%s║%s     %s██╔══██║██║     ██║   ██║██║╚██╗██║██╔══╝  ██║███╗██║ %s    %s%s║%s\n", dim, cyan, reset, bold, reset, dim, cyan, reset)
	fmt.Printf("  %s%s║%s     %s██║  ██║███████╗╚██████╔╝██║ ╚████║███████╗╚███╔███╔╝ %s    %s%s║%s\n", dim, cyan, reset, bold, reset, dim, cyan, reset)
	fmt.Printf("  %s%s║%s     %s╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝ ╚══╝╚══╝  %s    %s%s║%s\n", dim, cyan, reset, bold, reset, dim, cyan, reset)
	fmt.Printf("  %s%s║%s                      %sLab-Everything%s                         %s%s║%s\n", dim, cyan, reset, brightCyan, reset, dim, cyan, reset)
	fmt.Printf("  %s%s║%s                                                                       %s%s║%s\n", dim, cyan, reset, dim, cyan, reset)
	fmt.Printf("  %s%s╚═══════════════════════════════════════════════════════════════════╝%s\n", dim, cyan, reset)
	fmt.Println()
	fmt.Printf("  %sPractice Kubernetes troubleshooting with broken scenarios%s\n", dimWhite, reset)
	fmt.Printf("  %sReal kubectl. Real problems. Real solutions.%s\n", dimWhite, reset)
	fmt.Println()
}

func PrintLabList(labList []labs.Lab) {
	PrintLabListWithProgress(labList, false)
}

func PrintLabListWithProgress(labList []labs.Lab, showProgress bool) {
	if len(labList) == 0 {
		fmt.Printf("\n  %s%sNo labs available.%s\n\n", dim, italic, reset)
		return
	}

	fmt.Println()

	if showProgress {
		completed := progress.CompletedCount()
		total := len(labList)
		fmt.Printf("  %sProgress%s  %s\n\n", bold, reset, progressBar(completed, total, 40))
	}

	// Group by category
	grouped := make(map[string][]labs.Lab)
	order := []string{"control-plane", "workloads", "networking", "scheduling", "dns", "storage", "security", "rbac"}
	seen := make(map[string]bool)

	for _, lab := range labList {
		cat := string(lab.Category())
		grouped[cat] = append(grouped[cat], lab)
		if !seen[cat] {
			order = append(order, cat)
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
		fmt.Printf("  %s%s %s %s(%d)%s\n", color, icon, strings.ToUpper(cat), dimWhite, len(catLabs), reset)
		fmt.Printf("  %s%s%s\n", dim, strings.Repeat("─", 70), reset)

		for _, lab := range catLabs {
			info := labs.GetInfo(lab)
			check := "  "
			if showProgress && progress.IsCompleted(info.ID) {
				check = fmt.Sprintf("%s✔%s", green, reset)
			}
			fmt.Printf("  %s  %-28s  %s%-36s  %s%s\n",
				check,
				dimWhite+info.ID+reset,
				bold,
				truncate(info.Title, 34),
				reset,
				diffBadge(string(info.Difficulty)),
			)
		}
		fmt.Println()
	}

	fmt.Printf("  %s%d lab(s) listed%s\n\n", dim, len(labList), reset)
}

func PrintLabDetails(lab labs.Lab) {
	w := 62
	fmt.Println()
	fmt.Printf("  %s┌%s%s┐%s\n", cyan, strings.Repeat("─", w), cyan, reset)
	fmt.Printf("  %s│%s  %s%-*s%s  %s│%s\n", cyan, reset, bold, w-4, lab.Title(), reset, cyan, reset)
	fmt.Printf("  %s└%s%s┘%s\n", cyan, strings.Repeat("─", w), cyan, reset)
	fmt.Println()

	fmt.Printf("  %s%-16s%s %s%s%s\n", bold, "ID", reset, dimWhite, lab.ID(), reset)
	fmt.Printf("  %s%-16s%s %s\n", bold, "Category", reset, catColor(string(lab.Category()))+string(lab.Category())+reset)
	fmt.Printf("  %s%-16s%s %s\n", bold, "Difficulty", reset, diffBadge(string(lab.Difficulty())))
	fmt.Printf("  %s%-16s%s %s%d min%s\n", bold, "Est. Time", reset, dimWhite, lab.EstimatedTime(), reset)

	if domain := labs.GetDomain(lab); domain != "" {
		fmt.Printf("  %s%-16s%s %s%s%s\n", bold, "CKA Domain", reset, magenta, domain, reset)
	}

	if prereqs := labs.GetPrerequisites(lab); len(prereqs) > 0 {
		fmt.Printf("  %s%-16s%s %s%s%s\n", bold, "Prerequisites", reset, yellow, strings.Join(prereqs, ", "), reset)
	}

	tags := lab.Tags()
	if len(tags) > 0 {
		fmt.Printf("  %s%-16s%s %s%s%s\n", bold, "Tags", reset, dim, strings.Join(tags, "  "), reset)
	}

	if progress.IsCompleted(lab.ID()) {
		fmt.Printf("  %s%-16s%s %s✔ COMPLETED%s\n", bold, "Status", reset, green, reset)
	}

	fmt.Println()
	fmt.Printf("  %s%sDescription%s\n", bold, underline, reset)
	fmt.Println()
	for _, line := range strings.Split(lab.Description(), "\n") {
		if strings.TrimSpace(line) == "" {
			fmt.Println()
		} else {
			fmt.Printf("  %s│%s %s\n", dim, reset, line)
		}
	}
	fmt.Println()

	hints := lab.Hints()
	if len(hints) > 0 {
		fmt.Printf("  %s%sHints%s\n", bold, underline, reset)
		fmt.Println()
		for i, hint := range hints {
			fmt.Printf("  %s  %d.%s  %s▸%s %s%s%s\n", dim, i+1, reset, yellow, reset, italic, hint, reset)
		}
		fmt.Println()
	}
}

func Success(message string) {
	fmt.Printf("\n  %s✔%s  %s%s%s\n", green, reset, bold, message, reset)
}

func Error(message string) {
	fmt.Printf("\n  %s✖%s  %s%s%s\n", brightRed, reset, bold, message, reset)
}

func Info(message string) {
	fmt.Printf("\n  %s▸%s  %s%s%s\n", brightCyan, reset, dimWhite, message, reset)
}

func Warning(message string) {
	fmt.Printf("\n  %s⚠%s  %s%s%s\n", yellow, reset, dimYellow, message, reset)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
