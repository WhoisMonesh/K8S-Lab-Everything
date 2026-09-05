package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/cli"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/cluster"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/config"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/labs"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/progress"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/update"

	// Import labs to register them
	_ "github.com/WhoisMonesh/K8S-Lab-Everything/internal/labs"
)

var (
	cfgFile string
	cfg     *config.Config

	bold      = "\033[1m"
	dimW      = "\033[90m"
	reset     = "\033[0m"
	brRed     = "\033[91m"
	brGreen   = "\033[92m"
	brYellow  = "\033[93m"
	brBlue    = "\033[94m"
	brMagenta = "\033[95m"
	brCyan    = "\033[96m"
	brWhite   = "\033[97m"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}


var newCmd = &cobra.Command{
    Use:   "new <lab-id>",
    Short: "Scaffold a new lab file",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        id := args[0]
        // Flags for metadata
        title, _ := cmd.Flags().GetString("title")
        cat, _ := cmd.Flags().GetString("category")
        diff, _ := cmd.Flags().GetString("difficulty")
        est, _ := cmd.Flags().GetInt("time")
        tagsStr, _ := cmd.Flags().GetString("tags")
        // Basic validation
        if title == "" {
            return fmt.Errorf("--title is required")
        }
        if cat == "" {
            return fmt.Errorf("--category is required")
        }
        if diff == "" {
            return fmt.Errorf("--difficulty is required")
        }
        // Convert category and difficulty to constant names
        catConst, err := categoryToConst(cat)
        if err != nil { return err }
        diffConst, err := difficultyToConst(diff)
        if err != nil { return err }
        // Prepare tags slice literal
        tags := strings.Split(tagsStr, ",")
        var tagsLit strings.Builder
        for i, t := range tags {
            t = strings.TrimSpace(t)
            tagsLit.WriteString("\"")
            tagsLit.WriteString(t)
            tagsLit.WriteString("\"")
            if i < len(tags)-1 {
                tagsLit.WriteString(", ")
            }
        }
        // Determine filename
        filename := fmt.Sprintf("internal/labs/lab_%s.go", id)
        if _, err := os.Stat(filename); err == nil {
            return fmt.Errorf("lab file %s already exists", filename)
        }
        // Build file content
        tmpl := `package labs

import (
    "context"
    "fmt"
)

type Lab%s struct {
    BaseLab
}

var _ Lab = (*Lab%s)(nil)

func (l *Lab%s) ID() string { return "%s" }
func (l *Lab%s) Title() string { return "%s" }
func (l *Lab%s) Category() Category { return %s }
func (l *Lab%s) Difficulty() Difficulty { return %s }
func (l *Lab%s) Description() string { return "TODO: add description" }
func (l *Lab%s) Hints() []string { return []string{} }
func (l *Lab%s) EstimatedTime() int { return %d }
func (l *Lab%s) Tags() []string { return []string{%s} }

func (l *Lab%s) Break(ctx context.Context, kubeconfigPath string) error {
    // TODO: implement break logic
    return fmt.Errorf("break not implemented")
}

func (l *Lab%s) Verify(ctx context.Context, kubeconfigPath string) error {
    // TODO: implement verify logic
    return fmt.Errorf("verify not implemented")
}

func (l *Lab%s) SolutionSteps() []SolutionStep {
    return []SolutionStep{{
        Description: "TODO: write steps",
        Command:     "",
        Notes:       "",
    }}
}
`
        content := fmt.Sprintf(tmpl,
            strings.Title(id), strings.Title(id), strings.Title(id), id,
            strings.Title(id), title,
            strings.Title(id), catConst,
            strings.Title(id), diffConst,
            strings.Title(id), // Description placeholder
            strings.Title(id), // Hints placeholder
            strings.Title(id), est,
            strings.Title(id), tagsLit.String(),
            strings.Title(id), // Break
            strings.Title(id), // Verify
            strings.Title(id), // SolutionSteps
        )
        // Write the file
        if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
            return err
        }
        fmt.Printf("Created new lab scaffold at %s\n", filename)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
    rootCmd.AddCommand(updateCmd)
    rootCmd.AddCommand(initCmd)
    rootCmd.AddCommand(upCmd)
    rootCmd.AddCommand(downCmd)
    rootCmd.AddCommand(labCmd)
    rootCmd.AddCommand(newCmd)
    // Flags for the new command
    newCmd.Flags().String("title", "", "Lab title (required)")
    newCmd.Flags().String("category", "", "Lab category (required, e.g., cluster-architecture)")
    newCmd.Flags().String("difficulty", "", "Difficulty (easy|medium|hard, required)")
    newCmd.Flags().Int("time", 20, "Estimated time in minutes")
    newCmd.Flags().String("tags", "", "Comma‑separated list of tags")
}

func categoryToConst(cat string) (string, error) {
    switch strings.ToLower(cat) {
    case "cluster-architecture", "cka_cluster_architecture", "clusterarchitecture":
        return "CategoryClusterArchitecture", nil
    case "workloads-scheduling", "cka_workloads_scheduling":
        return "CategoryWorkloadsScheduling", nil
    case "services-networking", "cka_services_networking":
        return "CategoryServicesNetworking", nil
    case "storage":
        return "CategoryStorage", nil
    case "troubleshooting":
        return "CategoryTroubleshooting", nil
    case "app-design-build":
        return "CategoryAppDesignBuild", nil
    case "app-deployment":
        return "CategoryAppDeployment", nil
    case "app-observability":
        return "CategoryAppObservability", nil
    case "app-config-security":
        return "CategoryAppConfigSecurity", nil
    case "services-networking-ckad":
        return "CategoryServicesNetworkCKAD", nil
    case "cluster-setup-cks":
        return "CategoryClusterSetupCKS", nil
    case "cluster-hardening":
        return "CategoryClusterHardening", nil
    case "system-hardening":
        return "CategorySystemHardening", nil
    case "microservice-vulns":
        return "CategoryMicroserviceVulns", nil
    case "supply-chain":
        return "CategorySupplyChain", nil
    case "monitoring-logging":
        return "CategoryMonitoringLogging", nil
    default:
        return "", fmt.Errorf("unknown category %s", cat)
    }
}

