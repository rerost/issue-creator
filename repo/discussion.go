package repo

import (
	"context"

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
	return types.Issue{}, nil
}

func (r *discussionRepositoryImpl) FindLastIssueByLabel(ctx context.Context, issue types.Issue) (types.Issue, error) {
	return types.Issue{}, nil
}

func (r *discussionRepositoryImpl) CloseByURL(ctx context.Context, issueURL string) (types.Issue, error) {
	return types.Issue{}, nil
}
