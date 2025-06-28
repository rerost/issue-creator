package repo

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/go-github/v73/github"
	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/types"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
)

type IssueRepository interface {
	Create(ctx context.Context, issue types.Issue) (types.Issue, error)
	FindByURL(ctx context.Context, issueURL string) (types.Issue, error)
	FindLastIssue(ctx context.Context, templateIssue types.Issue) (types.Issue, error)
	CloseByURL(ctx context.Context, issueURL string) error
	IsValidTemplateIssue(types.Issue) bool
}

type Repository struct {
	discussionRepo IssueRepository
	issueRepo      IssueRepository
}

func (r Repository) Selector(url string) IssueRepository {
	zap.L().Debug("Selector", zap.Bool("isDiscussion", isDiscussion(url)))
	if isDiscussion(url) {
		return r.discussionRepo
	}
	return r.issueRepo
}

func NewRepository(githubClient *github.Client, githubGraphqlClient *githubv4.Client) Repository {
	return Repository{
		discussionRepo: NewDisscussionRepository(githubGraphqlClient),
		issueRepo: &issueRepositoryImpl{
			ghc: githubClient,
		},
	}
}

type issueRepositoryImpl struct {
	ghc *github.Client
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
	issueData, err := ParseIssueURL(issueURL)
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

func (ir *issueRepositoryImpl) FindLastIssue(ctx context.Context, templateIssue types.Issue) (types.Issue, error) {
	labelsQueries := []string{}
	for _, l := range templateIssue.Labels {
		labelsQueries = append(labelsQueries, fmt.Sprintf(`label:"%s"`, l))
	}

	queries := append(
		labelsQueries,
		fmt.Sprintf("repo:%s/%s", templateIssue.Owner, templateIssue.Repository),
		"sort:created-desc",
		"is:issue",
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
		Owner:      templateIssue.Owner,
		Repository: templateIssue.Repository,

		Title:  r.Issues[0].GetTitle(),
		Body:   r.Issues[0].GetBody(),
		Labels: labels,
		URL:    r.Issues[0].HTMLURL,
	}, nil
}

var (
	// State
	closed = "closed"
)

func (ir *issueRepositoryImpl) CloseByURL(ctx context.Context, issueURL string) error {
	issueData, err := ParseIssueURL(issueURL)
	if err != nil {
		return errors.WithStack(err)
	}

	_, _, err = ir.ghc.Issues.Edit(ctx, issueData.Owner, issueData.Repository, issueData.IssueNumber, &github.IssueRequest{State: &closed})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

type issueURLData struct {
	Owner       string
	Repository  string
	IssueNumber int
}

func ParseIssueURL(u string) (issueURLData, error) {
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

func (r *issueRepositoryImpl) IsValidTemplateIssue(i types.Issue) bool {
	return len(i.Labels) != 0
}

func isDiscussion(templateIssueURL string) bool {
	pu, err := url.Parse(templateIssueURL)
	if err != nil {
		zap.L().Debug("error", zap.String("url parse err", err.Error()))
		return false
	}

	path := pu.Path
	s := strings.Split(path, "/")
	zap.L().Debug("", zap.String("path", path))
	// Expect: /:owner/:repository/[discussions|issues]/:number
	if len(s) != 5 {
		zap.L().Debug("error", zap.Int("path length", len(s)))
		return false
	}
	return s[3] == "discussions"
}
