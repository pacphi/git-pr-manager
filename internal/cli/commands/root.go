package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cphillipson/multi-gitter-pr-automation/pkg/config"
	"github.com/cphillipson/multi-gitter-pr-automation/pkg/utils"
)

// GlobalFlags contains global command line flags
type GlobalFlags struct {
	ConfigFile string
	LogLevel   string
	LogFormat  string
	Verbose    bool
	Quiet      bool
	Debug      bool
}

// NewRootCommand creates the root CLI command
func NewRootCommand(version, buildTime, commitSHA string) *cobra.Command {
	var globalFlags GlobalFlags

	rootCmd := &cobra.Command{
		Use:   "git-pr-cli",
		Short: "Git PR automation CLI tool",
		Long: `Git PR CLI is a tool for automating pull request management across multiple repositories.

It supports GitHub, GitLab, and Bitbucket, focusing on safely merging dependency updates
from trusted bots like dependabot and renovate.`,
		Version:      version,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return setupGlobalConfig(globalFlags)
		},
	}

	// Set version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(`{{.Name}} version %s
Build time: %s
Commit: %s
`, version, buildTime, commitSHA))

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&globalFlags.ConfigFile, "config", "c", "", "config file (default is config.yaml)")
	rootCmd.PersistentFlags().StringVar(&globalFlags.LogLevel, "log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&globalFlags.LogFormat, "log-format", "text", "log format (text, json)")
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.Quiet, "quiet", "q", false, "quiet output")
	rootCmd.PersistentFlags().BoolVar(&globalFlags.Debug, "debug", false, "enable debug logging (equivalent to --log-level debug)")

	// Bind flags to viper
	_ = viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("log_format", rootCmd.PersistentFlags().Lookup("log-format"))

	// Add subcommands
	rootCmd.AddCommand(NewCheckCommand())
	rootCmd.AddCommand(NewMergeCommand())
	rootCmd.AddCommand(NewSetupCommand())
	rootCmd.AddCommand(NewValidateCommand())
	rootCmd.AddCommand(NewStatsCommand())
	rootCmd.AddCommand(NewWatchCommand())
	rootCmd.AddCommand(NewTestCommand())
	rootCmd.AddCommand(NewInfoCommand())

	return rootCmd
}

// setupGlobalConfig configures global settings based on flags
func setupGlobalConfig(flags GlobalFlags) error {
	// Set log level (debug flag takes precedence)
	if flags.Debug {
		_ = os.Setenv("LOG_LEVEL", "debug")
	} else if flags.Verbose {
		_ = os.Setenv("LOG_LEVEL", "debug")
	} else if flags.Quiet {
		_ = os.Setenv("LOG_LEVEL", "error")
	} else if flags.LogLevel != "" {
		_ = os.Setenv("LOG_LEVEL", flags.LogLevel)
	}

	// Set log format
	if flags.LogFormat != "" {
		_ = os.Setenv("LOG_FORMAT", flags.LogFormat)
	}

	// Reinitialize logger with new settings
	logger := utils.NewLogger()
	utils.SetGlobalLogger(logger)

	// Set config file
	if flags.ConfigFile != "" {
		viper.SetConfigFile(flags.ConfigFile)
	} else {
		// Look for config in current directory and home directory
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		if home, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(filepath.Join(home, ".config", "git-pr"))
		}
		viper.AddConfigPath("/etc/git-pr")
	}

	return nil
}

// LoadConfig loads the configuration file
func LoadConfig() (*config.Config, error) {
	loader := config.NewLoader()
	return loader.Load(viper.ConfigFileUsed())
}

