package schedule

import (
	"context"

	"github.com/rerost/issue-creator/domain/schedule"
	"github.com/spf13/cobra"
)

func NewRenderCommand(ctx context.Context, srv schedule.ScheduleService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply schedule",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
