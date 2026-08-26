package cli

import (
	"fmt"
	"strings"

	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/labs"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/progress"
)

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
)

func colorDiff(difficulty string) string {
	switch strings.ToLower(difficulty) {
	case "easy":
		return colorGreen + "easy" + colorReset
	case "medium":
		return colorYellow + "medium" + colorReset
	case "hard":
		return colorRed + "hard" + colorReset
	default:
		return difficulty
	}
}

func colorCategory(cat string) string {
	switch strings.ToLower(cat) {
	case "control-plane":
		return colorCyan + cat + colorReset
	case "networking":
		return colorBlue + cat + colorReset
	case "scheduling":
		return colorMagenta + cat + colorReset
	case "dns":
		return colorYellow + cat + colorReset
	case "storage":
		return colorWhite + cat + colorReset
	case "security":
		return colorRed + cat + colorReset
	case "rbac":
		return colorGreen + cat + colorReset
	case "workloads":
		return colorCyan + cat + colorReset
	default:
		return cat
	}
}

func PrintLabList(labList []labs.Lab) {
	PrintLabListWithProgress(labList, false)
}

func PrintLabListWithProgress(labList []labs.Lab, showProgress bool) {
	if len(labList) == 0 {
		fmt.Println("No labs available.")
		return
	}

	fmt.Println()
	if showProgress {
		completed := progress.CompletedCount()
		total := len(labList)
		pct := 0
		if total > 0 {
			pct = completed * 100 / total
		}
		barWidth := 30
		filled := pct * barWidth / 100
		bar := strings.Repeat("=", filled) + strings.Repeat("-", barWidth-filled)

		fmt.Printf("  %sProgress:%s %s[%s]%s %d/%d (%d%%)\n",
			colorBold, colorReset, colorGreen, bar, colorReset, completed, total, pct)
		fmt.Println()
	}

	fmt.Printf("  %s%-3s  %-28s  %-38s  %-16s  %-10s%s\n",
		colorBold, "ID", "Title", "Category", "Difficulty", "", colorReset)
	fmt.Printf("  %s%s%s\n", colorDim, strings.Repeat("─", 98), colorReset)

	for _, lab := range labList {
		info := labs.GetInfo(lab)
		marker := "  "
		if showProgress && progress.IsCompleted(info.ID) {
			marker = colorGreen + "+" + colorReset
		}
		fmt.Printf("  %s%-3s  %s%-28s  %s%-38s  %s  %s%s\n",
			marker,
			info.ID,
			colorBold,
			truncate(info.Title, 26),
			colorReset,
			colorCategory(string(info.Category)),
			colorDiff(string(info.Difficulty)),
			"",
			colorReset)
	}

	fmt.Printf("\n  %s%d lab(s) listed%s\n", colorDim, len(labList), colorReset)
	fmt.Println()
}

func PrintLabDetails(lab labs.Lab) {
	fmt.Println()
	fmt.Printf("  %s╔══════════════════════════════════════════════════════════════╗%s\n", colorCyan, colorReset)
	fmt.Printf("  %s║%s  %s%-56s%s  %s║%s\n", colorCyan, colorReset, colorBold, lab.Title(), colorReset, colorCyan, colorReset)
	fmt.Printf("  %s╚══════════════════════════════════════════════════════════════╝%s\n", colorCyan, colorReset)
	fmt.Println()

	fmt.Printf("  %sID:%s            %s%s%s\n", colorBold, colorReset, colorWhite, lab.ID(), colorReset)
	fmt.Printf("  %sCategory:%s      %s\n", colorBold, colorReset, colorCategory(string(lab.Category())))
	fmt.Printf("  %sDifficulty:%s    %s\n", colorBold, colorReset, colorDiff(string(lab.Difficulty())))
	fmt.Printf("  %sEst. Time:%s     %s%d min%s\n", colorBold, colorReset, colorWhite, lab.EstimatedTime(), colorReset)

	if domain := labs.GetDomain(lab); domain != "" {
		fmt.Printf("  %sCKA Domain:%s    %s%s%s\n", colorBold, colorReset, colorMagenta, domain, colorReset)
	}

	if prereqs := labs.GetPrerequisites(lab); len(prereqs) > 0 {
		fmt.Printf("  %sPrerequisites:%s %s%s%s\n", colorBold, colorReset, colorYellow, strings.Join(prereqs, ", "), colorReset)
	}

	tags := lab.Tags()
	if len(tags) > 0 {
		fmt.Printf("  %sTags:%s          %s%s%s\n", colorBold, colorReset, colorDim, strings.Join(tags, ", "), colorReset)
	}

	if progress.IsCompleted(lab.ID()) {
		fmt.Printf("  %sStatus:%s        %sCOMPLETED%s\n", colorBold, colorReset, colorGreen, colorReset)
	}

	fmt.Println()
	fmt.Printf("  %sDescription:%s\n", colorBold, colorReset)
	for _, line := range strings.Split(lab.Description(), "\n") {
		fmt.Printf("    %s%s%s\n", colorWhite, line, colorReset)
	}
	fmt.Println()

	hints := lab.Hints()
	if len(hints) > 0 {
		fmt.Printf("  %sHints:%s\n", colorBold, colorReset)
		for i, hint := range hints {
			fmt.Printf("    %s%d.%s %s%s%s\n", colorDim, i+1, colorReset, colorYellow, hint, colorReset)
		}
		fmt.Println()
	}
}

func Success(message string) {
	fmt.Printf("\n  %s✓  %s%s\n", colorGreen, message, colorReset)
}

func Error(message string) {
	fmt.Printf("\n  %s✗  %s%s\n", colorRed, message, colorReset)
}

func Info(message string) {
	fmt.Printf("\n  %s●  %s%s\n", colorCyan, message, colorReset)
}

func Warning(message string) {
	fmt.Printf("\n  %s⚠  %s%s\n", colorYellow, message, colorReset)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
