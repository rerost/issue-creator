package repo

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ScheduleRepository interface {
	Apply(ctx context.Context, manifest string) error
}

type scheduleRepositoryImpl struct {
	k8scommands []string
}

func NewScheduleRepository(k8scommands []string) ScheduleRepository {
	return &scheduleRepositoryImpl{
		k8scommands: k8scommands,
	}
}

func (s *scheduleRepositoryImpl) Apply(ctx context.Context, manifest string) error {
	commands := s.k8scommands
	manifestFile, err := ioutil.TempFile("", "schedule.yaml")
	if err != nil {
		return errors.Wrap(err, "Failed to create temp file")
	}
	defer os.Remove(manifestFile.Name())

	_, err = manifestFile.Write([]byte(manifest))
	if err != nil {
		return errors.Wrap(err, "Failed to write manifest to file")
	}
	manifestFile.Close()

	commands = append(
		commands,
		"apply",
		"-f",
		manifestFile.Name(),
	)

	c := strings.Join(commands, " ")
	zap.L().Debug("command", zap.String("config", c))

	err = exec.CommandContext(ctx, "apply to k8s", commands...).Run()
	if err != nil {
		return errors.Wrap(err, "Failed to execute k8s command")
	}

	return nil
}
