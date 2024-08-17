package repo_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rerost/issue-creator/repo"
)

func TestScheduleRepository(t *testing.T) {
	t.Parallel()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	testCase := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "normal",
			in:   "test",
			out:  "test",
		},
		{
			name: "not escape",
			in:   "<>.-",
			out:  "<>.-",
		},
	}

	for _, test := range testCase {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			output, err := os.CreateTemp(currentDir, "output")
			if err != nil {
				t.Error(err)
			}
			output.Close()
			defer os.Remove(output.Name())

			repo := repo.NewScheduleRepository(
				[]string{
					"./dummy_apply_cmd.sh",
					output.Name(),
				},
			)

			err = repo.Apply(ctx, test.in)
			if err != nil {
				t.Error(err)
			}

			out, err := os.ReadFile(output.Name())
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(string(out), test.out); diff != "" {
				t.Error(diff)
			}

		})
	}
}
