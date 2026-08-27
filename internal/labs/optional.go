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

// Certer is implemented by labs that target a specific certification.
type Certer interface {
	Cert() Cert
}

// DomainWeighter is implemented by labs that report an exam domain weight.
type DomainWeighter interface {
	DomainWeight() int
}

// HintLeveler is implemented by labs that provide tiered hints.
type HintLeveler interface {
	HintLevel(level int) string
}

// Domain constants for CKA exam domains
const (
	DomainWorkloadsScheduling = "workloads-scheduling"
	DomainServicesNetworking  = "services-networking"
	DomainStorage             = "storage"
	DomainTroubleshooting     = "troubleshooting"
	DomainClusterArchitecture = "cluster-architecture"
)

// Domain constants for CKAD exam domains
const (
	DomainAppDesignBuild      = "app-design-build"
	DomainAppDeployment       = "app-deployment"
	DomainAppObservability    = "app-observability"
	DomainAppConfigSecurity   = "app-config-security"
	DomainServicesNetworkCKAD = "services-networking-ckad"
)

// Domain constants for CKS exam domains
const (
	DomainClusterSetupCKS   = "cluster-setup-cks"
	DomainClusterHardening  = "cluster-hardening"
	DomainSystemHardening   = "system-hardening"
	DomainMicroserviceVulns = "microservice-vulns"
	DomainSupplyChain       = "supply-chain"
	DomainMonitoringLogging = "monitoring-logging"
)

// Legacy domain constants for backward compatibility
const (
	DomainWorkloads  = DomainWorkloadsScheduling
	DomainNetworking = DomainServicesNetworking
	DomainScheduling = DomainWorkloadsScheduling
	DomainSecurity   = DomainAppConfigSecurity
	DomainRBAC       = DomainClusterHardening
	DomainDNS        = DomainServicesNetworking
)

// DomainWeight returns the exam weight percentage for a category
func DomainWeightForCategory(c Category) int {
	switch c {
	case CategoryClusterArchitecture:
		return 25
	case CategoryWorkloadsScheduling:
		return 15
	case CategoryServicesNetworking:
		return 20
	case CategoryStorage:
		return 10
	case CategoryTroubleshooting:
		return 30
	case CategoryAppDesignBuild:
		return 20
	case CategoryAppDeployment:
		return 20
	case CategoryAppObservability:
		return 15
	case CategoryAppConfigSecurity:
		return 25
	case CategoryServicesNetworkCKAD:
		return 20
	case CategoryClusterSetupCKS:
		return 15
	case CategoryClusterHardening:
		return 15
	case CategorySystemHardening:
		return 10
	case CategoryMicroserviceVulns:
		return 20
	case CategorySupplyChain:
		return 20
	case CategoryMonitoringLogging:
		return 20
	default:
		return 0
	}
}

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

// GetCert returns the certification for a lab, falling back to category-based lookup
func GetCert(lab Lab) Cert {
	if c, ok := lab.(Certer); ok {
		return c.Cert()
	}
	return CertForCategory(lab.Category())
}

// GetDomainWeight returns the exam weight percentage for a lab
func GetDomainWeight(lab Lab) int {
	if d, ok := lab.(DomainWeighter); ok {
		return d.DomainWeight()
	}
	return DomainWeightForCategory(lab.Category())
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
