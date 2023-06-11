package repo_test

import (
	"context"
	_ "embed"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/rerost/issue-creator/repo"
	"github.com/rerost/issue-creator/types"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func ToPtr[T any](v T) *T {
	return &v
}

func NewTestDiscussionRepository(ctx context.Context) repo.IssueRepository {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("TEST_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	return repo.NewDisscussionRepository(githubv4.NewClient(tc))

}

func TestCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCase := []struct {
		in  string
		out types.Issue
	}{
		{
			in: "https://github.com/rerost/issue-creator-for-test/discussions/5",
			out: types.Issue{
				Owner:      "rerost",
				Repository: "issue-creator-for-test",
				Title:      "[TEST] TestCreate",
				Body:       "## Test\r\n## Test",
				Labels:     []string{"LA_kwDOJt6V-s8AAAABTiHX9w"},
				Meta:       &map[string]string{"categoryId": "DIC_kwDOJt6V-s4CXH0p"},
			},
		},
	}

	for _, test := range testCase {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			discussionRepo := NewTestDiscussionRepository(ctx)
			issue, err := discussionRepo.FindByURL(ctx, test.in)
			if err != nil {
				t.Error(err)
				return
			}
			res, err := discussionRepo.Create(ctx, issue)
			if err != nil {
				t.Error(err)
				return
			}

			defer discussionRepo.CloseByURL(ctx, *res.URL)

			if diff := cmp.Diff(res, test.out, cmpopts.IgnoreFields(res, "URL")); diff != "" {
				t.Error(diff)
				return
			}
		})
	}
}

func TestFindByURL(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCase := []struct {
		in  string
		out types.Issue
	}{
		{
			in: "https://github.com/rerost/issue-creator-for-test/discussions/1",
			out: types.Issue{
				Owner:      "rerost",
				Repository: "issue-creator-for-test",
				Title:      "[TEST] TestFindByURL",
				Body:       "## test\r\n\r\n## test",
				Labels:     []string{},
				URL:        ToPtr("https://github.com/rerost/issue-creator-for-test/discussions/1"),
				Meta:       &map[string]string{"categoryId": "DIC_kwDOJt6V-s4CXH0p"},
			},
		},
	}

	for _, test := range testCase {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			discussionRepo := NewTestDiscussionRepository(ctx)
			res, err := discussionRepo.FindByURL(ctx, test.in)
			if err != nil {
				t.Error(err)
				return
			}

			if diff := cmp.Diff(res, test.out); diff != "" {
				t.Error(diff)
				return
			}
		})
	}
}

func TestFindLastIssue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// MEMO: インターフェースとしてはissueの中身までチェックするのが正しいが、挙動としては1つ前のissueということだけが重要なのでそこに絞ってテストしている
	testCase := []struct {
		in  string
		out string
	}{
		{
			in:  "https://github.com/rerost/issue-creator-for-test/discussions/2",
			out: "https://github.com/rerost/issue-creator-for-test/discussions/2",
		},
	}

	for _, test := range testCase {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			discussionRepo := NewTestDiscussionRepository(ctx)
			issue, err := discussionRepo.FindByURL(ctx, test.in)
			if err != nil {
				t.Error(err)
				return
			}
			res, err := discussionRepo.FindLastIssue(ctx, issue)
			if err != nil {
				t.Error(err)
				return
			}

			if diff := cmp.Diff(res.URL, &test.out); diff != "" {
				t.Error(diff)
				return
			}
		})
	}
}

func ArchiveToGeneral(t *testing.T, ctx context.Context, url string) {
	t.Helper()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("TEST_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient := githubv4.NewClient(tc)

	discussionData, err := repo.ParseIssueURL(url)
	if err != nil {
		t.Error(err)
		return
	}

	var q struct {
		Repository struct {
			Id                 githubv4.String
			DiscussionCategory struct {
				Nodes []struct {
					Id   githubv4.ID
					Name githubv4.String
				}
			} `graphql:"discussionCategories(first: 100)"`
			Discussion struct {
				Id githubv4.ID
			} `graphql:"discussion(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"number": githubv4.Int(discussionData.IssueNumber),
		"owner":  githubv4.String(discussionData.Owner),
		"name":   githubv4.String(discussionData.Repository),
	}

	err = githubClient.Query(ctx, &q, variables)
	if err != nil {
		t.Error(err)
		return
	}

	var m struct {
		UpdateDiscussion struct {
			Discussion struct {
				Id githubv4.ID
			}
		} `graphql:"updateDiscussion(input: $input)"`
	}

	input := githubv4.UpdateDiscussionInput{
		DiscussionID: q.Repository.Discussion.Id,
		CategoryID:   githubv4.NewID("DIC_kwDOJt6V-s4CXH0p"), /* General Category */
	}

	err = githubClient.Mutate(ctx, &m, input, nil)
	if err != nil {
		t.Error(err)
		return
	}
}

// WARNING: https://github.com/rerost/issue-creator-for-test の状態が変わるので、並列でこのテストが走ると問題になる
func TestCloseByURL(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// MEMO: インターフェースとしてはissueの中身までチェックするのが正しいが、挙動としては1つ前のissueということだけが重要なのでそこに絞ってテストしている
	testCase := []struct {
		in  string
		out string
	}{
		{
			in: "https://github.com/rerost/issue-creator-for-test/discussions/4",
		},
	}

	for _, test := range testCase {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			discussionRepo := NewTestDiscussionRepository(ctx)
			oldState, err := discussionRepo.FindByURL(ctx, test.in)
			if err != nil {
				t.Error(err)
				return
			}
			// validate
			if (*oldState.Meta)["categoryId"] != "DIC_kwDOJt6V-s4CXH0p" /* General Category */ {
				t.Errorf("%v is not valid state", test.in)
				return
			}

			err = discussionRepo.CloseByURL(ctx, test.in)
			if err != nil {
				t.Error(err)
				return
			}
			defer ArchiveToGeneral(t, ctx, test.in)

			newState, err := discussionRepo.FindByURL(ctx, test.in)
			if err != nil {
				t.Error(err)
				return
			}

			if (*oldState.Meta)["categoryId"] != (*newState.Meta)["categoryId"] && (*newState.Meta)["categoryId"] != "DIC_kwDOJt6V-s4CXH0u" /* Archive Category */ {
				t.Errorf("%v is not close(archived)", test.in)
			}
		})
	}
}

func TestIsValidTemplateIssue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testCase := []struct {
		in  types.Issue
		out bool
	}{
		{
			in: types.Issue{
				Meta: nil,
			},
			out: false,
		},
		{
			in: types.Issue{
				Meta: &map[string]string{"categoryId": "test-id"},
			},
			out: true,
		},
	}
	for _, test := range testCase {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			discussionRepo := NewTestDiscussionRepository(ctx)
			res := discussionRepo.IsValidTemplateIssue(test.in)

			if diff := cmp.Diff(res, test.out); diff != "" {
				t.Error(diff)
				return
			}
		})
	}
}
