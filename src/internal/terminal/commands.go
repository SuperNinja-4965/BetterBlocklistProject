package terminal

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Global cancellation context
var (
	globalCancel context.CancelFunc
	globalMutex  sync.Mutex
)

// Command represents a single command to execute
type Command struct {
	Name        string
	Binary      string
	Args        []string
	Description string
}

// Cancel currently running command
func cancelCurrentCommand() {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	if globalCancel != nil {
		globalCancel()
		globalCancel = nil
	}
}

// Execute a single command in detached mode (fire and forget)
func ExecuteDetachedCommand(command string, args []string) tea.Cmd {
	return func() tea.Msg {
		// Parse the command and arguments
		command = strings.TrimSpace(command)
		if command == "" {
			return OutputLine{
				Line: "Error: No command provided",
			}
		}

		// Start the command in detached mode
		cmd := exec.Command(command, args...)

		// Detach from parent process (Unix-specific)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}

		// Redirect stdout/stderr to prevent hanging
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil

		// Start the process
		err := cmd.Start()
		if err != nil {
			return OutputLine{
				Line: fmt.Sprintf("Error starting detached command '%s': %s\n\nTip: Make sure the command exists and is executable.", command, err.Error()),
			}
		}

		// Don't wait for the process - let it run independently
		go func() {
			_ = cmd.Wait() // Clean up the process when it finishes (ignore error)
		}()

		return OutputLine{
			Line: fmt.Sprintf("✅ Command started successfully in detached mode!\n\nCommand: %s\nProcess ID: %d\n\nThe process is now running independently of this application.\nYou can close this app and the command will continue running.", command, cmd.Process.Pid),
		}
	}
}

// Execute multiple commands sequentially with live output streaming and cancellation support
func ExecuteOSCommandStreaming(choice string, commands []Command) tea.Cmd {
	return tea.Batch(
		// First, set the command as running
		func() tea.Msg {
			return commandProgress{
				lines: []string{"Starting " + choice + "...", fmt.Sprintf("Executing %d command(s) sequentially:", len(commands))},
				done:  false,
			}
		},
		// Then start the actual execution
		ExecuteCommandsWithCancellation(choice, commands),
	)
}

// Execute commands with cancellation support using context
func ExecuteCommandsWithCancellation(choice string, commands []Command) tea.Cmd {
	return func() tea.Msg {
		// Create a context that can be cancelled
		ctx, cancel := context.WithCancel(context.Background())

		// Set the global cancel function
		globalMutex.Lock()
		globalCancel = cancel
		globalMutex.Unlock()

		defer func() {
			globalMutex.Lock()
			globalCancel = nil
			globalMutex.Unlock()
			cancel()
		}()

		var allOutput []string
		allOutput = append(allOutput, "Starting "+choice+"...")
		allOutput = append(allOutput, fmt.Sprintf("Executing %d command(s) sequentially:\n", len(commands)))

		// Execute each command sequentially
		for i, cmd := range commands {
			// Check if cancellation was requested
			select {
			case <-ctx.Done():
				allOutput = append(allOutput, "", "Command execution cancelled.")
				return OutputLine{
					Line: strings.Join(allOutput, "\n"),
				}
			default:
				// Continue with command execution
			}

			allOutput = append(allOutput, fmt.Sprintf("[%d/%d] %s: %s", i+1, len(commands), cmd.Name, cmd.Description))
			allOutput = append(allOutput, fmt.Sprintf("Running command: %s %s\n", cmd.Binary, strings.Join(cmd.Args, " ")))

			// Execute the command with context for cancellation
			execCmd := exec.CommandContext(ctx, cmd.Binary, cmd.Args...)

			// Get stdout pipe
			stdout, err := execCmd.StdoutPipe()
			if err != nil {
				allOutput = append(allOutput, fmt.Sprintf("Error creating stdout pipe for %s: %s", cmd.Name, err.Error()))
				continue
			}

			// Get stderr pipe
			stderr, err := execCmd.StderrPipe()
			if err != nil {
				allOutput = append(allOutput, fmt.Sprintf("Error creating stderr pipe for %s: %s", cmd.Name, err.Error()))
				continue
			}

			// Start the command
			if err := execCmd.Start(); err != nil {
				allOutput = append(allOutput, fmt.Sprintf("Error starting %s: %s", cmd.Name, err.Error()))
				continue
			}

			// Create channels to handle concurrent reading
			outputChan := make(chan string, 100)
			doneChan := make(chan bool)

			// Read stdout in goroutine
			go func() {
				defer close(outputChan)
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					select {
					case <-ctx.Done():
						return
					case outputChan <- scanner.Text():
					}
				}
			}()

			// Read stderr in goroutine
			go func() {
				scanner := bufio.NewScanner(stderr)
				for scanner.Scan() {
					select {
					case <-ctx.Done():
						return
					case outputChan <- scanner.Text():
					}
				}
			}()

			// Wait for command completion in goroutine
			go func() {
				_ = execCmd.Wait() // Ignore error here as we check ProcessState below
				close(doneChan)
			}()

			// Collect output until command completes or is cancelled
		readLoop:
			for {
				select {
				case <-ctx.Done():
					// Cancel the command and break
					if execCmd.Process != nil {
						_ = execCmd.Process.Kill() // Ignore error as process may already be dead
					}
					allOutput = append(allOutput, "Command cancelled")
					break readLoop
				case line, ok := <-outputChan:
					if !ok {
						continue
					}
					allOutput = append(allOutput, line)
				case <-doneChan:
					// Command completed, drain remaining output
					for {
						select {
						case line, ok := <-outputChan:
							if !ok {
								break readLoop
							}
							allOutput = append(allOutput, line)
						case <-time.After(100 * time.Millisecond):
							break readLoop
						}
					}
				}
			}

			// Check exit status
			if execCmd.ProcessState != nil {
				if execCmd.ProcessState.Success() {
					allOutput = append(allOutput, fmt.Sprintf("Command %s completed successfully!", cmd.Name))
				} else {
					allOutput = append(allOutput, fmt.Sprintf("Command %s failed with exit code: %d", cmd.Name, execCmd.ProcessState.ExitCode()))
				}
			}

			// Add separator between commands (except for the last one)
			if i < len(commands)-1 {
				allOutput = append(allOutput, "")
			}
		}

		allOutput = append(allOutput, "", "All commands completed!")

		return commandProgress{
			lines: allOutput,
			done:  true,
		}
	}
}
