package repo

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/types"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
)

const LastDiscussionNotFound = "Not Found"
const categoryKey = "categoryId"

func NewDisscussionRepository(ghc *githubv4.Client) *discussionRepositoryImpl {
	return &discussionRepositoryImpl{
		ghc: ghc,
	}
}

type discussionRepositoryImpl struct {
	ghc *githubv4.Client
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
		CreatedAt githubv4.Date
		Labels    struct {
			Nodes []struct {
				Id githubv4.String
			}
		} `graphql:"labels(first: 100)"`
	}
)

func (d Discussion) LabelIDs() []string {
	labelIDs := make([]string, 0, len(d.Labels.Nodes))

	for _, label := range d.Labels.Nodes {
		labelIDs = append(labelIDs, string(label.Id))
	}

	return labelIDs
}

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

	// Create Discussion
	var createDisscussionMutation struct {
		CreateDiscussion struct {
			Discussion struct {
				Discussion
			}
		} `graphql:"createDiscussion(input: $input)"`
	}
	{
		input := githubv4.CreateDiscussionInput{
			RepositoryID: q.Repository.Id,
			Title:        githubv4.String(issue.Title),
			Body:         githubv4.String(issue.Body),
			CategoryID:   githubv4.String((*issue.Meta)[categoryKey]),
		}

		err = r.ghc.Mutate(ctx, &createDisscussionMutation, input, nil)
		if err != nil {
			return types.Issue{}, errors.WithStack(err)
		}
		zap.L().Debug("created discussion", zap.String("ID", createDisscussionMutation.CreateDiscussion.Discussion.Id.(string)))
	}

	// Add Labels
	{
		var addLabelMutation struct {
			AddLabelsToLabelable struct {
				ClientMutationId githubv4.String
			} `graphql:"addLabelsToLabelable(input: $input)"`
		}
		labelIDs := []githubv4.ID{}
		for _, labelID := range issue.Labels {
			labelIDs = append(labelIDs, githubv4.ID(labelID))
		}
		input := githubv4.AddLabelsToLabelableInput{
			LabelableID: createDisscussionMutation.CreateDiscussion.Discussion.Id,
			LabelIDs:    labelIDs,
		}

		err = r.ghc.Mutate(ctx, &addLabelMutation, input, nil)
		if err != nil {
			return types.Issue{}, errors.WithStack(err)
		}
		zap.L().Debug("add labels to discussion", zap.Strings("label ids", issue.Labels))
	}

	// 最終的なdiscussionを取得
	var query struct {
		Node struct {
			Discussion `graphql:"... on Discussion"`
		} `graphql:"node(id: $id)"`
	}

	err = r.ghc.Query(
		ctx,
		&query,
		map[string]interface{}{
			"id": createDisscussionMutation.CreateDiscussion.Discussion.Id,
		},
	)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	d := query.Node.Discussion

	meta := map[string]string{
		categoryKey: fmt.Sprintf("%+v", d.Category.Id),
	}
	return types.Issue{
		Owner:      issue.Owner,
		Repository: issue.Repository,
		Title:      string(d.Title),
		Body:       string(d.Body),
		URL:        (*string)(&d.Url),
		Labels:     d.LabelIDs(),
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

	meta := map[string]string{
		categoryKey: fmt.Sprintf("%+v", q.Repository.Discussion.Category.Id),
	}

	return types.Issue{
		Owner:      discussionData.Owner,
		Repository: discussionData.Repository,

		Title:  string(q.Repository.Discussion.Title),
		Body:   string(q.Repository.Discussion.Body),
		URL:    (*string)(&q.Repository.Discussion.Url),
		Labels: q.Repository.Discussion.LabelIDs(),
		Meta:   &meta,
	}, nil
}

func (r *discussionRepositoryImpl) FindLastIssue(ctx context.Context, templateIssue types.Issue) (types.Issue, error) {
	var q struct {
		Search struct {
			Nodes []struct {
				Discussion `graphql:"... on Discussion"`
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

	lastDiscussion := q.Search.Nodes[0].Discussion
	for _, node := range q.Search.Nodes {
		if lastDiscussion.CreatedAt.Time.Before(node.Discussion.CreatedAt.Time) {
			// when finding more recent discussion
			lastDiscussion = node.Discussion
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
		Labels:     lastDiscussion.LabelIDs(),
		Meta:       &meta,
	}, nil
}

func (r *discussionRepositoryImpl) CloseByURL(ctx context.Context, issueURL string) error {
	discussionData, err := parseIssueURL(issueURL)
	if err != nil {
		return errors.WithStack(err)
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
		return errors.WithStack(err)
	}

	var archiveCategoryId githubv4.ID
	for _, category := range q.Repository.DiscussionCategory.Nodes {
		if category.Name == "Archive" {
			archiveCategoryId = category.Id
			break
		}
	}
	if archiveCategoryId == nil {
		return errors.New("Archive Category Not Found")
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
		return errors.WithStack(err)
	}

	return nil
}

func (r *discussionRepositoryImpl) IsValidTemplateIssue(i types.Issue) bool {
	if i.Meta == nil {
		return false
	}
	_, ok := (*i.Meta)[categoryKey]
	return ok
}
