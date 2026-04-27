// Package github provides MCP server functionality for interacting with the GitHub API.
package github

import (
	"context"
	"fmt"

	"github.com/github/github-mcp-server/pkg/toolsets"
	"github.com/go-github/go-github/v62/github"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/oauth2"
)

// NewServer creates a new MCP server configured with GitHub tools.
// It accepts a GitHub personal access token used to authenticate API requests.
func NewServer(token string, opts ...server.ServerOption) (*server.MCPServer, error) {
	// Create an authenticated GitHub client
	ghClient, err := newGitHubClient(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Initialize the MCP server with GitHub-specific metadata
	s := server.NewMCPServer(
		"github-mcp-server",
		"0.1.0",
		opts...,
	)

	// Register all GitHub toolsets
	if err := registerToolsets(s, ghClient); err != nil {
		return nil, fmt.Errorf("failed to register toolsets: %w", err)
	}

	return s, nil
}

// newGitHubClient creates an authenticated GitHub API client using the provided token.
func newGitHubClient(token string) (*github.Client, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token must not be empty")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, ts)

	return github.NewClient(httpClient), nil
}

// registerToolsets registers all available GitHub toolsets with the MCP server.
func registerToolsets(s *server.MCPServer, client *github.Client) error {
	sets := []toolsets.Toolset{
		toolsets.NewRepositoryToolset(client),
		toolsets.NewIssueToolset(client),
		toolsets.NewPullRequestToolset(client),
		toolsets.NewSearchToolset(client),
	}

	for _, ts := range sets {
		if err := ts.Register(s); err != nil {
			return fmt.Errorf("failed to register toolset %q: %w", ts.Name(), err)
		}
	}

	return nil
}
