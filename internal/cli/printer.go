package cli

import (
	"fmt"
	"strings"

	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/labs"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/progress"
)

func PrintLabList(labList []labs.Lab) {
	PrintLabListWithProgress(labList, false)
}

func PrintLabListWithProgress(labList []labs.Lab, showProgress bool) {
	if len(labList) == 0 {
		fmt.Println("No labs available.")
		return
	}

	if showProgress {
		fmt.Printf("%-4s %-25s %-40s %-18s %-10s\n", "", "ID", "Title", "Category", "Difficulty")
		fmt.Println(strings.Repeat("─", 100))

		for _, lab := range labList {
			info := labs.GetInfo(lab)
			marker := " "
			if progress.IsCompleted(info.ID) {
				marker = "+"
			}
			fmt.Printf("[%-2s] %-25s %-40s %-18s %-10s\n",
				marker,
				info.ID,
				truncate(info.Title, 38),
				info.Category,
				info.Difficulty,
			)
		}
		completed := progress.CompletedCount()
		fmt.Printf("\n  %d/%d labs completed\n", completed, len(labList))
	} else {
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
}

func PrintLabDetails(lab labs.Lab) {
	fmt.Printf("\n")
	fmt.Printf("=============================================================\n")
	fmt.Printf(" Lab: %s\n", lab.Title())
	fmt.Printf("=============================================================\n")
	fmt.Printf("\n")
	fmt.Printf("ID:              %s\n", lab.ID())
	fmt.Printf("Category:        %s\n", lab.Category())
	fmt.Printf("Difficulty:      %s\n", lab.Difficulty())
	fmt.Printf("Estimated Time:  %d minutes\n", lab.EstimatedTime())

	if domain := labs.GetDomain(lab); domain != "" {
		fmt.Printf("CKA Domain:      %s\n", domain)
	}

	if prereqs := labs.GetPrerequisites(lab); len(prereqs) > 0 {
		fmt.Printf("Prerequisites:   %s\n", strings.Join(prereqs, ", "))
	}

	tags := lab.Tags()
	if len(tags) > 0 {
		fmt.Printf("Tags:            %s\n", strings.Join(tags, ", "))
	}

	if progress.IsCompleted(lab.ID()) {
		fmt.Printf("Status:          COMPLETED\n")
	}

	fmt.Printf("\n")
	fmt.Printf("Description:\n")
	fmt.Printf("%s\n", lab.Description())
	fmt.Printf("\n")

	hints := lab.Hints()
	if len(hints) > 0 {
		fmt.Printf("Hints (use 'lab hint %s --level N'):\n", lab.ID())
		for i, hint := range hints {
			fmt.Printf("  %d. %s\n", i+1, hint)
		}
		fmt.Printf("\n")
	}
}

func Success(message string) {
	fmt.Printf("[OK] %s\n", message)
}

func Error(message string) {
	fmt.Printf("[ERROR] %s\n", message)
}

func Info(message string) {
	fmt.Printf("[INFO] %s\n", message)
}

func Warning(message string) {
	fmt.Printf("[WARN] %s\n", message)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
