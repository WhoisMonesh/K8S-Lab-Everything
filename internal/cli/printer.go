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
	bgBlue    = "\033[44m"
	bgMagenta = "\033[45m"
	bgCyan    = "\033[46m"
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

func certBadge(cert labs.Cert) string {
	switch cert {
	case labs.CertCKA:
		return bgBlue + bold + " CKA " + reset
	case labs.CertCKAD:
		return bgCyan + bold + " CKAD " + reset
	case labs.CertCKS:
		return bgMagenta + bold + " CKS " + reset
	default:
		return ""
	}
}

func catColor(c string) string {
	switch strings.ToLower(c) {
	// CKA domains
	case "cluster-architecture":
		return brCyan
	case "workloads-scheduling":
		return brBlue
	case "services-networking":
		return brMagenta
	case "storage":
		return brWhite
	case "troubleshooting":
		return brRed
	// CKAD domains
	case "app-design-build":
		return cyan
	case "app-deployment":
		return brGreen
	case "app-observability":
		return brYellow
	case "app-config-security":
		return magenta
	case "services-networking-ckad":
		return brBlue
	// CKS domains
	case "cluster-setup-cks":
		return brCyan
	case "cluster-hardening":
		return brGreen
	case "system-hardening":
		return brYellow
	case "microservice-vulns":
		return brRed
	case "supply-chain":
		return white
	case "monitoring-logging":
		return brMagenta
	// Legacy aliases
	case "control-plane":
		return brCyan
	case "networking":
		return brBlue
	case "scheduling":
		return brMagenta
	case "dns":
		return brYellow
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
	// CKA domains
	case "cluster-architecture":
		return "⚙"
	case "workloads-scheduling":
		return "📦"
	case "services-networking":
		return "⛓"
	case "storage":
		return "💾"
	case "troubleshooting":
		return "🔍"
	// CKAD domains
	case "app-design-build":
		return "🏗"
	case "app-deployment":
		return "🚀"
	case "app-observability":
		return "📊"
	case "app-config-security":
		return "🔒"
	case "services-networking-ckad":
		return "🌐"
	// CKS domains
	case "cluster-setup-cks":
		return "🛡"
	case "cluster-hardening":
		return "🔐"
	case "system-hardening":
		return "🖥"
	case "microservice-vulns":
		return "🐛"
	case "supply-chain":
		return "📦"
	case "monitoring-logging":
		return "📋"
	// Legacy aliases
	case "control-plane":
		return "⚙"
	case "networking":
		return "⛓"
	case "scheduling":
		return "📅"
	case "dns":
		return "🔍"
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
	fmt.Printf("  %sCKA%s │ %sCKAD%s │ %sCKS%s  │  378 Hands-On Labs  │  Interactive TUI\n",
		bold+bgBlue, reset, bold+bgCyan, reset, bold+bgMagenta, reset)
	fmt.Println()
	fmt.Printf("  %s▸%s Run %scka-lab-runner lab list%s to see all available labs\n", brCyan, reset, bold, reset)
	fmt.Printf("  %s▸%s Run %scka-lab-runner lab list --cert CKA%s to filter by certification\n", brCyan, reset, bold, reset)
	fmt.Printf("  %s▸%s Run %scka-lab-runner lab pick%s to select a lab interactively\n\n", brCyan, reset, bold, reset)
}

func PrintLabList(labList []labs.Lab) {
	PrintLabListWithProgress(labList, false)
}

func PrintLabListWithProgress(labList []labs.Lab, showProgress bool) {
	if len(labList) == 0 {
		fmt.Printf("\n  %sNo labs available.%s\n\n", dim, reset)
		return
	}

	w := 90

	fmt.Println()
	fmt.Printf("  %s┌%s%s┐%s\n", cyan, strings.Repeat("─", w-2), cyan, reset)
	title := "K8S-Lab-Everything — CKA │ CKAD │ CKS Labs"
	pad := (w - 2 - len(title)) / 2
	leftPad := strings.Repeat(" ", pad)
	rightPad := strings.Repeat(" ", w-2-pad-len(title))
	fmt.Printf("  %s│%s%s%s%s%s%s%s│%s\n",
		cyan, reset, leftPad, bold, brWhite, title, reset, rightPad, cyan)
	fmt.Printf("  %s└%s%s┘%s\n", cyan, strings.Repeat("─", w-2), cyan, reset)
	fmt.Println()

	if showProgress {
		completed := progress.CompletedCount()
		total := len(labList)
		fmt.Printf("  %s%sProgress%s\n", bold, reset, reset)
		fmt.Printf("  %s\n\n", progressBar(completed, total, 40))
	}

	// Group by cert, then by category
	certOrder := []labs.Cert{labs.CertCKA, labs.CertCKAD, labs.CertCKS}
	certLabels := map[labs.Cert]string{
		labs.CertCKA:  "CKA — Certified Kubernetes Administrator",
		labs.CertCKAD: "CKAD — Certified Kubernetes Application Developer",
		labs.CertCKS:  "CKS — Certified Kubernetes Security Specialist",
	}
	certColors := map[labs.Cert]string{
		labs.CertCKA:  bgBlue,
		labs.CertCKAD: bgCyan,
		labs.CertCKS:  bgMagenta,
	}

	// Group labs by cert then category
	type certGroup struct {
		cert       labs.Cert
		categories map[string][]labs.Lab
		order      []string
	}
	groups := map[labs.Cert]*certGroup{}
	for _, cert := range certOrder {
		groups[cert] = &certGroup{cert: cert, categories: make(map[string][]labs.Lab)}
	}

	for _, lab := range labList {
		cert := labs.GetCert(lab)
		if cert == labs.CertAll {
			continue
		}
		g, ok := groups[cert]
		if !ok {
			continue
		}
		cat := string(lab.Category())
		g.categories[cat] = append(g.categories[cat], lab)
		if !contains(g.order, cat) {
			g.order = append(g.order, cat)
		}
	}

	for _, cert := range certOrder {
		g := groups[cert]
		if len(g.categories) == 0 {
			continue
		}

		certColor := certColors[cert]

		// Count labs for this cert
		totalCert := 0
		for _, catLabs := range g.categories {
			totalCert += len(catLabs)
		}

		fmt.Printf("  %s┌%s%s┐%s\n", certColor, strings.Repeat("─", w-2), certColor, reset)
		certHeader := fmt.Sprintf("%s  %s  %s", certBadge(cert), certLabels[cert], fmt.Sprintf("(%d labs)", totalCert))
		fmt.Printf("  %s│%s  %s%s%-*s%s%s│%s\n",
			certColor, reset, certColor, bold, w-6, certHeader, reset, certColor, reset)
		fmt.Printf("  %s├%s%s┤%s\n", certColor, strings.Repeat("─", w-2), certColor, reset)

		for _, cat := range g.order {
			catLabs := g.categories[cat]
			if len(catLabs) == 0 {
				continue
			}

			color := catColor(cat)
			icon := catIcon(cat)
			weight := labs.DomainWeightForCategory(labs.Category(cat))

			fmt.Printf("  %s│%s  %s%s %s %s %s%s(%d labs, %d%%%s)%s│%s\n",
				certColor, reset, color, bold, icon,
				strings.ToUpper(cat), reset,
				dimW, len(catLabs), weight, reset, certColor, reset)

			for _, lab := range catLabs {
				info := labs.GetInfo(lab)
				check := "  "
				if showProgress && progress.IsCompleted(info.ID) {
					check = fmt.Sprintf("%s✔%s", brGreen, reset)
				}

				diffBadge := diffTag(string(info.Difficulty))

				idStr := padRight(info.ID, 28)
				titleStr := padRight(truncate(info.Title, 30), 32)

				fmt.Printf("  %s│%s %s%s%s  %s%s%s  %s%s%s  %s%s│%s\n",
					certColor, reset,
					check, "", "",
					dimW, idStr, reset,
					brWhite, titleStr, reset,
					diffBadge, "", reset)
			}
		}

		fmt.Printf("  %s└%s%s┘%s\n", certColor, strings.Repeat("─", w-2), certColor, reset)
		fmt.Println()
	}

	diffCounts := make(map[string]int)
	for _, lab := range labList {
		diffCounts[string(lab.Difficulty())]++
	}

	fmt.Printf("  %s┌%s%s┐%s\n", cyan, strings.Repeat("─", w-2), cyan, reset)
	fmt.Printf("  %s│%s  %sTotal:%s %s%d labs%s", cyan, reset, bold, reset, brWhite, len(labList), reset)
	if c, ok := diffCounts["easy"]; ok {
		fmt.Printf("  %s● %d easy%s", brGreen, c, reset)
	}
	if c, ok := diffCounts["medium"]; ok {
		fmt.Printf("  %s● %d medium%s", brYellow, c, reset)
	}
	if c, ok := diffCounts["hard"]; ok {
		fmt.Printf("  %s● %d hard%s", brRed, c, reset)
	}
	fmt.Printf("  %s│%s\n", cyan, reset)
	fmt.Printf("  %s└%s%s┘%s\n", cyan, strings.Repeat("─", w-2), cyan, reset)
	fmt.Println()
}

func PrintLabDetails(lab labs.Lab) {
	w := 70

	fmt.Println()
	fmt.Printf("  %s┌%s%s┐%s\n", cyan, strings.Repeat("─", w), cyan, reset)
	fmt.Printf("  %s│%s  %s%-*s%s  %s│%s\n", cyan, reset, bold, w-4, lab.Title(), reset, cyan, reset)
	fmt.Printf("  %s└%s%s┘%s\n", cyan, strings.Repeat("─", w), cyan, reset)
	fmt.Println()

	cert := labs.GetCert(lab)
	certStr := string(cert)
	if cert != labs.CertAll {
		certStr = fmt.Sprintf("%s %s%s", certBadge(cert), certStr, reset)
	} else {
		certStr = "N/A"
	}

	fmt.Printf("  %s%s▸ Details%s\n", bold, cyan, reset)
	fmt.Println()
	fmt.Printf("  %s  ID %s│%s  %s%s%s\n", dimW, dim, reset, bold, lab.ID(), reset)
	fmt.Printf("  %s  Category %s│%s  %s%s%s\n", dimW, dim, reset, bold, catColor(string(lab.Category()))+strings.ToUpper(string(lab.Category())), reset)
	fmt.Printf("  %s  Cert %s│%s  %s%s\n", dimW, dim, reset, certStr, reset)
	if weight := labs.GetDomainWeight(lab); weight > 0 {
		fmt.Printf("  %s  Domain Weight %s│%s  %s%d%%%s\n", dimW, dim, reset, brYellow, weight, reset)
	}
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
