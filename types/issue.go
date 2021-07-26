package types

import "github.com/google/go-github/v37/github"

type Issue struct {
	Owner      string
	Repository string

	Title        string
	Body         string
	Labels       []string
	URL          *string // nil when befor create
	LastIssueURL string
}

func FromGithubLabels(labels []*github.Label) []string {
	ls := make([]string, 0, len(labels))
	for _, label := range labels {
		if label == nil {
			continue
		}
		ls = append(ls, label.GetName())
	}
	return ls
}
