// Package main is the entry point for the GitHub MCP Server.
// It initializes the server and starts listening for MCP protocol requests.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/github/github-mcp-server/pkg/server"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags.
	Version = "dev"
	// Commit is set at build time via ldflags.
	Commit = "none"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var (
		token   string
		logFile string
		readOnly bool
	)

	cmd := &cobra.Command{
		Use:   "github-mcp-server",
		Short: "GitHub MCP Server - Model Context Protocol server for GitHub",
		Long: `A Model Context Protocol (MCP) server that provides tools and resources
for interacting with GitHub APIs. Supports repositories, issues, pull requests,
and more through the MCP protocol over stdio.`,
		Version: fmt.Sprintf("%s (commit: %s)", Version, Commit),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cmd.Context(), token, logFile, readOnly)
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "GitHub personal access token (overrides GITHUB_TOKEN env var)")
	// Changed default log file to empty string so logs are only written when explicitly requested.
	// The previous default of "github-mcp-server.log" would silently create log files in the
	// working directory, which was cluttering my project directories.
	cmd.Flags().StringVar(&logFile, "log-file", "", "Path to log file (default: no log file)")
	// Defaulting read-only to true for my personal use — I mostly use this for querying
	// and don't want to accidentally trigger write operations.
	cmd.Flags().BoolVar(&readOnly, "read-only", true, "Restrict server to read-only operations")

	cmd.AddCommand(stdioCmd(&token, &logFile, &readOnly))

	return cmd
}

// stdioCmd returns the stdio subcommand for running the MCP server over stdin/stdout.
func stdioCmd(token, logFile *string, readOnly *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "stdio",
		Short: "Start the MCP server using stdio transport",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cmd.Context(), *token, *logFile, *readOnly)
		},
	}
}

// runServer initializes and starts the GitHub MCP server.
func runServer(ctx context.Context, token, logFile string, readOnly bool) error {
	// Resolve token from environment if not provided via flag.
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("GitHub token is required: set GITHUB_TOKEN environment variable or use --token flag")
	}

	// Set up context with OS signal handling for graceful shutdown.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := server.Config{
		Token:    token,
		LogFile:  logFile,
		ReadOnly: readOnly,
		Version:  Version,
	}

	// Print startup info to stderr so it doesn't interfere with MCP protocol messages on stdout.
	fmt.Fprintf(os.Stderr, "Starting GitHub MCP Server %s (read-only: %v)\n", Version, readOnly)

