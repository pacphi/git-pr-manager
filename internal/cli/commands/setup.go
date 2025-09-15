package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pacphi/git-pr-manager/pkg/utils"
	"github.com/pacphi/git-pr-manager/pkg/wizard"
)

// SetupFlags contains flags for setup commands
type SetupFlags struct {
	ConfigPath  string
	Interactive bool
	Preview     bool
	Additive    bool
	Backup      bool
}

// NewSetupCommand creates the setup command group
func NewSetupCommand() *cobra.Command {
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Set up Git PR automation configuration",
		Long: `Setup commands help you configure Git PR automation.

Available setup options:
- config: Copy sample configuration file
- wizard: Interactive repository discovery and configuration
- preview: Preview what the wizard would configure
- additive: Add repositories to existing configuration`,
	}

	// Add subcommands
	setupCmd.AddCommand(NewSetupConfigCommand())
	setupCmd.AddCommand(NewSetupWizardCommand())
	setupCmd.AddCommand(NewSetupPreviewCommand())
	setupCmd.AddCommand(NewSetupAdditiveCommand())
	setupCmd.AddCommand(NewSetupRestoreCommand())

	return setupCmd
}

// NewSetupConfigCommand creates the setup config command
func NewSetupConfigCommand() *cobra.Command {
	var flags SetupFlags

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Copy sample configuration file",
		Long:  "Copies config.sample to config.yaml for manual configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupConfig(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVarP(&flags.ConfigPath, "config", "c", "config.yaml", "configuration file path")
	cmd.Flags().BoolVar(&flags.Backup, "backup", true, "backup existing configuration before setup")

	return cmd
}

// NewSetupWizardCommand creates the setup wizard command
func NewSetupWizardCommand() *cobra.Command {
	var flags SetupFlags

	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Interactive repository discovery and configuration wizard",
		Long: `Interactive wizard that discovers repositories from your Git providers
and generates a complete configuration file based on your selections.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupWizard(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVarP(&flags.ConfigPath, "config", "c", "config.yaml", "configuration file path")
	cmd.Flags().BoolVar(&flags.Backup, "backup", true, "backup existing configuration before setup")
	cmd.Flags().BoolVarP(&flags.Interactive, "interactive", "i", true, "run in interactive mode")

	return cmd
}

// NewSetupPreviewCommand creates the setup preview command
func NewSetupPreviewCommand() *cobra.Command {
	var flags SetupFlags

	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview what the wizard would configure",
		Long:  "Shows what repositories would be configured without making changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			flags.Preview = true
			return runSetupWizard(cmd.Context(), flags)
		},
	}

	return cmd
}

// NewSetupAdditiveCommand creates the setup additive command
func NewSetupAdditiveCommand() *cobra.Command {
	var flags SetupFlags

	cmd := &cobra.Command{
		Use:   "additive",
		Short: "Add repositories to existing configuration",
		Long:  "Adds newly discovered repositories to your existing configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			flags.Additive = true
			return runSetupWizard(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVarP(&flags.ConfigPath, "config", "c", "config.yaml", "configuration file path")
	cmd.Flags().BoolVar(&flags.Backup, "backup", true, "backup existing configuration before setup")

	return cmd
}

// runSetupConfig copies the sample configuration file
func runSetupConfig(_ context.Context, flags SetupFlags) error {
	logger := utils.GetGlobalLogger()

	samplePath := "config.sample"
	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		return fmt.Errorf("config.sample not found in current directory")
	}

	// Check if config already exists
	if _, err := os.Stat(flags.ConfigPath); err == nil {
		logger.Warnf("Configuration file %s already exists", flags.ConfigPath)

		// Create backup if requested using the proper BackupConfig function
		if flags.Backup {
			if err := BackupConfig(flags.ConfigPath); err != nil {
				logger.Warnf("Failed to backup configuration: %v", err)
			} else {
				logger.Info("Configuration backed up successfully")
			}
		}

		// Ask for confirmation
		fmt.Printf("Configuration file %s already exists. Overwrite? [y/N]: ", flags.ConfigPath)
		var response string
		_, _ = fmt.Scanln(&response) // Ignore read errors, treat as "no"
		if response != "y" && response != "Y" && response != "yes" {
			logger.Info("Setup cancelled")
			return nil
		}
	}

	// Read sample file
	data, err := os.ReadFile(samplePath)
	if err != nil {
		return fmt.Errorf("failed to read config.sample: %w", err)
	}

	// Write to config file
	err = os.WriteFile(flags.ConfigPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	logger.Infof("Configuration file copied to %s", flags.ConfigPath)
	logger.Info("Please edit the configuration file with your settings")
	logger.Info("Run 'git-pr-cli validate' to check your configuration")

	return nil
}

// runSetupWizard runs the interactive setup wizard
func runSetupWizard(ctx context.Context, flags SetupFlags) error {
	logger := utils.GetGlobalLogger()

	// Create wizard
	wizardConfig := wizard.Config{
		ConfigPath:  flags.ConfigPath,
		Preview:     flags.Preview,
		Additive:    flags.Additive,
		Interactive: flags.Interactive,
	}

	w := wizard.New(wizardConfig)

	if flags.Preview {
		logger.Info("Running setup wizard in preview mode...")
		return w.Preview(ctx)
	}

	if flags.Additive {
		logger.Info("Running additive setup wizard...")
		return w.RunAdditive(ctx)
	}

	logger.Info("Running interactive setup wizard...")
	return w.Run(ctx)
}

// BackupConfig creates a backup of the current configuration
func BackupConfig(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // No existing config to backup
	}

	// Create backup directory
	backupDir := filepath.Join(filepath.Dir(configPath), ".backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename with timestamp
	basename := filepath.Base(configPath)
	timestamp := utils.FormatTimestamp()
	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.%s.bak", basename, timestamp))

	// Copy file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = os.WriteFile(backupPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	logger := utils.GetGlobalLogger()
	logger.Infof("Configuration backed up to %s", backupPath)

	return nil
}

// RestoreConfig restores configuration from the latest backup
func RestoreConfig(configPath string) error {
	backupDir := filepath.Join(filepath.Dir(configPath), ".backups")
	basename := filepath.Base(configPath)

	// Find latest backup
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	var latestBackup string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".bak" && strings.HasPrefix(name, basename) {
			if latestBackup == "" || name > latestBackup {
				latestBackup = name
			}
		}
	}

	if latestBackup == "" {
		return fmt.Errorf("no backup files found")
	}

	backupPath := filepath.Join(backupDir, latestBackup)

	// Read backup
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Write to config
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to restore configuration: %w", err)
	}

	logger := utils.GetGlobalLogger()
	logger.Infof("Configuration restored from %s", backupPath)

	return nil
}

// NewSetupRestoreCommand creates the setup restore command
func NewSetupRestoreCommand() *cobra.Command {
	var flags SetupFlags

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore configuration from the latest backup",
		Long:  "Restores configuration from the most recent backup file in the .backups directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupRestore(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVarP(&flags.ConfigPath, "config", "c", "config.yaml", "configuration file path to restore")

	return cmd
}

// runSetupRestore restores configuration from the latest backup
func runSetupRestore(_ context.Context, flags SetupFlags) error {
	return RestoreConfig(flags.ConfigPath)
}
