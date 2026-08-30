package cli

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/labs"
)

type ExamConfig struct {
	Cert       string
	Duration   int // minutes
	NumLabs    int
	ShowBanner bool
}

type ExamLab struct {
	Lab     labs.Lab
	Minutes int
}

func GenerateExamPlan(cert string, numLabs int) ([]ExamLab, int) {
	allLabs := labs.List()

	var certFilter labs.Cert
	switch strings.ToUpper(cert) {
	case "CKA":
		certFilter = labs.CertCKA
	case "CKAD":
		certFilter = labs.CertCKAD
	case "CKS":
		certFilter = labs.CertCKS
	default:
		certFilter = labs.CertAll
	}

	var filtered []labs.Lab
	for _, l := range allLabs {
		if certFilter != labs.CertAll && labs.GetCert(l) != certFilter {
			continue
		}
		filtered = append(filtered, l)
	}

	if numLabs > len(filtered) {
		numLabs = len(filtered)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(filtered), func(i, j int) {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	})

	selected := filtered[:numLabs]
	var plan []ExamLab
	totalTime := 0

	perLab := 120 / numLabs // divide 2 hours evenly
	if perLab < 5 {
		perLab = 5
	}

	for _, l := range selected {
		plan = append(plan, ExamLab{Lab: l, Minutes: perLab})
		totalTime += perLab
	}

	return plan, totalTime
}

func PrintExamBanner(cert string, plan []ExamLab, totalMinutes int) {
	w := 70
	fmt.Println()
	fmt.Printf("  %s┌%s%s┐%s\n", cyan, strings.Repeat("─", w-2), cyan, reset)
	fmt.Printf("  %s│%s  %s%s EXAM SIMULATION MODE%s%*s%s│%s\n",
		cyan, reset, bold+brRed, cert, reset, w-4-len(cert)-20, "", cyan, reset)
	fmt.Printf("  %s├%s%s┤%s\n", cyan, strings.Repeat("─", w-2), cyan, reset)
	fmt.Printf("  %s│%s  %sDuration:%s %d minutes (%d hours)                         %s│%s\n",
		cyan, reset, brYellow, reset, totalMinutes, totalMinutes/60, cyan, reset)
	fmt.Printf("  %s│%s  %sLabs:%s    %d selected                                  %s│%s\n",
		cyan, reset, brWhite, reset, len(plan), cyan, reset)
	fmt.Printf("  %s│%s                                                           %s│%s\n", cyan, reset, cyan, reset)
	fmt.Printf("  %s│%s  %sRules:%s                                                  %s│%s\n",
		cyan, reset, cyan, reset, cyan, reset)
	fmt.Printf("  %s│%s    1. No hints allowed                                    %s│%s\n", cyan, reset, cyan, reset)
	fmt.Printf("  %s│%s    2. No solution viewing                                 %s│%s\n", cyan, reset, cyan, reset)
	fmt.Printf("  %s│%s    3. Use only kubectl (like real exam)                   %s│%s\n", cyan, reset, cyan, reset)
	fmt.Printf("  %s│%s    4. Complete all %d labs within the time limit             %s│%s\n",
		cyan, reset, len(plan), cyan, reset)
	fmt.Printf("  %s│%s                                                           %s│%s\n", cyan, reset, cyan, reset)
	fmt.Printf("  %s╚%s%s╝%s\n", cyan, strings.Repeat("─", w-2), cyan, reset)
	fmt.Println()

	fmt.Printf("  %sLabs in this exam:%s\n", bold, reset)
	for i, examLab := range plan {
		fmt.Printf("    %s%d.%s %-35s %s(%d min)%s\n",
			dim, i+1, reset, examLab.Lab.Title(), dimW, examLab.Minutes, reset)
	}
	fmt.Println()
}
