package schedule

import "context"

type ScheduleService interface {
	Render(ctx context.Context, templateFile string, schedule string, templateIssueURL string) (string, error)
	Apply(ctx context.Context, templateFile string, schedule string, templateIssueURL string) error
}

type TemplateData struct {
	Schedule string
	Commands  []string
}

type scheduleServiceImpl struct {
	sr ScheduleRepository
}

func NewScheduleService(scheduleRepository ScheduleRepository) ScheduleService {
	return &scheduleServiceImpl{
		sr: scheduleRepository,
	}
}

func (s *scheduleServiceImpl) Render(ctx context.Context, templateFile string, schedule string, templateIssueURL string) (string, error) {
	templateData := TemplateData{
		Schedule: schedule,
		Command: ["issue-creator", "create", templateIssueURL]
	}

	manifestTpl, err := template.New("manifest").Parse(templateFile)
	if err != nil {
		return "", errors.Wrap(err, "Failed to parse manifest template")
	}

	w := bytes.NewBufferString("")
	err = bodyTmpl.Execute(w, manifestTpl)
	if err != nil {
		return "", errors.Wrap(err, "Failed to render manifest")
	}

	return string(w), nil
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

