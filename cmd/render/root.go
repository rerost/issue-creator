package render

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rerost/issue-creator/domain/issue"
	"github.com/spf13/cobra"
)

func NewRenderCommand(ctx context.Context, srv issue.IssueService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render",
		Short: "Render github issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			result, err := srv.Render(ctx, args[0])
			if err != nil {
				return errors.WithStack(err)
			}

			br, err := json.Marshal(result)
			if err != nil {
				return errors.WithStack(err)
			}
			fmt.Println(string(br))
			return nil
		},
	}

	return cmd
}
