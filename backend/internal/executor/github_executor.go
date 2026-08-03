package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/iashu2k/mcp-gateway/backend/internal/domain"
)

var ErrGitHubUpstream = errors.New("github upstream request failed")

type GitHubExecutor struct {
	client *github.Client
}

func NewGitHubExecutor(token string) *GitHubExecutor {
	client := github.NewClient(nil)
	if token != "" {
		client = client.WithAuthToken(token)
	}

	return &GitHubExecutor{client: client}
}

func (e *GitHubExecutor) Execute(
	ctx context.Context,
	_ domain.MCPServer,
	tool domain.MCPTool,
	arguments json.RawMessage,
) (json.RawMessage, error) {
	switch tool.Name {
	case "list_issues":
		return e.listIssues(ctx, arguments)
	case "search_repositories":
    return e.searchRepositories(ctx, arguments)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedTool, tool.Name)
	}
}

func (e *GitHubExecutor) listIssues(
	ctx context.Context,
	arguments json.RawMessage,
) (json.RawMessage, error) {
	var input struct {
		Owner   string `json:"owner"`
		Repo    string `json:"repo"`
		State   string `json:"state"`
		PerPage int    `json:"per_page"`
	}

	if err := json.Unmarshal(arguments, &input); err != nil {
		return nil, fmt.Errorf("decode list_issues arguments: %w", err)
	}

	if input.State == "" {
		input.State = "open"
	}

	if input.PerPage <= 0 || input.PerPage > 100 {
		input.PerPage = 30
	}

	opts := &github.IssueListByRepoOptions{
		State: input.State,
		ListOptions: github.ListOptions{
			PerPage: input.PerPage,
		},
	}

	issues, _, err := e.client.Issues.ListByRepo(
		ctx,
		input.Owner,
		input.Repo,
		opts,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGitHubUpstream, err)
	}

	// Slim the payload: the audit table should not store full API responses.
	items := make([]map[string]any, 0, len(issues))
	for _, issue := range issues {
		if issue.IsPullRequest() {
			continue // /issues returns PRs too; keep this tool issue-only
		}

		items = append(items, map[string]any{
			"number": issue.GetNumber(),
			"title":  issue.GetTitle(),
			"state":  issue.GetState(),
			"user":   issue.GetUser().GetLogin(),
			"url":    issue.GetHTMLURL(),
		})
	}

	return json.Marshal(map[string]any{
		"repository": input.Owner + "/" + input.Repo,
		"state":      input.State,
		"count":      len(items),
		"issues":     items,
	})
}

func (e *GitHubExecutor) searchRepositories(
    ctx context.Context,
    arguments json.RawMessage,
) (json.RawMessage, error) {
    var input struct {
        Query   string `json:"query"`
        PerPage int    `json:"per_page"`
    }

    if err := json.Unmarshal(arguments, &input); err != nil {
        return nil, fmt.Errorf("decode search_repositories arguments: %w", err)
    }

    if input.PerPage <= 0 || input.PerPage > 100 {
        input.PerPage = 10
    }

    opts := &github.SearchOptions{
        ListOptions: github.ListOptions{PerPage: input.PerPage},
    }

    result, _, err := e.client.Search.Repositories(ctx, input.Query, opts)
    if err != nil {
        return nil, fmt.Errorf("%w: %v", ErrGitHubUpstream, err)
    }

    items := make([]map[string]any, 0, len(result.Repositories))
    for _, repo := range result.Repositories {
        items = append(items, map[string]any{
            "name":        repo.GetFullName(),
            "description": repo.GetDescription(),
            "stars":       repo.GetStargazersCount(),
            "language":    repo.GetLanguage(),
            "url":         repo.GetHTMLURL(),
        })
    }

    return json.Marshal(map[string]any{
        "query":        input.Query,
        "total_count":  result.GetTotal(),
        "count":        len(items),
        "repositories": items,
    })
}
