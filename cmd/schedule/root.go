package schedule

import (
	"context"

	"github.com/rerost/issue-creator/domain/schedule"
	"github.com/spf13/cobra"
)

func NewScheduleCommand(ctx context.Context, srv schedule.ScheduleService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Schedule create github issue",
	}

	cmd.AddCommand(
		NewRenderCommand(ctx, srv),
		NewApplyCommand(ctx, srv),
	)

	cmd.PersistentFlags().StringP("schedule", "s", "", "Schedule time(crontab)")
	cmd.PersistentFlags().StringP("manifest_template", "i", "", "manifest template")

	return cmd
}
