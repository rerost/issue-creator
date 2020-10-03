package schedule

import (
	"bytes"
	"context"
	"html/template"
	"net/url"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/repo"
	"go.uber.org/zap"
)

type ScheduleService interface {
	Render(ctx context.Context, templateFile string, schedule string, templateIssueURL string) (string, error)
	Apply(ctx context.Context, templateFile string, schedule string, templateIssueURL string) error
}

type TemplateData struct {
	Name     string
	Schedule string
	Commands []string
}

type scheduleServiceImpl struct {
	sr             repo.ScheduleRepository
	closeLastIssue bool
}

func NewScheduleService(scheduleRepository repo.ScheduleRepository, closeLastIssue bool) ScheduleService {
	return &scheduleServiceImpl{
		sr:             scheduleRepository,
		closeLastIssue: closeLastIssue,
	}
}

func (s *scheduleServiceImpl) Render(ctx context.Context, templateFile string, schedule string, templateIssueURL string) (string, error) {
	if valid := CheckSchedule(schedule); !valid {
		return "", errors.New("schedule is not valid")
	}

	scheduleName, err := ConvertToName(templateIssueURL)
	if err != nil {
		return "", errors.WithStack(err)
	}

	commands := []string{"issue-creator", "create", templateIssueURL}
	if s.closeLastIssue {
		commands = append(commands, "--CloseLastIssue")
	}

	templateData := TemplateData{
		Name:     scheduleName,
		Schedule: schedule,
		Commands: commands,
	}

	manifestTpl, err := template.New("manifest").Funcs(sprig.FuncMap()).Parse(templateFile)
	if err != nil {
		return "", errors.Wrap(err, "Failed to parse manifest template")
	}

	w := bytes.NewBufferString("")
	err = manifestTpl.Execute(w, templateData)
	if err != nil {
		return "", errors.Wrap(err, "Failed to render manifest")
	}

	return w.String(), nil
}

func (s *scheduleServiceImpl) Apply(ctx context.Context, templateFile string, schedule string, templateIssueURL string) error {
	manifest, err := s.Render(ctx, templateFile, schedule, templateIssueURL)
	if err != nil {
		return errors.Wrap(err, "Failed to render manifest")
	}
	err = s.sr.Apply(ctx, manifest)
	if err != nil {
		return errors.Wrap(err, "Failed to apply manifest")
	}
	return nil
}

func CheckSchedule(schedule string) bool {
	schedules := strings.Split(schedule, " ")

	return len(schedules) == 5
}

func ConvertToName(templateIssueURL string) (string, error) {
	zap.L().Debug("templateIssueURL", zap.String("templateIssueURL", templateIssueURL))

	p, err := url.Parse(templateIssueURL)
	if err != nil {
		return "", errors.WithStack(err)
	}

	path := p.Path
	name := ""
	for _, p := range strings.Split(path, "/") {
		if len(p) == 0 {
			continue
		}
		if len(name) == 0 {
			name += p
			continue
		}
		name += "-" + p
	}

	return name, nil
}