func difficultyToConst(diff string) (string, error) {
    switch strings.ToLower(diff) {
    case "easy":
        return "DifficultyEasy", nil
    case "medium":
        return "DifficultyMedium", nil
    case "hard":
        return "DifficultyHard", nil
    default:
        return "", fmt.Errorf("unknown difficulty %s", diff)
    }
}

	Use:   "cka-lab-runner",
	Short: "A CKA/CKAD/CKS practice lab runner",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "version" || cmd.Name() == "update" || cmd.Name() == "help" {
			return
		}
		theme, _ := cmd.Flags().GetString("theme")
		if theme != "" {
			cli.SetTheme(theme)
		}
		update.CheckForUpdate()
	},
	Run: func(cmd *cobra.Command, args []string) {
		cli.PrintBanner()
		fmt.Println("  Quick Start:")
		fmt.Println()
		fmt.Printf("    %scka-lab-runner init%s                  %sCreate config file%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner up%s                    %sCreate local cluster%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner up --version v1.35.0%s  %sSelect KinD node version%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab list%s              %sList all labs%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab list --cert CKA%s   %sFilter by certification%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab list --search pod%s %sSearch labs%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab pick%s              %sPick interactively%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab run <id>%s          %sStart a lab%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab run <id> --timer%s  %sExam simulation (2h)%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab verify <id>%s       %sCheck your fix%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab solution <id>%s     %sShow solution%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab status%s            %sView progress%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab stats%s             %sView statistics%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner down%s                  %sDelete cluster%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner update%s                %sUpdate tool%s\n", bold, reset, dimW, reset)
		fmt.Println()
		fmt.Printf("  %sRun 'cka-lab-runner --help' for full command list%s\n\n", dimW, reset)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of cka-lab-runner",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("cka-lab-runner %s (commit: %s)\n", update.GetVersion(), update.GitCommit)
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update cka-lab-runner to the latest version",
	Long:  `Downloads and installs the latest release from GitHub.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return update.SelfUpdate()
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new cka-lab-runner configuration",
	Long:  `Creates a cka-lab-runner.yaml configuration file in the current directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := config.DefaultConfigFile

		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("config file already exists: %s", configPath)
		}

		defaultCfg := config.Default()
		if err := config.Save(defaultCfg, configPath); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		cli.Success(fmt.Sprintf("Created config file: %s", configPath))
		cli.Info("Edit this file to customize your cluster settings")
		return nil
	},
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create the local Kubernetes cluster",
	Long:  `Creates a local Kubernetes cluster based on the configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		recreate, _ := cmd.Flags().GetBool("recreate")
		random, _ := cmd.Flags().GetBool("random")
		versionFlag, _ := cmd.Flags().GetString("version")
		workers, _ := cmd.Flags().GetInt("workers")

		if err := loadConfig(); err != nil {
			return err
		}

		// Override config version if flag is set
		if versionFlag != "" {
			cfg.Cluster.KubernetesVersion = versionFlag
		}

		// Override config workers if flag is set
		if workers > 0 {
			cfg.Cluster.Workers = workers
		}

		// Show version suggestion if no version specified
		if cfg.Cluster.KubernetesVersion == "" || cfg.Cluster.KubernetesVersion == "v1.30.0" {
			fmt.Println()
			fmt.Printf("  %s╔══════════════════════════════════════════════════════════════╗%s\n", bold, reset)
			fmt.Printf("  %s║%s  %sKinD Node Version Selection%s                              %s║%s\n", bold, reset, bold+brCyan, reset, bold, reset)
			fmt.Printf("  %s╠══════════════════════════════════════════════════════════════╣%s\n", bold, reset)
			fmt.Printf("  %s║%s                                                           %s║%s\n", bold, reset, bold, reset)
			fmt.Printf("  %s║%s  %sRecommended for CKA/CKAD/CKS exams:%s                     %s║%s\n", bold, reset, brGreen, reset, bold, reset)
			fmt.Printf("  %s║%s    %s► v1.35.0%s  (exam version - recommended)               %s║%s\n", bold, reset, brGreen, reset, bold, reset)
			fmt.Printf("  %s║%s    %s► v1.34.0%s  (previous stable)                          %s║%s\n", bold, reset, brWhite, reset, bold, reset)
			fmt.Printf("  %s║%s    %s► v1.33.0%s  (older stable)                             %s║%s\n", bold, reset, brWhite, reset, bold, reset)
			fmt.Printf("  %s║%s                                                           %s║%s\n", bold, reset, bold, reset)
			fmt.Printf("  %s║%s  %sUse --version flag to select:%s                           %s║%s\n", bold, reset, brYellow, reset, bold, reset)
			fmt.Printf("  %s║%s    cka-lab-runner up --version v1.35.0                    %s║%s\n", bold, reset, bold, reset)
			fmt.Printf("  %s║%s                                                           %s║%s\n", bold, reset, bold, reset)
			fmt.Printf("  %s╚══════════════════════════════════════════════════════════════╝%s\n", bold, reset)
			fmt.Println()

			// Default to exam version
			if cfg.Cluster.KubernetesVersion == "" {
				cfg.Cluster.KubernetesVersion = "v1.35.0"
				fmt.Printf("  %s▸%s Using default: %sv1.35.0%s (CKA/CKAD/CKS exam version)\n\n", brCyan, reset, bold, reset)
			}
		}

		provider, err := createProvider()
		if err != nil {
			return fmt.Errorf("creating provider: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		exists, err := provider.Exists(ctx)
		if err != nil {
			return fmt.Errorf("checking if cluster exists: %w", err)
		}

		if exists {
			if recreate {
				cli.Info(fmt.Sprintf("Deleting existing cluster: %s", provider.Name()))
				if err := provider.Down(ctx); err != nil {
					return fmt.Errorf("deleting cluster: %w", err)
				}
			} else {
				cli.Info(fmt.Sprintf("Cluster already exists: %s (use --recreate to recreate)", provider.Name()))
				if random {
					return runRandomLab(nil)
				}
				return nil
			}
		}

		clusterInfo := fmt.Sprintf("Creating cluster: %s with %s", provider.Name(), cfg.Cluster.KubernetesVersion)
		if cfg.Cluster.Workers > 0 {
			clusterInfo += fmt.Sprintf(" (%d worker nodes)", cfg.Cluster.Workers)
		}
		cli.Info(clusterInfo)
		if err := provider.Up(ctx); err != nil {
			return fmt.Errorf("creating cluster: %w", err)
		}

		createdInfo := fmt.Sprintf("Cluster created: %s (%s)", provider.Name(), cfg.Cluster.KubernetesVersion)
		if cfg.Cluster.Workers > 0 {
			createdInfo += fmt.Sprintf(" [%d control-plane + %d workers]", 1, cfg.Cluster.Workers)
		}
		cli.Success(createdInfo)

		if random {
			return runRandomLab(nil)
		}
		return nil
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Delete the local Kubernetes cluster",
	Long:  `Deletes the local Kubernetes cluster.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(); err != nil {
			return err
		}

		provider, err := createProvider()
		if err != nil {
			return fmt.Errorf("creating provider: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		cli.Info(fmt.Sprintf("Deleting cluster: %s", provider.Name()))
		if err := provider.Down(ctx); err != nil {
			return fmt.Errorf("deleting cluster: %w", err)
		}

		cli.Success(fmt.Sprintf("Cluster deleted: %s", provider.Name()))
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cluster and lab progress",
	Long:  `Displays cluster health (nodes, pods) and lab completion progress.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(); err != nil {
			return err
		}

		provider, err := createProvider()
		if err != nil {
			return fmt.Errorf("creating provider: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		exists, err := provider.Exists(ctx)
		if err != nil {
			return fmt.Errorf("checking if cluster: %w", err)
		}

		fmt.Println()
		fmt.Printf("  %s╔══════════════════════════════════════════════════════════════╗%s\n", bold, reset)
		fmt.Printf("  %s║%s  %s📊 CLUSTER STATUS%s                                        %s║%s\n", bold, reset, bold+brCyan, reset, bold, reset)
		fmt.Printf("  %s╠══════════════════════════════════════════════════════════════╣%s\n", bold, reset)

		if !exists {
			fmt.Printf("  %s║%s  %sCluster:%s %-47s %s║%s\n", bold, reset, brWhite, reset, "NOT RUNNING", bold, reset)
			fmt.Printf("  %s║%s  %sRun:%s cka-lab-runner up                                   %s║%s\n", bold, reset, brYellow, reset, bold, reset)
		} else {
			fmt.Printf("  %s║%s  %sCluster:%s %-47s %s║%s\n", bold, reset, brWhite, reset, fmt.Sprintf("%s (%s)", provider.Name(), cfg.Cluster.KubernetesVersion), bold, reset)

			kubeconfigPath, err := provider.KubeconfigPath(ctx)
			if err == nil {
				// Node status
				nodeOut, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfigPath,
					"get", "nodes", "-o", "wide", "--no-headers").CombinedOutput()
				if err == nil {
					nodes := strings.Split(strings.TrimSpace(string(nodeOut)), "\n")
					readyCount := 0
					for _, n := range nodes {
						if strings.Contains(n, " Ready") {
							readyCount++
						}
					}
					fmt.Printf("  %s║%s  %sNodes:%s %-47s %s║%s\n", bold, reset, brWhite, reset, fmt.Sprintf("%d total, %d Ready", len(nodes), readyCount), bold, reset)

					for _, n := range nodes {
						fields := strings.Fields(n)
						if len(fields) >= 5 {
							name := fields[0]
							status := "NotReady"
							for _, f := range fields {
								if f == "Ready" || strings.HasPrefix(f, "Ready,") {
									status = "Ready"
								}
							}
							role := "worker"
							if strings.Contains(n, "control-plane") || strings.Contains(n, "master") {
								role = "control-plane"
							}
							color := brGreen
							if status != "Ready" {
								color = brRed
							}
							fmt.Printf("  %s║%s    %s%-30s %s%-10s %s%-12s%s  %s║%s\n",
								bold, reset, brWhite, name, color, status, dimW, role, reset, bold, reset)
						}
					}
				}

				// System pods
				podOut, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfigPath,
					"get", "pods", "-n", "kube-system", "--no-headers").CombinedOutput()
				if err == nil {
					pods := strings.Split(strings.TrimSpace(string(podOut)), "\n")
					runningCount := 0
					for _, p := range pods {
						if strings.Contains(p, " Running ") || strings.Contains(p, "Completed") {
							runningCount++
						}
					}
					fmt.Printf("  %s║%s  %sSystem Pods:%s %-40s %s║%s\n", bold, reset, brWhite, reset, fmt.Sprintf("%d total, %d running", len(pods), runningCount), bold, reset)
				}
			}
		}

		fmt.Printf("  %s║%s                                                           %s║%s\n", bold, reset, bold, reset)

		// Lab progress
		completed := progress.CompletedCount()
		total := len(labs.List())
		activeLab := progress.ActiveLab()
		streak := progress.CurrentStreak()

		fmt.Printf("  %s║%s  %s📊 LAB PROGRESS%s                                          %s║%s\n", bold, reset, bold+brGreen, reset, bold, reset)
		fmt.Printf("  %s║%s  %sCompleted:%s %-43s %s║%s\n", bold, reset, brWhite, reset, fmt.Sprintf("%d/%d (%d%%)", completed, total, completed*100/max(total, 1)), bold, reset)

		if activeLab != "" {
			fmt.Printf("  %s║%s  %sActive:%s %-45s %s║%s\n", bold, reset, brYellow, reset, activeLab, bold, reset)
		}
		if streak > 0 {
			fmt.Printf("  %s║%s  %sStreak:%s %-45s %s║%s\n", bold, reset, brGreen, reset, fmt.Sprintf("%d day(s)", streak), bold, reset)
		}

		fmt.Printf("  %s╚══════════════════════════════════════════════════════════════╝%s\n", bold, reset)
		fmt.Println()
		return nil
	},
}

var labCmd = &cobra.Command{
	Use:   "lab",
	Short: "Manage practice labs",
	Long:  `Commands for listing, running, and viewing solutions for practice labs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("unknown lab command %q\n\nDid you mean:\n  cka-lab-runner lab run %s      # start the lab\n  cka-lab-runner lab verify %s    # check your fix\n  cka-lab-runner lab solution %s  # show solution", args[0], args[0], args[0], args[0])
		}
		return cmd.Help()
	},
}

var labListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available labs",
	Long:  `Lists all available practice labs with their metadata.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		categoryFilter, _ := cmd.Flags().GetString("category")
		difficultyFilter, _ := cmd.Flags().GetString("difficulty")
		certFilter, _ := cmd.Flags().GetString("cert")
		domainFilter, _ := cmd.Flags().GetString("domain")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		csvOutput, _ := cmd.Flags().GetBool("csv")
		showProgress, _ := cmd.Flags().GetBool("progress")
		tagFilter, _ := cmd.Flags().GetString("tag")
		searchFilter, _ := cmd.Flags().GetString("search")
		resourceFilter, _ := cmd.Flags().GetString("resource")

		allLabs := labs.List()
		var filteredLabs []labs.Lab

		for _, lab := range allLabs {
			matches := true
			if categoryFilter != "" && string(lab.Category()) != categoryFilter {
				matches = false
			}
			if difficultyFilter != "" && string(lab.Difficulty()) != difficultyFilter {
				matches = false
			}
			if certFilter != "" {
				cert := labs.Cert(strings.ToUpper(certFilter))
				if cert != labs.CertCKA && cert != labs.CertCKAD && cert != labs.CertCKS {
					return fmt.Errorf("invalid cert %q: must be CKA, CKAD, or CKS", certFilter)
				}
				if labs.GetCert(lab) != cert {
					matches = false
				}
			}
			if domainFilter != "" {
				d := labs.GetDomain(lab)
				if d == "" {
					d = string(lab.Category())
				}
				if d != domainFilter {
					matches = false
				}
			}
			if tagFilter != "" {
				found := false
				for _, t := range lab.Tags() {
					if t == tagFilter {
						found = true
						break
					}
				}
				if !found {
					matches = false
				}
			}
			if searchFilter != "" {
				searchLower := strings.ToLower(searchFilter)
				idMatch := strings.Contains(strings.ToLower(lab.ID()), searchLower)
				titleMatch := strings.Contains(strings.ToLower(lab.Title()), searchLower)
				descMatch := strings.Contains(strings.ToLower(lab.Description()), searchLower)
				tagMatch := false
				for _, t := range lab.Tags() {
					if strings.Contains(strings.ToLower(t), searchLower) {
						tagMatch = true
						break
					}
				}
				if !idMatch && !titleMatch && !descMatch && !tagMatch {
					matches = false
				}
			}
			if resourceFilter != "" {
				resourceLower := strings.ToLower(resourceFilter)
				found := false
				idLower := strings.ToLower(lab.ID())
				titleLower := strings.ToLower(lab.Title())
				descLower := strings.ToLower(lab.Description())
				if strings.Contains(idLower, resourceLower) || strings.Contains(titleLower, resourceLower) || strings.Contains(descLower, resourceLower) {
					found = true
				}
				for _, t := range lab.Tags() {
					if strings.Contains(strings.ToLower(t), resourceLower) {
						found = true
						break
					}
				}
				if !found {
					matches = false
				}
			}
			if matches {
				filteredLabs = append(filteredLabs, lab)
			}
		}

		if jsonOutput {
			type labJSON struct {
				ID           string   `json:"id"`
				Title        string   `json:"title"`
				Category     string   `json:"category"`
				Cert         string   `json:"cert"`
				DomainWeight int      `json:"domain_weight"`
				Difficulty   string   `json:"difficulty"`
				Estimated    int      `json:"estimated_minutes"`
				Tags         []string `json:"tags"`
				Domain       string   `json:"domain,omitempty"`
				Completed    bool     `json:"completed"`
			}
			var out []labJSON
			for _, lab := range filteredLabs {
				out = append(out, labJSON{
					ID:           lab.ID(),
					Title:        lab.Title(),
					Category:     string(lab.Category()),
					Cert:         string(labs.GetCert(lab)),
					DomainWeight: labs.GetDomainWeight(lab),
					Difficulty:   string(lab.Difficulty()),
					Estimated:    lab.EstimatedTime(),
					Tags:         lab.Tags(),
					Domain:       labs.GetDomain(lab),
					Completed:    progress.IsCompleted(lab.ID()),
				})
			}
			data, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if csvOutput {
			fmt.Println("ID,Title,Category,Cert,Difficulty,Estimated,Completed")
			for _, lab := range filteredLabs {
				completed := "false"
				if progress.IsCompleted(lab.ID()) {
					completed = "true"
				}
				fmt.Printf("%s,%s,%s,%s,%s,%d,%s\n",
					lab.ID(),
					strings.ReplaceAll(lab.Title(), ",", ";"),
					lab.Category(),
					labs.GetCert(lab),
					lab.Difficulty(),
					lab.EstimatedTime(),
					completed,
				)
			}
			return nil
		}

		cli.PrintLabListWithProgress(filteredLabs, showProgress)
		return nil
	},
}

var labRunCmd = &cobra.Command{
	Use:   "run <lab-id>",
	Short: "Run a practice lab",
	Long:  `Applies a broken scenario to the cluster for practice.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		labID := args[0]
		timed, _ := cmd.Flags().GetBool("timed")
		timer, _ := cmd.Flags().GetBool("timer")
		timeLimit, _ := cmd.Flags().GetInt("time-limit")
		ns, _ := cmd.Flags().GetString("namespace")

		lab, err := labs.Get(labID)
		if err != nil {
			return err
		}

		if err := loadConfig(); err != nil {
			return err
		}

		provider, err := createProvider()
		if err != nil {
			return fmt.Errorf("creating provider: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Provision a cluster matching the lab's declared topology (if any),
		// creating/recreating it when the running version or worker count
		// differs from what the lab scenario requires.
		if spec, ok := labs.GetClusterSpec(lab); ok {
			if err := provisionClusterForSpec(cmd, spec); err != nil {
				return err
			}
		} else {
			exists, err := provider.Exists(ctx)
			if err != nil {
				return fmt.Errorf("checking if cluster exists: %w", err)
			}
			if !exists {
				return fmt.Errorf("cluster does not exist. Run 'cka-lab-runner up' first")
			}
		}

		// Provision any pending (unjoined) node containers the lab requires.
		for _, pn := range labs.GetRequiredNodes(lab) {
			if err := provisionPendingNode(ctx, provider.Name(), pn); err != nil {
				return fmt.Errorf("provisioning pending node for lab: %w", err)
			}
		}

		// Check prerequisites
		if prereqs := labs.GetPrerequisites(lab); len(prereqs) > 0 {
			var uncompleted []string
			for _, p := range prereqs {
				if !progress.IsCompleted(p) {
					uncompleted = append(uncompleted, p)
				}
			}
			if len(uncompleted) > 0 {
				cli.Warning(fmt.Sprintf("Prerequisites not completed: %s", strings.Join(uncompleted, ", ")))
				cli.Info("Complete these labs first, or run anyway with --force")
				force, _ := cmd.Flags().GetBool("force")
				if !force {
					return fmt.Errorf("prerequisites not met")
				}
			}
		}

		kubeconfigPath, err := provider.KubeconfigPath(ctx)
		if err != nil {
			return fmt.Errorf("getting kubeconfig: %w", err)
		}

		cli.Info("Preparing lab environment...")
		if ns != "" {
			nsYaml := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, ns)
			nsCmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfigPath, "apply", "-f", "-")
			nsCmd.Stdin = strings.NewReader(nsYaml)
			if out, err := nsCmd.CombinedOutput(); err != nil {
				cli.Warning(fmt.Sprintf("Creating namespace %s: %v (%s)", ns, err, string(out)))
			}
			cli.Info(fmt.Sprintf("Using namespace: %s", ns))
		}
		if err := lab.Prepare(ctx, kubeconfigPath); err != nil {
			cli.Warning(fmt.Sprintf("Prepare step failed (may be optional): %v", err))
		}

		cli.Info("Applying broken scenario...")
		if err := lab.Break(ctx, kubeconfigPath); err != nil {
			return fmt.Errorf("breaking cluster: %w", err)
		}

		if err := lab.VerifyBroken(ctx, kubeconfigPath); err != nil {
			cli.Warning(fmt.Sprintf("Verify broken step failed (may be optional): %v", err))
		}

		cli.PrintLabDetails(lab)
		cli.Success("Lab scenario applied successfully!")

		progress.SetActiveLab(labID)

		// Exam simulation mode (2 hours)
		if timer {
			os.Setenv("CKA_LAB_EXAM_MODE", "1")
			timeLimit = 120 // 2 hours like real CKA/CKAD/CKS exam
			fmt.Println()
			fmt.Printf("  %s╔══════════════════════════════════════════════════════════════╗%s\n", bold, reset)
			fmt.Printf("  %s║%s  %s⏰ EXAM SIMULATION MODE%s                                    %s║%s\n", bold, reset, bold+brRed, reset, bold, reset)
			fmt.Printf("  %s╠══════════════════════════════════════════════════════════════╣%s\n", bold, reset)
			fmt.Printf("  %s║%s  %sTime Limit:%s 2 hours (120 minutes)                        %s║%s\n", bold, reset, brYellow, reset, bold, reset)
			fmt.Printf("  %s║%s  %sLab:%s %-50s %s║%s\n", bold, reset, brWhite, reset, lab.ID(), bold, reset)
			fmt.Printf("  %s║%s                                                           %s║%s\n", bold, reset, bold, reset)
			fmt.Printf("  %s║%s  %sRules:%s                                                  %s║%s\n", bold, reset, brCyan, reset, bold, reset)
			fmt.Printf("  %s║%s    • No hints allowed                                     %s║%s\n", bold, reset, bold, reset)
			fmt.Printf("  %s║%s    • No solution viewing                                  %s║%s\n", bold, reset, bold, reset)
			fmt.Printf("  %s║%s    • Use only kubectl (like real exam)                    %s║%s\n", bold, reset, bold, reset)
			fmt.Printf("  %s║%s                                                           %s║%s\n", bold, reset, bold, reset)
			fmt.Printf("  %s╚══════════════════════════════════════════════════════════════╝%s\n", bold, reset)
			fmt.Println()
			cli.Info(fmt.Sprintf("Timer started! You have %d minutes to fix this lab!", timeLimit))
			cli.Info("Fix the issue before time runs out!")
			go runCountdown(timeLimit, labID)
		} else if timed {
			if timeLimit <= 0 {
				timeLimit = lab.EstimatedTime()
			}
			cli.Info(fmt.Sprintf("TIMED MODE: You have %d minutes to fix this lab!", timeLimit))
			cli.Info("Timer started now. Fix the issue before time runs out!")
			go runCountdown(timeLimit, labID)
		} else {
			cli.Info(fmt.Sprintf("Use 'cka-lab-runner lab solution %s' to see the solution", labID))
		}
		cli.Info(fmt.Sprintf("Use 'cka-lab-runner lab verify %s' to check your fix", labID))

		_ = ns // namespace isolation handled by the lab's Break/Verify if they use it
		return nil
	},
}

