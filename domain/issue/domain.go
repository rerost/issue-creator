package issue

import (
	"bytes"
	"context"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/rerost/issue-scheduler/repo"
	"github.com/rerost/issue-scheduler/types"
)

type TemplateData struct {
	CurrentTime time.Time
	LastIssue   types.Issue
}

type IssueService interface {
	Create(ctx context.Context, templateURL string) (types.Issue, error)
}

type issueServiceImpl struct {
	ir repo.IssueRepository
	ct time.Time
}

func NewIssueService(issueRepository repo.IssueRepository, currentTime time.Time) IssueService {
	return &issueServiceImpl{
		ir: issueRepository,
		ct: currentTime,
	}
}

// Render return not saved issue
func (is *issueServiceImpl) render(ctx context.Context, templateIssueURL string) (types.Issue, error) {
	_templateIssue, err := is.ir.FindByURL(ctx, templateIssueURL)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	titleTmpl, err := template.ParseGlob(_templateIssue.Title)
	if err != nil {
		return types.Issue{}, errors.Wrap(err, "Failed to parse title")
	}
	bodyTmpl, err := template.ParseGlob(_templateIssue.Body)
	if err != nil {
		return types.Issue{}, errors.Wrap(err, "Failed to parse body")
	}

	if len(_templateIssue.Labels) == 0 {
		return types.Issue{}, errors.Wrap(err, "Requires at least one label")
	}

	lastIssue, err := is.ir.FindLastIssueByLabel(ctx, _templateIssue)
	if err != nil {
		return types.Issue{}, errors.Wrap(err, "Failed to get last issue")
	}

	tw := bytes.NewBufferString("")
	err = titleTmpl.Execute(tw, TemplateData{CurrentTime: is.ct, LastIssue: lastIssue})
	if err != nil {
		return types.Issue{}, errors.Wrap(err, "Failed to render title")
	}
	title := string(tw.Bytes())

	bw := bytes.NewBufferString("")
	err = bodyTmpl.Execute(bw, TemplateData{CurrentTime: is.ct, LastIssue: lastIssue})
	if err != nil {
		return types.Issue{}, errors.Wrap(err, "Failed to render body")
	}
	body := string(bw.Bytes())

	if lastIssue.URL == nil {
		return types.Issue{}, errors.New("Invalid last issue passed(empty URL)")
	}

	return types.Issue{
		Title:        title,
		Body:         body,
		Labels:       _templateIssue.Labels,
		LastIssueURL: *lastIssue.URL,
	}, nil
}

func (is *issueServiceImpl) Create(ctx context.Context, templateURL string) (types.Issue, error) {
	i, err := is.render(ctx, templateURL)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	i, err = is.ir.Create(ctx, i)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	return i, nil
}
