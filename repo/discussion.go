package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/types"
	"github.com/shurcooL/githubv4"
)

const LastDiscussionNotFound = "Not Found"
const categoryKey = "categoryId"

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
		Id       githubv4.ID
		Body     githubv4.String
		Url      githubv4.String
		Title    githubv4.String
		Category struct {
			Id   githubv4.ID
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
	if issue.Meta == nil {
		return types.Issue{}, errors.New("Category not found")
	}
	if _, ok := (*issue.Meta)[categoryKey]; !ok {
		return types.Issue{}, errors.New("Category not found")
	}
	var q struct {
		Repository struct {
			Id githubv4.ID
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	variables := map[string]interface{}{
		"owner": githubv4.String(issue.Owner),
		"name":  githubv4.String(issue.Repository),
	}

	err := r.ghc.Query(ctx, &q, variables)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	var m struct {
		CreateDiscussion struct {
			Discussion struct {
				Discussion
			}
		} `graphql:"createDiscussion(input: $input)"`
	}
	input := githubv4.CreateDiscussionInput{
		RepositoryID: q.Repository.Id,
		Title:        githubv4.String(issue.Title),
		Body:         githubv4.String(issue.Body),
		CategoryID:   githubv4.String((*issue.Meta)[categoryKey]),
	}

	err = r.ghc.Mutate(ctx, &m, input, nil)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	d := m.CreateDiscussion.Discussion.Discussion
	ls := make([]string, 0, len(d.Labels.Nodes))
	for _, label := range d.Labels.Nodes {
		ls = append(ls, string(label.Name))
	}
	meta := map[string]string{
		categoryKey: fmt.Sprintf("%+v", d.Category.Id),
	}
	return types.Issue{
		Owner:      issue.Owner,
		Repository: issue.Repository,
		Title:      string(d.Title),
		Body:       string(d.Body),
		Labels:     ls,
		URL:        (*string)(&d.Url),
		Meta:       &meta,
	}, nil
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

	meta := map[string]string{
		categoryKey: fmt.Sprintf("%+v", q.Repository.Discussion.Category.Id),
	}

	return types.Issue{
		Owner:      discussionData.Owner,
		Repository: discussionData.Repository,

		Title:  string(q.Repository.Discussion.Title),
		Body:   string(q.Repository.Discussion.Body),
		Labels: ls,
		URL:    (*string)(&q.Repository.Discussion.Url),
		Meta:   &meta,
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
		return issue, errors.New(LastDiscussionNotFound)
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
	meta := map[string]string{
		categoryKey: fmt.Sprintf("%+v", lastDiscussion.Category.Id),
	}
	return types.Issue{
		Owner:      issue.Owner,
		Repository: issue.Repository,
		Title:      string(lastDiscussion.Title),
		Body:       string(lastDiscussion.Body),
		Labels:     ls,
		URL:        (*string)(&lastDiscussion.Url),
		Meta:       &meta,
	}, nil
}

func (r *discussionRepositoryImpl) CloseByURL(ctx context.Context, issueURL string) (types.Issue, error) {
	discussionData, err := parseIssueURL(issueURL)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
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

	var archiveCategoryId githubv4.ID
	for _, category := range q.Repository.DiscussionCategory.Nodes {
		if category.Name == "Archive" {
			archiveCategoryId = category.Id
			break
		}
	}
	if archiveCategoryId == nil {
		return types.Issue{}, errors.New("Archive Category Not Found")
	}

	var m struct {
		UpdateDiscussion struct {
			Discussion struct {
				Discussion
			}
		} `graphql:"updateDiscussion(input: $input)"`
	}

	input := githubv4.UpdateDiscussionInput{
		DiscussionID: q.Repository.Discussion.Id,
		Title:        &q.Repository.Discussion.Title,
		Body:         &q.Repository.Discussion.Body,
		CategoryID:   &archiveCategoryId,
	}

	err = r.ghc.Mutate(ctx, &m, input, nil)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	d := m.UpdateDiscussion.Discussion.Discussion
	ls := make([]string, 0, len(d.Labels.Nodes))
	for _, label := range d.Labels.Nodes {
		ls = append(ls, string(label.Name))
	}
	meta := map[string]string{
		categoryKey: fmt.Sprintf("%+v", d.Category.Id),
	}
	return types.Issue{
		Owner:      discussionData.Owner,
		Repository: discussionData.Repository,
		Title:      string(d.Title),
		Body:       string(d.Body),
		Labels:     ls,
		URL:        (*string)(&d.Url),
		Meta:       &meta,
	}, nil
}
