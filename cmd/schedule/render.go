package schedule

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/domain/schedule"
	"github.com/spf13/cobra"
)

func NewRenderCommand(ctx context.Context, templateFile string, srv schedule.ScheduleService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply schedule",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
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
