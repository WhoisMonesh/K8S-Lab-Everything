package cli

import (
	"fmt"
	"strings"

	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/labs"
)

// PrintLabList prints a formatted list of labs
func PrintLabList(labList []labs.Lab) {
	if len(labList) == 0 {
		fmt.Println("No labs available.")
		return
	}

	fmt.Printf("%-25s %-40s %-18s %-10s\n", "ID", "Title", "Category", "Difficulty")
	fmt.Println(strings.Repeat("─", 95))

	for _, lab := range labList {
		info := labs.GetInfo(lab)
		fmt.Printf("%-25s %-40s %-18s %-10s\n",
			info.ID,
			truncate(info.Title, 38),
			info.Category,
			info.Difficulty,
		)
	}
}

// PrintLabDetails prints detailed information about a lab
func PrintLabDetails(lab labs.Lab) {
	fmt.Printf("\n")
	fmt.Printf("╔═══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║ Lab: %-60s ║\n", lab.Title())
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════╝\n")
	fmt.Printf("\n")
	fmt.Printf("ID:              %s\n", lab.ID())
	fmt.Printf("Category:        %s\n", lab.Category())
	fmt.Printf("Difficulty:      %s\n", lab.Difficulty())
	fmt.Printf("Estimated Time:  %d minutes\n", lab.EstimatedTime())

	tags := lab.Tags()
	if len(tags) > 0 {
		fmt.Printf("Tags:            %s\n", strings.Join(tags, ", "))
	}

	fmt.Printf("\n")
	fmt.Printf("Description:\n")
	fmt.Printf("%s\n", lab.Description())
	fmt.Printf("\n")

	hints := lab.Hints()
	if len(hints) > 0 {
		fmt.Printf("Hints:\n")
		for i, hint := range hints {
			fmt.Printf("  %d. %s\n", i+1, hint)
		}
		fmt.Printf("\n")
	}
}

// Success prints a success message
func Success(message string) {
	fmt.Printf("✓ %s\n", message)
}

// Error prints an error message
func Error(message string) {
	fmt.Printf("✗ %s\n", message)
}

// Info prints an info message
func Info(message string) {
	fmt.Printf("ℹ %s\n", message)
}

// Warning prints a warning message
func Warning(message string) {
	fmt.Printf("⚠ %s\n", message)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
