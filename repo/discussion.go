package repo

import (
	"context"

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
			Discussion struct {
				Id       githubv4.String
				Body     githubv4.String
				Url      githubv4.String
				Title    githubv4.String
				Category struct {
					Id githubv4.String
				}
				Labels struct {
					Nodes []struct {
						Name githubv4.String
					}
				} `graphql:"labels(first: 10)"`
			} `graphql:"discussion(number: $number)"`
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
	return types.Issue{}, nil
}

func (r *discussionRepositoryImpl) CloseByURL(ctx context.Context, issueURL string) (types.Issue, error) {
	return types.Issue{}, nil
}