var labSolutionCmd = &cobra.Command{
	Use:   "solution <lab-id>",
	Short: "Show the solution for a lab",
	Long:  `Displays step-by-step instructions for solving a lab.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("CKA_LAB_EXAM_MODE") == "1" {
			return fmt.Errorf("solutions are blocked during exam simulation mode")
		}
		labID := args[0]

		lab, err := labs.Get(labID)
		if err != nil {
			return err
		}

		solution := labs.FormatSolution(lab)
		fmt.Println(solution)
		return nil
	},
}

var labRandomCmd = &cobra.Command{
	Use:   "random",
	Short: "Select a random lab",
	Long:  `Selects and runs a random lab, optionally filtered by category, difficulty, and certification.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		seed, _ := cmd.Flags().GetInt64("seed")
		categoryFilter, _ := cmd.Flags().GetString("category")
		difficultyFilter, _ := cmd.Flags().GetString("difficulty")
		certFilter, _ := cmd.Flags().GetString("cert")

		if seed == 0 {
			seed = time.Now().UnixNano()
		}

		var category labs.Category
		if categoryFilter != "" {
			category = labs.Category(categoryFilter)
		}

		var difficulty labs.Difficulty
		if difficultyFilter != "" {
			difficulty = labs.Difficulty(difficultyFilter)
		}

		var cert labs.Cert
		if certFilter != "" {
			cert = labs.Cert(strings.ToUpper(certFilter))
		}

		lab, err := labs.Random(seed, category, difficulty, cert)
		if err != nil {
			return err
		}

		cli.Info(fmt.Sprintf("Selected lab: %s", lab.ID()))
		return labRunCmd.RunE(cmd, []string{lab.ID()})
	},
}

