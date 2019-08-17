package repo

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/types"
	"go.uber.org/zap"
)

type IssueRepository interface {
	Create(ctx context.Context, issue types.Issue) (types.Issue, error)
	FindByURL(ctx context.Context, issueURL string) (types.Issue, error)
	FindLastIssueByLabel(ctx context.Context, issue types.Issue) (types.Issue, error)
}

type issueRepositoryImpl struct {
	ghc *github.Client
}

func NewIssueRepository(githubClient *github.Client) IssueRepository {
	return &issueRepositoryImpl{
		ghc: githubClient,
	}
}

func (ir *issueRepositoryImpl) Create(ctx context.Context, issue types.Issue) (types.Issue, error) {
	gi := github.IssueRequest{
		Title:  &issue.Title,
		Body:   &issue.Body,
		Labels: &issue.Labels,
	}
	zap.L().Debug("create issue", zap.String("owner", issue.Owner))
	zap.L().Debug("create issue", zap.String("repository", issue.Repository))
	i, _, err := ir.ghc.Issues.Create(ctx, issue.Owner, issue.Repository, &gi)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}
	return types.Issue{
		Owner:      issue.Owner,
		Repository: issue.Repository,

		Title:        i.GetTitle(),
		Body:         i.GetBody(),
		Labels:       types.FromGithubLabels(i.Labels),
		URL:          i.HTMLURL,
		LastIssueURL: issue.LastIssueURL,
	}, nil
}

func (ir *issueRepositoryImpl) FindByURL(ctx context.Context, issueURL string) (types.Issue, error) {
	issueData, err := parseIssueURL(issueURL)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	i, _, err := ir.ghc.Issues.Get(ctx, issueData.Owner, issueData.Repository, issueData.IssueNumber)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	labels := types.FromGithubLabels(i.Labels)
	htmlurl := i.GetHTMLURL()

	return types.Issue{
		Owner:      issueData.Owner,
		Repository: issueData.Repository,

		Title:  i.GetTitle(),
		Body:   i.GetBody(),
		Labels: labels,
		URL:    &htmlurl,
	}, nil
}

func (ir *issueRepositoryImpl) FindLastIssueByLabel(ctx context.Context, issue types.Issue) (types.Issue, error) {
	labelsQueries := []string{}
	for _, l := range issue.Labels {
		labelsQueries = append(labelsQueries, fmt.Sprintf(`label:"%s"`, l))
	}

	queries := append(
		labelsQueries,
		fmt.Sprintf("repo:%s/%s", issue.Owner, issue.Repository),
		"sort:created-desc",
	)
	githubSearchQuery := strings.Join(queries, " ")
	zap.L().Debug("query", zap.String("github_search_query", githubSearchQuery))
	r, _, err := ir.ghc.Search.Issues(ctx, githubSearchQuery, nil)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	if r.GetTotal() == 0 || len(r.Issues) == 0 {
		return types.Issue{}, errors.New("Not found last issue")
	}

	labels := types.FromGithubLabels(r.Issues[0].Labels)
	return types.Issue{
		Owner:      issue.Owner,
		Repository: issue.Repository,

		Title:  r.Issues[0].GetTitle(),
		Body:   r.Issues[0].GetBody(),
		Labels: labels,
		URL:    r.Issues[0].HTMLURL,
	}, nil
}

type issueURLData struct {
	Owner       string
	Repository  string
	IssueNumber int
}

func parseIssueURL(u string) (issueURLData, error) {
	pu, err := url.Parse(u)
	if err != nil {
		return issueURLData{}, errors.WithStack(err)
	}

	path := pu.Path
	s := strings.Split(path, "/")
	zap.L().Debug("", zap.String("path", path))
	// Expect: /:owner/:repository/issues/:issue_number
	if len(s) != 5 {
		zap.L().Debug("error", zap.Int("path length", len(s)))
		return issueURLData{}, errors.New("Failed to parse url")
	}

	issueNumber, err := strconv.Atoi(s[4])
	if err != nil {
		return issueURLData{}, errors.WithStack(err)
	}

	return issueURLData{
		Owner:       s[1],
		Repository:  s[2],
		IssueNumber: issueNumber,
	}, nil
}
