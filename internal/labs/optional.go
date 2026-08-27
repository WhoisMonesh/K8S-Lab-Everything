package labs

// Optional interfaces for labs that support extra features.
// Labs opt-in by implementing these — no changes needed to existing labs.

// Prerequisiter is implemented by labs that require other labs to be completed first.
type Prerequisiter interface {
	Prerequisites() []string
}

// Domainer is implemented by labs that map to a CKA exam domain.
type Domainer interface {
	Domain() string
}

// HintLeveler is implemented by labs that provide tiered hints.
type HintLeveler interface {
	HintLevel(level int) string
}

const (
	DomainWorkloads           = "workloads"
	DomainScheduling          = "scheduling"
	DomainNetworking          = "networking"
	DomainStorage             = "storage"
	DomainSecurity            = "security"
	DomainRBAC                = "rbac"
	DomainClusterArchitecture = "cluster-architecture"
	DomainTroubleshooting     = "troubleshooting"
	DomainLoggingMonitoring   = "logging-monitoring"
)

func GetPrerequisites(lab Lab) []string {
	if p, ok := lab.(Prerequisiter); ok {
		return p.Prerequisites()
	}
	return nil
}

func GetDomain(lab Lab) string {
	if d, ok := lab.(Domainer); ok {
		return d.Domain()
	}
	return ""
}

func GetHintLevel(lab Lab, level int) string {
	if h, ok := lab.(HintLeveler); ok {
		return h.HintLevel(level)
	}
	hints := lab.Hints()
	if len(hints) == 0 {
		return "No hints available for this lab."
	}
	if level <= 0 {
		level = 1
	}
	if level > len(hints) {
		level = len(hints)
	}
	return hints[level-1]
}
