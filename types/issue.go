package types

import (
	"github.com/google/go-github/github"
)

type Issue struct {
	Owner      string
	Repository string

	Title        string
	Body         string
	Labels       []string
	URL          *string // nil when befor create
	LastIssueURL string
}

func FromGithubLabels(labels []github.Label) []string {
	ls := make([]string, len(labels))
	for index, label := range labels {
		ls[index] = label.GetName()
	}
	return ls
}