var labVerifyCmd = &cobra.Command{
	Use:   "verify <lab-id>",
	Short: "Verify if you fixed the lab correctly",
	Long:  `Checks if the lab issue has been resolved correctly.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		labID := args[0]
		lab, err := labs.Get(labID)
		if err != nil {
			return err
		}

		if err := loadConfig(); err != nil {
			return err
		}

		provider, err := createProvider()
		if err != nil {
			return fmt.Errorf("creating provider: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		exists, err := provider.Exists(ctx)
		if err != nil {
			return fmt.Errorf("checking if cluster exists: %w", err)
		}
		if !exists {
			return fmt.Errorf("cluster does not exist")
		}

		kubeconfigPath, err := provider.KubeconfigPath(ctx)
		if err != nil {
			return fmt.Errorf("getting kubeconfig: %w", err)
		}

		cli.Info(fmt.Sprintf("Verifying lab: %s", lab.Title()))
		startTime := time.Now()
		if err := lab.Verify(ctx, kubeconfigPath); err != nil {
			duration := time.Since(startTime)
			progress.RecordAttempt(labID, false, duration)
			cli.Error(fmt.Sprintf("Lab not fixed yet: %v", err))
			cli.Info(fmt.Sprintf("Keep trying! Use 'cka-lab-runner lab hint %s' for help", labID))
			return nil
		}
		duration := time.Since(startTime)
		progress.RecordAttempt(labID, true, duration)

		cli.Success(fmt.Sprintf("Congratulations! You successfully fixed: %s", lab.Title()))

		if !progress.IsCompleted(labID) {
			ns, _ := cmd.Flags().GetString("namespace")
			progress.RecordCompletion(
				labID,
				lab.Title(),
				string(lab.Category()),
				string(lab.Difficulty()),
				duration,
				lab.EstimatedTime(),
				false, false, ns,
			)
			cli.Info(fmt.Sprintf("Progress saved! (took %s) Run 'cka-lab-runner lab status' to see your progress.", duration.Round(time.Second)))
		} else {
			cli.Info("Already recorded in progress. Nice work!")
		}

		return nil
	},
}

var labHintCmd = &cobra.Command{
	Use:   "hint <lab-id>",
	Short: "Get a hint for a lab",
	Long:  `Shows a progressive hint. Use --level 1-3 for increasingly specific help.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("CKA_LAB_EXAM_MODE") == "1" {
			return fmt.Errorf("hints are blocked during exam simulation mode")
		}
		labID := args[0]
		level, _ := cmd.Flags().GetInt("level")

		lab, err := labs.Get(labID)
		if err != nil {
			return err
		}

		if level <= 0 {
			level = 1
		}

		hint := labs.GetHintLevel(lab, level)

		maxLevel := len(lab.Hints())
		if maxLevel == 0 {
			maxLevel = 3
		}

		fmt.Printf("\nHint for: %s (level %d/%d)\n", lab.Title(), level, maxLevel)
		fmt.Println(strings.Repeat("─", 50))
		fmt.Printf("  %s\n\n", hint)

		if level < maxLevel {
			cli.Info(fmt.Sprintf("Need more help? Try: cka-lab-runner lab hint %s --level %d", labID, level+1))
		} else {
			cli.Info(fmt.Sprintf("Last hint. Stuck? Try: cka-lab-runner lab solution %s", labID))
		}
		return nil
	},
}

var labStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show your lab progress",
	Long:  `Displays a summary of completed labs and time spent.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		if jsonOutput {
			data, err := progress.ExportJSON()
			if err != nil {
				return fmt.Errorf("exporting progress: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(progress.Summary())
		return nil
	},
}

var labStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show detailed statistics",
	Long:  `Displays detailed statistics about your lab progress by certification and domain.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		allLabs := labs.List()
		completed := progress.CompletedCount()
		total := len(allLabs)

		// Count by certification
		ckaTotal, ckadTotal, cksTotal := 0, 0, 0
		ckaDone, ckadDone, cksDone := 0, 0, 0

		for _, lab := range allLabs {
			cert := labs.GetCert(lab)
			isDone := progress.IsCompleted(lab.ID())

			switch cert {
			case labs.CertCKA:
				ckaTotal++
				if isDone {
					ckaDone++
				}
			case labs.CertCKAD:
				ckadTotal++
				if isDone {
					ckadDone++
				}
			case labs.CertCKS:
				cksTotal++
				if isDone {
					cksDone++
				}
			}
		}

		fmt.Println()
		fmt.Printf("  %s╔══════════════════════════════════════════════════════════════╗%s\n", bold, reset)
		fmt.Printf("  %s║%s  %s📊 LAB STATISTICS%s                                          %s║%s\n", bold, reset, bold+brCyan, reset, bold, reset)
		fmt.Printf("  %s╠══════════════════════════════════════════════════════════════╣%s\n", bold, reset)
		fmt.Printf("  %s║%s                                                           %s║%s\n", bold, reset, bold, reset)
		fmt.Printf("  %s║%s  %sOverall:%s %d/%d labs completed (%d%%)                      %s║%s\n", bold, reset, brWhite, reset, completed, total, completed*100/total, bold, reset)
		fmt.Printf("  %s║%s                                                           %s║%s\n", bold, reset, bold, reset)
		fmt.Printf("  %s║%s  %sCKA:%s   %d/%d labs completed (%d%%)                        %s║%s\n", bold, reset, brBlue, reset, ckaDone, ckaTotal, ckaDone*100/ckaTotal, bold, reset)
		fmt.Printf("  %s║%s  %sCKAD:%s  %d/%d labs completed (%d%%)                        %s║%s\n", bold, reset, brCyan, reset, ckadDone, ckadTotal, ckadDone*100/ckadTotal, bold, reset)
		fmt.Printf("  %s║%s  %sCKS:%s   %d/%d labs completed (%d%%)                        %s║%s\n", bold, reset, brMagenta, reset, cksDone, cksTotal, cksDone*100/cksTotal, bold, reset)
		fmt.Printf("  %s║%s                                                           %s║%s\n", bold, reset, bold, reset)
		fmt.Printf("  %s╚══════════════════════════════════════════════════════════════╝%s\n", bold, reset)
		fmt.Println()

		return nil
	},
}

var labExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export completion history",
	Long:  `Exports your lab completion history as JSON, markdown report, or certificate.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")

		switch format {
		case "markdown", "md":
			fmt.Println(progress.ExportMarkdown())
		case "certificate", "cert":
			fmt.Println(progress.ExportCertificate())
		default:
			data, err := progress.ExportJSON()
			if err != nil {
				return fmt.Errorf("exporting progress: %w", err)
			}
			fmt.Println(string(data))
		}
		return nil
	},
}

var labStreakCmd = &cobra.Command{
	Use:   "streak",
	Short: "Show your practice streak",
	Long:  `Displays your current and longest practice streak.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println()
		fmt.Printf("  %s%sPractice Streak%s\n", bold, brCyan, reset)
		fmt.Println()
		fmt.Println(progress.StreakInfo())
		return nil
	},
}

var labRateCmd = &cobra.Command{
	Use:   "rate <lab-id> <1-5>",
	Short: "Rate a lab you completed",
	Long:  `Rate a lab from 1 (easy) to 5 (very hard) based on your experience.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		labID := args[0]
		var rating int
		if _, err := fmt.Sscanf(args[1], "%d", &rating); err != nil {
			return fmt.Errorf("invalid rating: %s (must be 1-5)", args[1])
		}

		if err := progress.RateLab(labID, rating); err != nil {
			return err
		}

		stars := ""
		for i := 0; i < 5; i++ {
			if i < rating {
				stars += "★"
			} else {
				stars += "☆"
			}
		}
		cli.Success(fmt.Sprintf("Rated %s: %s (%d/5)", labID, stars, rating))
		return nil
	},
}

var labExamCmd = &cobra.Command{
	Use:   "exam",
	Short: "Start an exam simulation",
	Long:  `Generates a timed exam with random labs matching real CKA/CKAD/CKS structure.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cert, _ := cmd.Flags().GetString("cert")
		numLabs, _ := cmd.Flags().GetInt("num-labs")

		if cert == "" {
			cert = "CKA"
		}
		if numLabs <= 0 {
			numLabs = 15
		}

		plan, totalMinutes := cli.GenerateExamPlan(cert, numLabs)
		if len(plan) == 0 {
			return fmt.Errorf("no labs found for cert %s", cert)
		}

		cli.PrintExamBanner(cert, plan, totalMinutes)

		os.Setenv("CKA_LAB_EXAM_MODE", "1")
		defer os.Unsetenv("CKA_LAB_EXAM_MODE")

		if err := loadConfig(); err != nil {
			return err
		}

		provider, err := createProvider()
		if err != nil {
			return fmt.Errorf("creating provider: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		exists, err := provider.Exists(ctx)
		if err != nil {
			return fmt.Errorf("checking if cluster exists: %w", err)
		}
		if !exists {
			return fmt.Errorf("cluster does not exist. Run 'cka-lab-runner up' first")
		}

		kubeconfigPath, err := provider.KubeconfigPath(ctx)
		if err != nil {
			return fmt.Errorf("getting kubeconfig: %w", err)
		}

		passed := 0
		failed := 0
		totalTime := time.Duration(0)
		examStart := time.Now()

		for i, examLab := range plan {
			fmt.Printf("\n  %s━━━ Lab %d/%d: %s (%d min) ━━━%s\n\n",
				bold, i+1, len(plan), examLab.Lab.Title(), examLab.Minutes, reset)

			labStart := time.Now()

			prepareCtx, prepareCancel := context.WithTimeout(context.Background(), 2*time.Minute)
			if err := examLab.Lab.Prepare(prepareCtx, kubeconfigPath); err != nil {
				cli.Warning(fmt.Sprintf("Prepare failed: %v", err))
			}
			prepareCancel()

			if err := examLab.Lab.Break(ctx, kubeconfigPath); err != nil {
				cli.Warning(fmt.Sprintf("Break failed: %v", err))
				failed++
				continue
			}

			examLab.Lab.VerifyBroken(ctx, kubeconfigPath)

			fmt.Printf("  %s▸%s Lab scenario applied. You have %d minutes.\n", brCyan, reset, examLab.Minutes)
			fmt.Printf("  %s▸%s Fix the issue, then run: %scka-lab-runner lab verify %s%s\n\n",
				brCyan, reset, brCyan, examLab.Lab.ID(), reset)

			go runCountdown(examLab.Minutes, examLab.Lab.ID())

			fmt.Printf("  %sPress Enter when done (or type 'skip' to skip this lab)...%s ", dimW, reset)
			var input string
			fmt.Scanln(&input)

			labDuration := time.Since(labStart)
			totalTime += labDuration

			if strings.ToLower(strings.TrimSpace(input)) == "skip" {
				fmt.Printf("  %s⏭ Skipped lab %s%s\n", brYellow, examLab.Lab.ID(), reset)
				failed++
				continue
			}

			verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 2*time.Minute)
			if err := examLab.Lab.Verify(verifyCtx, kubeconfigPath); err != nil {
				fmt.Printf("  %s✗ FAILED:%s %v\n", brRed, reset, err)
				failed++
			} else {
				fmt.Printf("  %s✓ PASSED%s (took %s)\n", brGreen, reset, labDuration.Round(time.Second))
				passed++
				progress.RecordCompletion(
					examLab.Lab.ID(),
					examLab.Lab.Title(),
					string(examLab.Lab.Category()),
					string(examLab.Lab.Difficulty()),
					labDuration,
					examLab.Minutes,
					true, false, "",
				)
			}
			verifyCancel()
		}

		totalExamTime := time.Since(examStart)

		fmt.Println()
		fmt.Printf("  %s╔══════════════════════════════════════════════════════════════╗%s\n", bold, reset)
		fmt.Printf("  %s║%s  %s📊 EXAM RESULTS%s                                          %s║%s\n", bold, reset, bold+brCyan, reset, bold, reset)
		fmt.Printf("  %s╠══════════════════════════════════════════════════════════════╣%s\n", bold, reset)
		fmt.Printf("  %s║%s  %sCertification:%s %-40s %s║%s\n", bold, reset, brWhite, reset, cert, bold, reset)
		fmt.Printf("  %s║%s  %sTotal Labs:%s   %-40d %s║%s\n", bold, reset, brWhite, reset, len(plan), bold, reset)
		fmt.Printf("  %s║%s  %sPassed:%s      %s%-40d%s %s║%s\n", bold, reset, brWhite, reset, brGreen, passed, reset, bold, reset)
		fmt.Printf("  %s║%s  %sFailed:%s      %s%-40d%s %s║%s\n", bold, reset, brWhite, reset, brRed, failed, reset, bold, reset)
		score := 0
		if len(plan) > 0 {
			score = passed * 100 / len(plan)
		}
		scoreColor := brRed
		if score >= 67 {
			scoreColor = brGreen
		} else if score >= 50 {
			scoreColor = brYellow
		}
		fmt.Printf("  %s║%s  %sScore:%s       %s%d%%%s%-36s %s║%s\n", bold, reset, brWhite, reset, scoreColor, score, reset, "", bold, reset)
		fmt.Printf("  %s║%s  %sTime Taken:%s  %-40s %s║%s\n", bold, reset, brWhite, reset, totalExamTime.Round(time.Second), bold, reset)
		fmt.Printf("  %s║%s                                                           %s║%s\n", bold, reset, bold, reset)
		fmt.Printf("  %s╚══════════════════════════════════════════════════════════════╝%s\n", bold, reset)
		fmt.Println()

		if score >= 67 {
			cli.Success(fmt.Sprintf("Great job! You scored %d%% on the %s exam simulation!", score, cert))
		} else {
			cli.Warning(fmt.Sprintf("Score: %d%%. Keep practicing! You need 67%% to pass.", score))
		}

		return nil
	},
}

