// Package executor provides command execution functionality with proper error handling and security measures.
package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/cphillipson/multi-gitter-pr-automation/mcp-server/internal/types"
)

// Executor handles command execution with proper error handling and security
type Executor struct {
	workingDir string
	timeout    time.Duration
}

// NewExecutor creates a new command executor
func NewExecutor(workingDir string) *Executor {
	if workingDir == "" {
		workingDir = "."
	}
	return &Executor{
		workingDir: workingDir,
		timeout:    5 * time.Minute, // Default timeout
	}
}

// SetTimeout sets the command execution timeout
func (e *Executor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// ExecuteMake runs a make command with the specified target and arguments
func (e *Executor) ExecuteMake(target string, args ...string) *types.CommandResult {
	// Build the make command
	cmdArgs := []string{target}
	cmdArgs = append(cmdArgs, args...)

	return e.executeCommand("make", cmdArgs...)
}

// ExecuteScript runs a shell script
func (e *Executor) ExecuteScript(scriptName string, args ...string) *types.CommandResult {
	// Ensure script exists and is executable
	scriptPath := filepath.Join(e.workingDir, scriptName)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return &types.CommandResult{
			Success:  false,
			Error:    fmt.Sprintf("script not found: %s", scriptName),
			ExitCode: 127,
		}
	}

	cmdArgs := append([]string{scriptPath}, args...)
	return e.executeCommand("bash", cmdArgs...)
}

// executeCommand is the core command execution function
func (e *Executor) executeCommand(command string, args ...string) *types.CommandResult {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	// Create the command
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = e.workingDir

	// Set up environment variables (inherit from parent)
	cmd.Env = os.Environ()

	// Execute command and capture output
	output, err := cmd.CombinedOutput()

	result := &types.CommandResult{
		Output: strings.TrimSpace(string(output)),
	}

	// Handle different types of errors
	if err != nil {
		result.Success = false
		result.Error = e.handleError(ctx, err)
		result.ExitCode = e.getExitCode(ctx, err)
		 return result
	}
	
	result.Success = true
	result.ExitCode = 0
	return result
}

// handleError processes different types of command errors
func (e *Executor) handleError(ctx context.Context, err error) string {
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf("command timed out after %v", e.timeout)
	}
	return err.Error()
}

// getExitCode extracts the exit code from command errors
func (e *Executor) getExitCode(ctx context.Context, err error) int {
	if ctx.Err() == context.DeadlineExceeded {
		return 124 // timeout exit code
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 1
}

// ValidateEnvironment checks if required environment variables are set
func (e *Executor) ValidateEnvironment() []string {
	var missing []string

	requiredEnvVars := []string{
		"GITHUB_TOKEN",
		"GITLAB_TOKEN",
		"BITBUCKET_USERNAME",
		"BITBUCKET_APP_PASSWORD",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	return missing
}

// CheckDependencies verifies that required tools are installed
func (e *Executor) CheckDependencies() *types.CommandResult {
	return e.ExecuteMake("check-deps")
}

// GetWorkingDir returns the current working directory
func (e *Executor) GetWorkingDir() string {
	absPath, _ := filepath.Abs(e.workingDir)
	return absPath
}
