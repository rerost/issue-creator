package repo

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/types"
	"github.com/shurcooL/githubv4"
)

const LastDiscussionNotFound = "Not Found"
const categoryKey = "categoryId"

type discussionRepositoryImpl struct {
	ghc *githubv4.Client
}

type (
	discussion struct {
		Id       githubv4.ID
		Body     githubv4.String
		Url      githubv4.String
		Title    githubv4.String
		Category struct {
			Id   githubv4.ID
			Name githubv4.String
		}
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
				discussion
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

	d := m.CreateDiscussion.Discussion.discussion
	meta := map[string]string{
		categoryKey: fmt.Sprintf("%+v", d.Category.Id),
	}
	return types.Issue{
		Owner:      issue.Owner,
		Repository: issue.Repository,
		Title:      string(d.Title),
		Body:       string(d.Body),
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
			discussion `graphql:"discussion(number: $number)"`
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

	meta := map[string]string{
		categoryKey: fmt.Sprintf("%+v", q.Repository.discussion.Category.Id),
	}

	return types.Issue{
		Owner:      discussionData.Owner,
		Repository: discussionData.Repository,

		Title: string(q.Repository.discussion.Title),
		Body:  string(q.Repository.discussion.Body),
		URL:   (*string)(&q.Repository.discussion.Url),
		Meta:  &meta,
	}, nil
}

func (r *discussionRepositoryImpl) FindLastIssue(ctx context.Context, templateIssue types.Issue) (types.Issue, error) {
	var q struct {
		Search struct {
			Nodes []struct {
				discussion `graphql:"... on Discussion"`
			}
		} `graphql:"search(query: $query, type: $type, first: 100)"`
	}
	p, _ := regexp.Compile("({{.*?}})")
	title := p.ReplaceAllString(templateIssue.Title, "")
	searchQueries := make([]string, 0)
	searchQueries = append(searchQueries, fmt.Sprintf("%s in:title", title))
	searchQueries = append(
		searchQueries,
		fmt.Sprintf("repo:%s/%s", templateIssue.Owner, templateIssue.Repository),
	)
	githubSearchQuery := strings.Join(searchQueries, " ")
	variables := map[string]interface{}{
		"query": githubv4.String(githubSearchQuery),
		"type":  githubv4.SearchTypeDiscussion,
	}
	err := r.ghc.Query(ctx, &q, variables)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}
	if len(q.Search.Nodes) == 0 {
		return templateIssue, errors.New(LastDiscussionNotFound)
	}

	lastDiscussion := q.Search.Nodes[0].discussion
	for _, node := range q.Search.Nodes {
		if lastDiscussion.CreatedAt.Time.Before(node.discussion.CreatedAt.Time) {
			// when finding more recent discussion
			lastDiscussion = node.discussion
		}
	}

	meta := map[string]string{
		categoryKey: fmt.Sprintf("%+v", lastDiscussion.Category.Id),
	}
	return types.Issue{
		Owner:      templateIssue.Owner,
		Repository: templateIssue.Repository,
		Title:      string(lastDiscussion.Title),
		Body:       string(lastDiscussion.Body),
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
			discussion `graphql:"discussion(number: $number)"`
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
				discussion
			}
		} `graphql:"updateDiscussion(input: $input)"`
	}

	input := githubv4.UpdateDiscussionInput{
		DiscussionID: q.Repository.discussion.Id,
		Title:        &q.Repository.discussion.Title,
		Body:         &q.Repository.discussion.Body,
		CategoryID:   &archiveCategoryId,
	}

	err = r.ghc.Mutate(ctx, &m, input, nil)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	d := m.UpdateDiscussion.Discussion.discussion
	meta := map[string]string{
		categoryKey: fmt.Sprintf("%+v", d.Category.Id),
	}
	return types.Issue{
		Owner:      discussionData.Owner,
		Repository: discussionData.Repository,
		Title:      string(d.Title),
		Body:       string(d.Body),
		URL:        (*string)(&d.Url),
		Meta:       &meta,
	}, nil
}

func (r *discussionRepositoryImpl) IsValidTemplateIssue(i types.Issue) bool {
	if i.Meta == nil {
		return false
	}
	_, ok := (*i.Meta)[categoryKey]
	return ok
}
