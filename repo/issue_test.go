package repo_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v75/github"
	"github.com/rerost/issue-creator/repo"
	"github.com/rerost/issue-creator/types"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func NewGithubGraphQLClient(ctx context.Context) *githubv4.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("TEST_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	return githubv4.NewClient(tc)
}

func NewGithubClient(ctx context.Context) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("TEST_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	c := github.NewClient(tc)
	return c
}

func NewTestIssueRepository(ctx context.Context) repo.IssueRepository {
	githubClient := NewGithubClient(ctx)
	graphQLClient := NewGithubGraphQLClient(ctx)

	repo := repo.NewRepository(githubClient, graphQLClient)

	return repo.Selector("https://github.com/rerost/issue-creator/issues/1") // GitHub Issueを管理するIssueRepositoryを返す
}

func TestIssueFindByURL(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := NewTestIssueRepository(ctx)

	url := "https://github.com/rerost/issue-creator-for-test/issues/336"

	out := types.Issue{
		Owner:      "rerost",
		Repository: "issue-creator-for-test",
		Title:      "Test for CI 1",
		Body:       "test",
		Labels:     []string{"test-for-ci-1"},
		URL:        ToPtr("https://github.com/rerost/issue-creator-for-test/issues/336"),
	}

	res, err := repo.FindByURL(ctx, url)
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(res, out); diff != "" {
		t.Error(diff)
	}
}

func TestIssueFindLastIssue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := NewTestIssueRepository(ctx)

	url := "https://github.com/rerost/issue-creator-for-test/issues/336"

	issue, err := repo.FindByURL(ctx, url)
	if err != nil {
		t.Error(err)
	}

	res, err := repo.FindLastIssue(ctx, issue)
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(res, issue); diff != "" {
		t.Error(diff)
	}
}

func TestIssueIsValidTemplateIssue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := NewTestIssueRepository(ctx)

	url := "https://github.com/rerost/issue-creator-for-test/issues/336"

	issue, err := repo.FindByURL(ctx, url)
	if err != nil {
		t.Error(err)
	}

	res := repo.IsValidTemplateIssue(issue)

	if diff := cmp.Diff(res, true); diff != "" {
		t.Error(diff)
	}
}
