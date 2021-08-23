package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/types"
	"github.com/shurcooL/githubv4"
)

type DiscussionRepository interface {
	Create(ctx context.Context, issue types.Issue) (types.Issue, error)
	FindByURL(ctx context.Context, issueURL string) (types.Issue, error)
	FindLastIssueByLabel(ctx context.Context, issue types.Issue) (types.Issue, error)
	CloseByURL(ctx context.Context, issueURL string) (types.Issue, error)
}

type discussionRepositoryImpl struct {
	ghc *githubv4.Client
}

func NewDiscussionRepository(githubClient *githubv4.Client) DiscussionRepository {
	return &discussionRepositoryImpl{
		ghc: githubClient,
	}
}

type (
	Discussion struct {
		Body     githubv4.String
		Url      githubv4.String
		Title    githubv4.String
		Category struct {
			Name githubv4.String
		}
		Labels struct {
			Nodes []struct {
				Name githubv4.String
			}
		} `graphql:"labels(first: 10)"`
		CreatedAt githubv4.Date
	}
)

func (r *discussionRepositoryImpl) Create(ctx context.Context, issue types.Issue) (types.Issue, error) {

	return types.Issue{}, nil
}

func (r *discussionRepositoryImpl) FindByURL(ctx context.Context, issueURL string) (types.Issue, error) {
	discussionData, err := parseIssueURL(issueURL)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	var q struct {
		Repository struct {
			Id         githubv4.String
			Discussion `graphql:"discussion(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"number": githubv4.Int(discussionData.IssueNumber),
		"owner":  githubv4.String(discussionData.Owner),
		"name":   githubv4.String(discussionData.Repository),
	}

	err = r.ghc.Query(ctx, &q, variables)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	ls := make([]string, 0, len(q.Repository.Discussion.Labels.Nodes))
	for _, label := range q.Repository.Discussion.Labels.Nodes {
		ls = append(ls, string(label.Name))
	}

	return types.Issue{
		Owner:      discussionData.Owner,
		Repository: discussionData.Repository,

		Title:  string(q.Repository.Discussion.Title),
		Body:   string(q.Repository.Discussion.Body),
		Labels: ls,
		URL:    (*string)(&q.Repository.Discussion.Url),
	}, nil
}

func (r *discussionRepositoryImpl) FindLastIssueByLabel(ctx context.Context, issue types.Issue) (types.Issue, error) {
	var q struct {
		Search struct {
			Nodes []struct {
				Discussion `graphql:"... on Discussion"`
			}
		} `graphql:"search(query: $query, type: $type, first: 10)"`
	}
	queries := make([]string, 0, len(issue.Labels)+1)
	queries = append(
		queries,
		fmt.Sprintf("repo:%s/%s", issue.Owner, issue.Repository),
	)
	for _, label := range issue.Labels {
		queries = append(queries, fmt.Sprintf(`label:%s`, label))
	}
	githubSearchQuery := strings.Join(queries, " ")
	variables := map[string]interface{}{
		"query": githubv4.String(githubSearchQuery),
		"type":  githubv4.SearchTypeDiscussion,
	}
	err := r.ghc.Query(ctx, &q, variables)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	if len(q.Search.Nodes) == 0 {
		return issue, errors.New("Not found last issue")
	}
	lastDiscussion := q.Search.Nodes[0].Discussion
	for _, node := range q.Search.Nodes {
		if lastDiscussion.CreatedAt.Time.Before(node.Discussion.CreatedAt.Time) {
			// when finding more recent discussion
			lastDiscussion = node.Discussion
		}
	}
	ls := make([]string, 0, len(lastDiscussion.Labels.Nodes))
	for _, label := range lastDiscussion.Labels.Nodes {
		ls = append(ls, string(label.Name))
	}
	url := string(lastDiscussion.Url)
	return types.Issue{
		Owner:      issue.Owner,
		Repository: issue.Repository,
		Title:      string(lastDiscussion.Title),
		Body:       string(lastDiscussion.Body),
		Labels:     ls,
		URL:        &url,
	}, nil
}

func (r *discussionRepositoryImpl) CloseByURL(ctx context.Context, issueURL string) (types.Issue, error) {
	return types.Issue{}, nil
}
