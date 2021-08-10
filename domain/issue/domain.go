package issue

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/repo"
	"github.com/rerost/issue-creator/types"
	"go.uber.org/zap"
)

type TemplateData struct {
	CurrentTime time.Time
	LastIssue   types.Issue
	AddDay      func(int) time.Time
}

type IssueService interface {
	Create(ctx context.Context, templateURL string) (types.Issue, error)
	Render(ctx context.Context, templateURL string) (types.Issue, error)
}

type issueServiceImpl struct {
	ir                     repo.IssueRepository
	dr                     repo.DiscussionRepository
	ct                     time.Time
	closeLastIssue         bool
	checkBeforeCreateIssue *string
	isDiscussion           bool
}

func NewIssueService(issueRepository repo.IssueRepository, discussionRepository repo.DiscussionRepository, currentTime time.Time, closeLastIssue bool, checkBeforeCreateIssue *string, isDiscussion bool) IssueService {
	return &issueServiceImpl{
		ir:                     issueRepository,
		dr:                     discussionRepository,
		ct:                     currentTime,
		closeLastIssue:         closeLastIssue,
		checkBeforeCreateIssue: checkBeforeCreateIssue,
		isDiscussion:           isDiscussion,
	}
}

// Render return not saved issue
func (is *issueServiceImpl) render(ctx context.Context, templateIssueURL string) (types.Issue, error) {
	zap.L().Debug("templateIssueURL", zap.String("templateIssueURL", templateIssueURL))
	var _templateIssue types.Issue
	if is.isDiscussion {
		ti, err := is.dr.FindByURL(ctx, templateIssueURL)
		if err != nil {
			return types.Issue{}, errors.WithStack(err)
		}
		_templateIssue = ti
	} else {
		ti, err := is.ir.FindByURL(ctx, templateIssueURL)
		if err != nil {
			return types.Issue{}, errors.WithStack(err)
		}
		_templateIssue = ti
	}

	zap.L().Debug("template", zap.String("Title", _templateIssue.Title))
	zap.L().Debug("template", zap.String("Body", _templateIssue.Body))
	titleTmpl, err := template.New("title").Funcs(map[string]interface{}{
		"AddDateAndFormat": func(format string, d int) string { return is.ct.AddDate(0, 0, d).Format(format) },
	}).Parse(_templateIssue.Title)
	if err != nil {
		return types.Issue{}, errors.Wrap(err, "Failed to parse title")
	}
	bodyTmpl, err := template.New("body").Funcs(map[string]interface{}{
		"AddDateAndFormat": func(format string, d int) string { return is.ct.AddDate(0, 0, d).Format(format) },
	}).Parse(_templateIssue.Body)
	if err != nil {
		return types.Issue{}, errors.Wrap(err, "Failed to parse body")
	}

	if len(_templateIssue.Labels) == 0 {
		return types.Issue{}, errors.New("Requires at least one label")
	}

	lastIssue, err := is.ir.FindLastIssueByLabel(ctx, _templateIssue)
	if err != nil {
		return types.Issue{}, errors.Wrap(err, "Failed to get last issue")
	}

	tw := bytes.NewBufferString("")
	err = titleTmpl.Execute(tw, TemplateData{CurrentTime: is.ct, LastIssue: lastIssue, AddDay: func(d int) time.Time { return is.ct.AddDate(0, 0, d) }})
	if err != nil {
		return types.Issue{}, errors.Wrap(err, "Failed to render title")
	}
	title := string(tw.Bytes())

	bw := bytes.NewBufferString("")
	err = bodyTmpl.Execute(bw, TemplateData{CurrentTime: is.ct, LastIssue: lastIssue, AddDay: func(d int) time.Time { return is.ct.AddDate(0, 0, d) }})
	if err != nil {
		return types.Issue{}, errors.Wrap(err, "Failed to render body")
	}
	body := string(bw.Bytes())

	if lastIssue.URL == nil {
		return types.Issue{}, errors.New("Invalid last issue passed(empty URL)")
	}

	res := types.Issue{
		Owner:        _templateIssue.Owner,
		Repository:   _templateIssue.Repository,
		Title:        title,
		Body:         body + " \n\n _Created from " + templateIssueURL + " by [issue-creator](https://github.com/rerost/issue-creator)_",
		Labels:       _templateIssue.Labels,
		LastIssueURL: *lastIssue.URL,
	}

	s, _ := json.Marshal(res)
	zap.L().Debug("template", zap.String("Issue", string(s)))

	return res, nil
}

func (is *issueServiceImpl) Create(ctx context.Context, templateURL string) (types.Issue, error) {
	i, err := is.render(ctx, templateURL)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	if is.checkBeforeCreateIssue != nil && *is.checkBeforeCreateIssue != "" {
		f, err := ioutil.TempFile("", "issue_creator_*.sh")
		if err != nil {
			return types.Issue{}, errors.WithStack(err)
		}

		_, err = f.WriteString(*is.checkBeforeCreateIssue)
		if err != nil {
			return types.Issue{}, errors.WithStack(err)
		}
		f.Chmod(0755)
		f.Close()

		out, err := exec.Command("bash", f.Name()).Output()
		if err != nil {
			zap.L().Error("Failed to exec check before create issue", zap.String("out", string(out)), zap.String("err", err.Error()))
			return types.Issue{}, errors.WithStack(err)
		}
	}

	created, err := is.ir.Create(ctx, i)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	if !is.closeLastIssue {
		return created, nil
	}

	_, err = is.ir.CloseByURL(ctx, i.LastIssueURL)
	if err != nil {
		return types.Issue{}, errors.WithStack(err)
	}

	return created, nil
}

func (is *issueServiceImpl) Render(ctx context.Context, templateURL string) (types.Issue, error) {
	return is.render(ctx, templateURL)
}
