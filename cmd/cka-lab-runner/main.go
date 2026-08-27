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

	bold  = "\033[1m"
	dimW  = "\033[90m"
	reset = "\033[0m"
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
		update.CheckForUpdate()
	},
	Run: func(cmd *cobra.Command, args []string) {
		cli.PrintBanner()
		fmt.Println("  Quick Start:")
		fmt.Println()
		fmt.Printf("    %scka-lab-runner init%s              %sCreate config file%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner up%s                %sCreate local cluster%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab list%s          %sList all labs%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab list --cert CKA%s  %sFilter by certification%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab pick%s          %sPick interactively%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab run <id>%s      %sStart a lab%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab verify <id>%s   %sCheck your fix%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab solution <id>%s %sShow solution%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner lab status%s        %sView progress%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner down%s              %sDelete cluster%s\n", bold, reset, dimW, reset)
		fmt.Printf("    %scka-lab-runner update%s            %sUpdate tool%s\n", bold, reset, dimW, reset)
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

		if err := loadConfig(); err != nil {
			return err
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

		cli.Info(fmt.Sprintf("Creating cluster: %s", provider.Name()))
		if err := provider.Up(ctx); err != nil {
			return fmt.Errorf("creating cluster: %w", err)
		}

		cli.Success(fmt.Sprintf("Cluster created: %s", provider.Name()))

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

		if timed {
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

var labExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export completion history as JSON",
	Long:  `Exports your lab completion history to stdout as JSON.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := progress.ExportJSON()
		if err != nil {
			return fmt.Errorf("exporting progress: %w", err)
		}
		fmt.Println(string(data))
		return nil
	},
}

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

	labListCmd.Flags().String("category", "", "Filter by category")
	labListCmd.Flags().String("difficulty", "", "Filter by difficulty")
	labListCmd.Flags().String("cert", "", "Filter by certification (CKA, CKAD, CKS)")
	labListCmd.Flags().String("domain", "", "Filter by CKA exam domain")
	labListCmd.Flags().String("tag", "", "Filter by tag")
	labListCmd.Flags().Bool("json", false, "Output as JSON")
	labListCmd.Flags().Bool("progress", false, "Show completion status next to each lab")

	labRunCmd.Flags().Bool("timed", false, "Enable timed challenge mode")
	labRunCmd.Flags().Int("time-limit", 0, "Time limit in minutes (default: lab estimated time)")
	labRunCmd.Flags().String("namespace", "", "Override target namespace for the lab")
	labRunCmd.Flags().Bool("force", false, "Run even if prerequisites are not completed")

	labRandomCmd.Flags().Int64("seed", 0, "Random seed for reproducible selection")
	labRandomCmd.Flags().String("category", "", "Filter by category")
	labRandomCmd.Flags().String("difficulty", "", "Filter by difficulty")
	labRandomCmd.Flags().String("cert", "", "Filter by certification (CKA, CKAD, CKS)")

	labHintCmd.Flags().Int("level", 1, "Hint level (1 = vague, 2 = moderate, 3 = specific)")

	labStatusCmd.Flags().Bool("json", false, "Output as JSON")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(labCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)

	labCmd.AddCommand(labListCmd)
	labCmd.AddCommand(labRunCmd)
	labCmd.AddCommand(labSolutionCmd)
	labCmd.AddCommand(labRandomCmd)
	labCmd.AddCommand(labVerifyCmd)
	labCmd.AddCommand(labHintCmd)
	labCmd.AddCommand(labStatusCmd)
	labCmd.AddCommand(labExportCmd)
	labCmd.AddCommand(labPickCmd)
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
	})
}
