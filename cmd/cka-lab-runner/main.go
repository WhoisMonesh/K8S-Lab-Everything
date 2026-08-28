package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

var rootCmd = &cobra.Command{
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

		exists, err := provider.Exists(ctx)
		if err != nil {
			return fmt.Errorf("checking if cluster exists: %w", err)
		}
		if !exists {
			return fmt.Errorf("cluster does not exist. Run 'cka-lab-runner up' first")
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

		// Exam simulation mode (2 hours)
		if timer {
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
		if err := lab.Verify(ctx, kubeconfigPath); err != nil {
			cli.Error(fmt.Sprintf("Lab not fixed yet: %v", err))
			cli.Info(fmt.Sprintf("Keep trying! Use 'cka-lab-runner lab hint %s' for help", labID))
			return nil
		}

		cli.Success(fmt.Sprintf("Congratulations! You successfully fixed: %s", lab.Title()))

		if !progress.IsCompleted(labID) {
			progress.RecordCompletion(
				labID,
				lab.Title(),
				string(lab.Category()),
				string(lab.Difficulty()),
				0,
				lab.EstimatedTime(),
				false, false, "",
			)
			cli.Info("Progress saved! Run 'cka-lab-runner lab status' to see your progress.")
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

		fmt.Printf("  %s▸%s Starting first lab...\n\n", brCyan, reset)
		return labRunCmd.RunE(cmd, []string{plan[0].Lab.ID()})
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

	labListCmd.Flags().String("resource", "", "Filter by Kubernetes resource type (pod, service, pv, etc.)")

	rootCmd.PersistentFlags().String("theme", "", "Color theme: dark, light (auto-detect if empty)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(labCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)

	completionCmd = cli.NewCompletionCmd(rootCmd)
	rootCmd.AddCommand(completionCmd)

	labCmd.AddCommand(labListCmd)
	labCmd.AddCommand(labRunCmd)
	labCmd.AddCommand(labSolutionCmd)
	labCmd.AddCommand(labRandomCmd)
	labCmd.AddCommand(labVerifyCmd)
	labCmd.AddCommand(labHintCmd)
	labCmd.AddCommand(labStatusCmd)
	labCmd.AddCommand(labStatsCmd)
	labCmd.AddCommand(labExportCmd)
	labCmd.AddCommand(labPickCmd)
	labCmd.AddCommand(labStreakCmd)
	labCmd.AddCommand(labRateCmd)
	labCmd.AddCommand(labExamCmd)
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