var completionCmd *cobra.Command

func runCountdown(minutes int, labID string) {
	total := time.Duration(minutes) * time.Minute
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(total)

	for remaining := range ticker.C {
		left := time.Until(deadline)
		if left <= 0 {
			fmt.Printf("\n⏰ TIME'S UP for lab %s! (%d minutes elapsed)\n", labID, minutes)
			fmt.Printf("   Run 'cka-lab-runner lab verify %s' to check your work.\n\n", labID)
			return
		}
		_ = remaining
		if int(left.Minutes())%5 == 0 && int(left.Seconds())%60 < 30 {
			fmt.Printf("  ⏳ %d:%02d remaining for lab %s\n", int(left.Minutes()), int(left.Seconds())%60, labID)
		}
	}
}

func runRandomLab(cmd *cobra.Command) error {
	seed := time.Now().UnixNano()
	lab, err := labs.Random(seed, "", "", labs.CertAll)
	if err != nil {
		return err
	}
	cli.Info(fmt.Sprintf("Random lab selected: %s", lab.ID()))
	args := []string{lab.ID()}
	if cmd != nil {
		return labRunCmd.RunE(cmd, args)
	}
	return labRunCmd.RunE(rootCmd, args)
}

var labCleanCmd = &cobra.Command{
	Use:   "clean <lab-id>",
	Short: "Clean up lab resources",
	Long:  `Deletes resources created by a specific lab (namespaces, deployments, etc.).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		labID := args[0]
		lab, err := labs.Get(labID)
		if err != nil {
			return err
		}

		if err := loadConfig(); err != nil {
			return err
		}

		provider, err := createProvider()
		if err != nil {
			return fmt.Errorf("creating provider: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		exists, err := provider.Exists(ctx)
		if err != nil {
			return fmt.Errorf("checking if cluster exists: %w", err)
		}
		if !exists {
			return fmt.Errorf("cluster does not exist")
		}

		kubeconfigPath, err := provider.KubeconfigPath(ctx)
		if err != nil {
			return fmt.Errorf("getting kubeconfig: %w", err)
		}

		cli.Info(fmt.Sprintf("Cleaning resources for lab: %s", lab.Title()))

		kubectlExec := func(args ...string) {
			cmdArgs := append([]string{"--kubeconfig", kubeconfigPath}, args...)
			c := exec.CommandContext(ctx, "kubectl", cmdArgs...)
			c.Run()
		}

		kubectlExec("delete", "namespace", "lab", "--ignore-not-found")
		kubectlExec("delete", "namespace", "monitoring", "--ignore-not-found")
		kubectlExec("delete", "namespace", "logging", "--ignore-not-found")
		kubectlExec("delete", "namespace", "microservices", "--ignore-not-found")
		kubectlExec("delete", "namespace", "production", "--ignore-not-found")
		kubectlExec("delete", "namespace", "staging", "--ignore-not-found")
		kubectlExec("delete", "namespace", "secure-ns", "--ignore-not-found")

		cli.Success(fmt.Sprintf("Cleaned up resources for lab: %s", lab.ID()))
		return nil
	},
}

var labResetCmd = &cobra.Command{
	Use:   "reset [lab-id]",
	Short: "Reset lab progress",
	Long:  `Clears progress for a specific lab, or all labs if no ID is given.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if len(args) == 0 && !all {
			return fmt.Errorf("provide a lab ID or use --all to reset everything")
		}

		p := progress.Load()
		p.ResetProgress(args, all)

		if all {
			cli.Success("All progress has been reset.")
		} else {
			cli.Success(fmt.Sprintf("Progress for lab '%s' has been reset.", args[0]))
		}
		return nil
	},
}

var labPickCmd = &cobra.Command{
	Use:   "pick",
	Short: "Pick a lab interactively",
	Long:  `Opens an interactive selector to browse and pick a lab with arrow keys and mouse.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		allLabs := labs.List()

		selectedLab, err := cli.RunInteractiveLabSelector(allLabs)
		if err != nil {
			if err.Error() == "no lab selected" {
				cli.Info("No lab selected")
				return nil
			}
			return err
		}

		return labRunCmd.RunE(cmd, []string{selectedLab.ID()})
	},
}

var labResumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume the last active lab",
	Long:  `Re-applies the broken scenario for the last lab you were working on.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		labID := progress.ActiveLab()
		if labID == "" {
			return fmt.Errorf("no active lab to resume. Run 'cka-lab-runner lab run <id>' first")
		}
		cli.Info(fmt.Sprintf("Resuming lab: %s", labID))
		return labRunCmd.RunE(cmd, []string{labID})
	},
}

