package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/cli"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/cluster"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/config"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/labs"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/update"

	// Import labs to register them
	_ "github.com/WhoisMonesh/K8S-Lab-Everything/internal/labs"
)

var (
	cfgFile string
	cfg     *config.Config
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "cka-lab-runner",
	Short: "A CKA practice lab runner",
	Long: `cka-lab-runner is a tool for practicing Kubernetes administration skills
by creating reproducible broken scenarios in a local cluster.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip update check for version/update commands
		if cmd.Name() == "version" || cmd.Name() == "update" || cmd.Name() == "help" {
			return
		}
		update.CheckForUpdate()
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

		// Check if config already exists
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("config file already exists: %s", configPath)
		}

		// Create default config
		defaultCfg := config.Default()

		// Save config
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
		recreate, err := cmd.Flags().GetBool("recreate")
		if err != nil {
			return fmt.Errorf("getting recreate flag: %w", err)
		}

		// Load config
		if err := loadConfig(); err != nil {
			return err
		}

		// Create provider
		provider, err := createProvider()
		if err != nil {
			return fmt.Errorf("creating provider: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		// Check if cluster exists
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
				return nil
			}
		}

		// Create cluster
		cli.Info(fmt.Sprintf("Creating cluster: %s", provider.Name()))
		if err := provider.Up(ctx); err != nil {
			return fmt.Errorf("creating cluster: %w", err)
		}

		cli.Success(fmt.Sprintf("Cluster created: %s", provider.Name()))
		return nil
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Delete the local Kubernetes cluster",
	Long:  `Deletes the local Kubernetes cluster.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		if err := loadConfig(); err != nil {
			return err
		}

		// Create provider
		provider, err := createProvider()
		if err != nil {
			return fmt.Errorf("creating provider: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Delete cluster
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
		categoryFilter, err := cmd.Flags().GetString("category")
		if err != nil {
			return fmt.Errorf("getting category flag: %w", err)
		}
		difficultyFilter, err := cmd.Flags().GetString("difficulty")
		if err != nil {
			return fmt.Errorf("getting difficulty flag: %w", err)
		}

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
			if matches {
				filteredLabs = append(filteredLabs, lab)
			}
		}

		cli.PrintLabList(filteredLabs)
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

		// Get the lab
		lab, err := labs.Get(labID)
		if err != nil {
			return err
		}

		// Load config
		if err := loadConfig(); err != nil {
			return err
		}

		// Create provider
		provider, err := createProvider()
		if err != nil {
			return fmt.Errorf("creating provider: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Check if cluster exists
		exists, err := provider.Exists(ctx)
		if err != nil {
			return fmt.Errorf("checking if cluster exists: %w", err)
		}

		if !exists {
			return fmt.Errorf("cluster does not exist. Run 'cka-lab-runner up' first")
		}

		// Get kubeconfig
		kubeconfigPath, err := provider.KubeconfigPath(ctx)
		if err != nil {
			return fmt.Errorf("getting kubeconfig: %w", err)
		}

		// Prepare the lab
		cli.Info("Preparing lab environment...")
		if err := lab.Prepare(ctx, kubeconfigPath); err != nil {
			cli.Warning(fmt.Sprintf("Prepare step failed (may be optional): %v", err))
		}

		// Break the cluster
		cli.Info("Applying broken scenario...")
		if err := lab.Break(ctx, kubeconfigPath); err != nil {
			return fmt.Errorf("breaking cluster: %w", err)
		}

		// Verify broken state
		if err := lab.VerifyBroken(ctx, kubeconfigPath); err != nil {
			cli.Warning(fmt.Sprintf("Verify broken step failed (may be optional): %v", err))
		}

		// Print lab details
		cli.PrintLabDetails(lab)

		cli.Success("Lab scenario applied successfully!")
		cli.Info(fmt.Sprintf("Use 'cka-lab-runner lab solution %s' to see the solution", labID))
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

		// Get the lab
		lab, err := labs.Get(labID)
		if err != nil {
			return err
		}

		// Print solution
		solution := labs.FormatSolution(lab)
		fmt.Println(solution)

		return nil
	},
}

var labRandomCmd = &cobra.Command{
	Use:   "random",
	Short: "Select a random lab",
	Long:  `Selects and runs a random lab, optionally filtered by category and difficulty.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		seed, err := cmd.Flags().GetInt64("seed")
		if err != nil {
			return fmt.Errorf("getting seed flag: %w", err)
		}
		categoryFilter, err := cmd.Flags().GetString("category")
		if err != nil {
			return fmt.Errorf("getting category flag: %w", err)
		}
		difficultyFilter, err := cmd.Flags().GetString("difficulty")
		if err != nil {
			return fmt.Errorf("getting difficulty flag: %w", err)
		}

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

		lab, err := labs.Random(seed, category, difficulty)
		if err != nil {
			return err
		}

		cli.Info(fmt.Sprintf("Selected lab: %s", lab.ID()))

		// Run the lab by calling the run command
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

		// Get the lab
		lab, err := labs.Get(labID)
		if err != nil {
			return err
		}

		// Load config
		if err := loadConfig(); err != nil {
			return err
		}

		// Create provider
		provider, err := createProvider()
		if err != nil {
			return fmt.Errorf("creating provider: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		// Check if cluster exists
		exists, err := provider.Exists(ctx)
		if err != nil {
			return fmt.Errorf("checking if cluster exists: %w", err)
		}

		if !exists {
			return fmt.Errorf("cluster does not exist")
		}

		// Get kubeconfig
		kubeconfigPath, err := provider.KubeconfigPath(ctx)
		if err != nil {
			return fmt.Errorf("getting kubeconfig: %w", err)
		}

		// Verify the fix
		cli.Info(fmt.Sprintf("Verifying lab: %s", lab.Title()))
		if err := lab.Verify(ctx, kubeconfigPath); err != nil {
			cli.Error(fmt.Sprintf("Lab not fixed yet: %v", err))
			cli.Info("Keep trying! Use 'cka-lab-runner lab solution' if you need help")
			return nil
		}

		cli.Success(fmt.Sprintf("Congratulations! You successfully fixed: %s", lab.Title()))
		return nil
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.DefaultConfigFile, "config file")

	// up command flags
	upCmd.Flags().Bool("recreate", false, "Recreate the cluster if it already exists")

	// lab list command flags
	labListCmd.Flags().String("category", "", "Filter by category")
	labListCmd.Flags().String("difficulty", "", "Filter by difficulty")

	// lab random command flags
	labRandomCmd.Flags().Int64("seed", 0, "Random seed for reproducible selection")
	labRandomCmd.Flags().String("category", "", "Filter by category")
	labRandomCmd.Flags().String("difficulty", "", "Filter by difficulty")

	// Add commands to root
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(labCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)

	// Add subcommands to lab
	labCmd.AddCommand(labListCmd)
	labCmd.AddCommand(labRunCmd)
	labCmd.AddCommand(labSolutionCmd)
	labCmd.AddCommand(labRandomCmd)
	labCmd.AddCommand(labVerifyCmd)
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
