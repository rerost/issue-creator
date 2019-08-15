package cmd

import (
	"context"

	"github.com/rerost/issue-creator/cmd/create"
	"github.com/rerost/issue-creator/cmd/render"
	cmdschedule "github.com/rerost/issue-creator/cmd/schedule"
	"github.com/rerost/issue-creator/domain/issue"
	"github.com/rerost/issue-creator/domain/schedule"
	"github.com/spf13/cobra"
)

func NewCmdRoot(
	ctx context.Context,
	issueService issue.IssueService,
	scheduleService schedule.ScheduleService,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue-creator",
		Short: "Issue create from template issue",
	}

	cmd.AddCommand(
		render.NewRenderCommand(ctx, issueService),
		create.NewCreateCommand(ctx, issueService),
		cmdschedule.NewScheduleCommand(ctx, scheduleService),
	)

	return cmd
}
