package types_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v92/github"
	"github.com/rerost/issue-creator/types"
)

func TestFromGithubLabels(t *testing.T) {
	testCase := []struct {
		name string
		in   []*github.Label
		out  []string
	}{
		{
			name: "nil",
			in: []*github.Label{
				nil,
			},
			out: []string{},
		},
		{
			name: "one",
			in: []*github.Label{
				{
					Name: "test",
				},
			},
			out: []string{"test"},
		},
		{
			name: "include nil",
			in: []*github.Label{
				{
					Name: "",
				},
			},
			out: []string{""},
		},
	}

	for _, test := range testCase {
		test := test
		t.Run(test.name, func(t *testing.T) {
			out := types.FromGithubLabels(test.in)

			if diff := cmp.Diff(out, test.out); diff != "" {
				t.Error(diff)
				return
			}
		})
	}
}
