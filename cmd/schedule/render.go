package schedule

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/domain/schedule"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewRenderCommand(ctx context.Context, templateFilePath string, srv schedule.ScheduleService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render",
		Short: "Render schedule manifest",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			zap.L().Debug("template path", zap.String("templateFilePath", templateFilePath))
			b, err := os.ReadFile(templateFilePath)
			if err != nil {
				errors.WithStack(err)
			}

			templateFile := string(b)
			zap.L().Debug("template file", zap.String("templateFile", templateFile))

			rendered, err := srv.Render(ctx, templateFile, args[0], args[1])
			if err != nil {
				return errors.WithStack(err)
			}
			fmt.Println(rendered)
			return nil
		},
	}

	return cmd
}
