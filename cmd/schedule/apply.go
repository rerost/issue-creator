package schedule

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/domain/schedule"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
)

func NewApplyCommand(ctx context.Context, templateFilePath string, srv schedule.ScheduleService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply schedule",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			b, err := ioutil.ReadFile(templateFilePath)
			if err != nil {
				errors.WithStack(err)
			}

			templateFile := string(b)
			zap.L().Debug("template file", zap.String("templateFile", templateFile))

			err = srv.Apply(ctx, templateFile, args[0], args[1])
			if err != nil {
				return errors.WithStack(err)
			}
			fmt.Println("Applyed")
			return nil
		},
	}

	return cmd
}
