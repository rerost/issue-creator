package schedule

func NewApplyCommand(ctx context.Context, srv schedule.ScheduleService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply schedule",
		Args:  cobra.ExactArgs(1),
		RunE:  func(_ *cobra.Command, args []string) error {
		}
	}

	return cmd
}
