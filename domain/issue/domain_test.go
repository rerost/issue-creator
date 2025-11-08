package issue_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-github/v78/github"
	"github.com/rerost/issue-creator/domain/issue"
	"github.com/rerost/issue-creator/repo"
	"github.com/rerost/issue-creator/types"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// Helper
func ToPtr[T any](v T) *T {
	return &v
}

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

func NewTestRepository(ctx context.Context) repo.Repository {
	githubClient := NewGithubClient(ctx)
	graphQLClient := NewGithubGraphQLClient(ctx)

	return repo.NewRepository(githubClient, graphQLClient)
}

func NewTestIssueService(
	ctx context.Context,
	repo repo.Repository,
	closeLastIssue bool,
	checkBeforeCreateIssue *string,
) issue.IssueService {
	return issue.NewIssueService(
		repo,
		time.Now(),
		closeLastIssue,
		checkBeforeCreateIssue,
	)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewTestRepository(ctx)
	issueRepo := repo.Selector("https://github.com/rerost/issue-creator/issues/1")

	type In struct {
		CheckBeforeCreateIssue *string
	}

	testCase := []struct {
		name string
		in   In
		out  types.Issue
	}{
		{
			name: "CheckBeforeCreateIssue is null",
			in: In{
				CheckBeforeCreateIssue: nil,
			},
			out: types.Issue{
				Owner:      "rerost",
				Repository: "issue-creator-for-test",
				Title:      "Test TestCreate/CheckBeforeCreateIssue_is_null",
				Body:       "Test Issue",
				Labels:     []string{"LA_kwDOJt6V-s8AAAABTiHX9w"},
			},
		},
		{
			name: "CheckBeforeCreateIssue is success",
			in: In{
				CheckBeforeCreateIssue: ToPtr(`test -z ""`),
			},
			out: types.Issue{
				Owner:      "rerost",
				Repository: "issue-creator-for-test",
				Title:      "Test TestCreate/CheckBeforeCreateIssue_is_success",
				Body:       "Test Issue",
				Labels:     []string{"LA_kwDOJt6V-s8AAAABTiHX9w"},
			},
		},
		{
			name: "CheckBeforeCreateIssue is failed",
			in: In{
				CheckBeforeCreateIssue: ToPtr(`test -z "1"`),
			},
			out: types.Issue{}, // Nothing
		},
	}

	for _, test := range testCase {
		test := test
		t.Run(test.name, func(t *testing.T) {
			tempIssue, err := issueRepo.Create(ctx, types.Issue{
				Owner:      "rerost",
				Repository: "issue-creator-for-test",
				Title:      "Test " + t.Name(),
				Body:       "Test Issue",
				Labels:     []string{"LA_kwDOJt6V-s8AAAABTiHX9w"},
			})

			if err != nil {
				t.Error(err)
				return
			}

			issueService := NewTestIssueService(ctx, repo, true, test.in.CheckBeforeCreateIssue)
			res, err := issueService.Create(ctx, *tempIssue.URL)
			if err != nil {
				fmt.Printf("%+v\n", err)
				t.Error(err)
				return
			}

			if diff := cmp.Diff(res, test.out, cmpopts.IgnoreFields(res, "Body", "URL", "LastIssueURL")); diff != "" {
				t.Error(diff)
				return
			}
		})
	}
}
