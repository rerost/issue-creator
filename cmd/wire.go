//go:build wireinject
// +build wireinject

package cmd

import (
	"context"
	"time"

	"github.com/google/go-github/v66/github"
	"github.com/google/wire"
	"github.com/rerost/issue-creator/domain/issue"
	"github.com/rerost/issue-creator/domain/schedule"
	"github.com/rerost/issue-creator/repo"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func CurrentTime(cfg Config) time.Time {
	return time.Now()
}

func NewGithubClient(ctx context.Context, cfg Config) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GithubAccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	c := github.NewClient(tc)
	return c
}

func NewGithubGraphqlClient(ctx context.Context, cfg Config) *githubv4.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GithubAccessToken},
	)
	httpClient := oauth2.NewClient(ctx, src)

	c := githubv4.NewClient(httpClient)
	return c
}

func NewK8sCommand(cfg Config) []string {
	return cfg.K8sCommands
}

func NewTemplateFile(cfg Config) string {
	return cfg.ManifestTemplateFile
}

func NewIssueService(cfg Config, issueRepo repo.Repository, ct time.Time) issue.IssueService {
	return issue.NewIssueService(
		issueRepo,
		ct,
		cfg.CloseLastIssue,
		cfg.CheckBeforeCreateIssue,
	)
}

func NewScheduleService(cfg Config, scheduleRepository repo.ScheduleRepository) schedule.ScheduleService {
	return schedule.NewScheduleService(scheduleRepository, cfg.CloseLastIssue, cfg.CheckBeforeCreateIssue)
}

func InitializeCmd(ctx context.Context, cfg Config) (*cobra.Command, error) {
	wire.Build(
		NewCmdRoot,
		NewIssueService,
		repo.NewRepository,
		CurrentTime,
		NewGithubClient,
		NewGithubGraphqlClient,
		NewScheduleService,
		repo.NewScheduleRepository,
		NewK8sCommand,
		NewTemplateFile,
	)
	return nil, nil
}
