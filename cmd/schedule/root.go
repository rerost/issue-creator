package schedule

import (
	"context"

	"github.com/rerost/issue-creator/domain/schedule"
	"github.com/spf13/cobra"
)

func NewScheduleCommand(ctx context.Context, templateFile string, srv schedule.ScheduleService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Schedule create github issue",
	}

	cmd.AddCommand(
		NewRenderCommand(ctx, templateFile, srv),
		NewApplyCommand(ctx, templateFile, srv),
	)

	return cmd
}
