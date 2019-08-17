package schedule

import (
	"bytes"
	"context"
	"html/template"
	"strings"

	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/repo"
)

type ScheduleService interface {
	Render(ctx context.Context, templateFile string, schedule string, templateIssueURL string) (string, error)
	Apply(ctx context.Context, templateFile string, schedule string, templateIssueURL string) error
}

type TemplateData struct {
	Schedule string
	Commands []string
}

type scheduleServiceImpl struct {
	sr repo.ScheduleRepository
}

func NewScheduleService(scheduleRepository repo.ScheduleRepository) ScheduleService {
	return &scheduleServiceImpl{
		sr: scheduleRepository,
	}
}

func (s *scheduleServiceImpl) Render(ctx context.Context, templateFile string, schedule string, templateIssueURL string) (string, error) {
	if valid := CheckSchedule(schedule); !valid {
		return "", errors.New("schedule is not valid")
	}

	templateData := TemplateData{
		Schedule: schedule,
		Commands: []string{"issue-creator", "create", templateIssueURL},
	}

	manifestTpl, err := template.New("manifest").Parse(templateFile)
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