func init() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nInterrupted. Cleaning up...")
		os.Exit(1)
	}()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.DefaultConfigFile, "config file")

	upCmd.Flags().Bool("recreate", false, "Recreate the cluster if it already exists")
	upCmd.Flags().Bool("random", false, "Run a random lab after creating the cluster")
	upCmd.Flags().String("version", "", "KinD node image version (e.g., v1.35.0, v1.34.0, v1.33.0)")
	upCmd.Flags().Int("workers", 0, "Number of worker nodes (0 = single-node, default)")

	labListCmd.Flags().String("category", "", "Filter by category")
	labListCmd.Flags().String("difficulty", "", "Filter by difficulty")
	labListCmd.Flags().String("cert", "", "Filter by certification (CKA, CKAD, CKS)")
	labListCmd.Flags().String("domain", "", "Filter by CKA exam domain")
	labListCmd.Flags().String("tag", "", "Filter by tag")
	labListCmd.Flags().String("search", "", "Search labs by ID, title, description, or tags")
	labListCmd.Flags().Bool("json", false, "Output as JSON")
	labListCmd.Flags().Bool("csv", false, "Output as CSV")
	labListCmd.Flags().Bool("progress", false, "Show completion status next to each lab")

	labRunCmd.Flags().Bool("timed", false, "Enable timed challenge mode")
	labRunCmd.Flags().Bool("timer", false, "Enable exam simulation mode (2 hours, no hints)")
	labRunCmd.Flags().Int("time-limit", 0, "Time limit in minutes (default: lab estimated time)")
	labRunCmd.Flags().String("namespace", "", "Override target namespace for the lab")
	labRunCmd.Flags().Bool("force", false, "Run even if prerequisites are not completed")

	labRandomCmd.Flags().Int64("seed", 0, "Random seed for reproducible selection")
	labRandomCmd.Flags().String("category", "", "Filter by category")
	labRandomCmd.Flags().String("difficulty", "", "Filter by difficulty")
	labRandomCmd.Flags().String("cert", "", "Filter by certification (CKA, CKAD, CKS)")

	labHintCmd.Flags().Int("level", 1, "Hint level (1 = vague, 2 = moderate, 3 = specific)")

	labStatusCmd.Flags().Bool("json", false, "Output as JSON")

	labExportCmd.Flags().String("format", "json", "Export format: json, markdown, certificate")

	labExamCmd.Flags().String("cert", "CKA", "Certification to simulate (CKA, CKAD, CKS)")
	labExamCmd.Flags().Int("num-labs", 15, "Number of labs in the exam")

	labResetCmd.Flags().Bool("all", false, "Reset all lab progress")

	labListCmd.Flags().String("resource", "", "Filter by Kubernetes resource type (pod, service, pv, etc.)")

	rootCmd.PersistentFlags().String("theme", "", "Color theme: dark, light (auto-detect if empty)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(labCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(cli.NewCompletionCmd(rootCmd))

	labCmd.AddCommand(labListCmd)
	labCmd.AddCommand(labRunCmd)
	labCmd.AddCommand(labVerifyCmd)
	labCmd.AddCommand(labSolutionCmd)
	labCmd.AddCommand(labRandomCmd)
	labCmd.AddCommand(labHintCmd)
	labCmd.AddCommand(labStatusCmd)
	labCmd.AddCommand(labExportCmd)
	labCmd.AddCommand(labStreakCmd)
	labCmd.AddCommand(labRateCmd)
	labCmd.AddCommand(labExamCmd)
	labCmd.AddCommand(labPickCmd)
	labCmd.AddCommand(labCleanCmd)
	labCmd.AddCommand(labResetCmd)
	labCmd.AddCommand(labResumeCmd)
}

func loadConfig() error {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w (run 'cka-lab-runner init' to create one)", err)
	}
	return nil
}

func createProvider() (cluster.Provider, error) {
	return cluster.NewProvider(cluster.Config{
		Provider:          cfg.Cluster.Provider,
		Name:              cfg.Cluster.Name,
		KubernetesVersion: cfg.Cluster.KubernetesVersion,
		Workers:           cfg.Cluster.Workers,
	})
}

// provisionClusterForSpec ensures a running cluster that satisfies the lab's
// ClusterSpec. If the cluster is absent or its running version/node count does
// not match the spec, it is (re)created with the spec's settings.
func provisionClusterForSpec(cmd *cobra.Command, spec labs.ClusterSpec) error {
	// Apply the spec to the in-memory config so the provider is built with it.
	applySpec := func() {
		if spec.Provider != "" {
			cfg.Cluster.Provider = spec.Provider
		}
		if spec.KubernetesVersion != "" {
			cfg.Cluster.KubernetesVersion = spec.KubernetesVersion
		}
		if spec.Workers >= 0 {
			cfg.Cluster.Workers = spec.Workers
		}
	}
	applySpec()

	provider, err := createProvider()
	if err != nil {
		return fmt.Errorf("creating provider: %w", err)
	}
	labs.SetClusterName(provider.Name())

	pctx, pcancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer pcancel()

	exists, err := provider.Exists(pctx)
	if err != nil {
		return err
	}

	needsRecreate := false
	if exists && spec.KubernetesVersion != "" {
		// The cluster exists but may be running the wrong version — check it.
		version, vErr := runningClusterVersion(pctx, provider)
		if vErr != nil {
			cli.Warning(fmt.Sprintf("Could not read running cluster version: %v", vErr))
		}
		if strings.TrimPrefix(version, "v") != "" &&
			spec.KubernetesVersion != "" &&
			normalizeK8sVersion(version) != normalizeK8sVersion(spec.KubernetesVersion) {
			cli.Info(fmt.Sprintf("Lab requires cluster %s but it is running %s — recreating.",
				spec.KubernetesVersion, version))
			needsRecreate = true
		}
	}

	if exists && needsRecreate {
		cli.Info(fmt.Sprintf("Deleting cluster %s to match lab requirements...", provider.Name()))
		if err := provider.Down(pctx); err != nil {
			return fmt.Errorf("deleting cluster for lab spec: %w", err)
		}
		exists = false
	}

	if !exists {
		cli.Info(fmt.Sprintf("Provisioning cluster for this lab: %s (version %s, %d worker(s))",
			provider.Name(), cfg.Cluster.KubernetesVersion, cfg.Cluster.Workers))
		if err := provider.Up(pctx); err != nil {
			return fmt.Errorf("provisioning cluster for lab spec: %w", err)
		}
		cli.Success("Cluster ready for this lab.")
	}

	return nil
}

// runningClusterVersion reads the Kubernetes version of the control-plane node
// of an existing cluster via kubectl.
func runningClusterVersion(ctx context.Context, provider cluster.Provider) (string, error) {
	kc, err := provider.KubeconfigPath(ctx)
	if err != nil {
		return "", err
	}
	out, err := outputOf(exec.CommandContext(ctx, "kubectl",
		"--kubeconfig", kc, "version", "-o", "json"))
	if err != nil {
		return "", err
	}
	out = extractJSONObject(out)
	var v struct {
		ServerVersion struct {
			GitVersion string `json:"gitVersion"`
		} `json:"serverVersion"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		return "", err
	}
	return v.ServerVersion.GitVersion, nil
}

// extractJSONObject returns the substring between the first '{' and the last
// '}', so kubectl JSON output can be parsed even when WARNING lines are merged
// in before or after the JSON on the same captured stream.
func extractJSONObject(s string) string {
	if i := strings.Index(s, "{"); i >= 0 {
		s = s[i:]
	}
	if j := strings.LastIndex(s, "}"); j >= 0 {
		s = s[:j+1]
	}
	return strings.TrimSpace(s)
}

// runKubectl executes kubectl against the given kubeconfig and returns stdout.
func runKubectl(ctx context.Context, kubeconfig string, args ...string) (string, error) {
	cmdArgs := []string{"--kubeconfig", kubeconfig}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.CommandContext(ctx, "kubectl", cmdArgs...)
	return outputOf(cmd)
}

// outputOf runs a command and returns trimmed combined output (best effort on error).
func outputOf(cmd *exec.Cmd) (string, error) {
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// normalizeK8sVersion trims a leading "v" for comparison.
func normalizeK8sVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// provisionPendingNode ensures an extra unjoined node container exists for the
// cluster, provisioning it from a matching node image if needed.
func provisionPendingNode(ctx context.Context, clusterName string, pn labs.PendingNode) error {
	containerName := clusterName + "-" + pn.Name
	version := normalizeK8sVersion(pn.Version)
	if version == "" {
		version = normalizeK8sVersion(cfg.Cluster.KubernetesVersion)
	}
	image := "kindest/node:v" + version

	created, err := cluster.EnsureNodeContainer(ctx, containerName, image)
	if err != nil {
		return err
	}
	if created {
		cli.Info(fmt.Sprintf("Provisioned pending node container: %s", containerName))
	} else {
		cli.Info(fmt.Sprintf("Using existing pending node container: %s", containerName))
	}
	return nil
}
