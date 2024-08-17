package types

import "github.com/google/go-github/v63/github"

type Issue struct {
	Owner      string
	Repository string

	Title        string
	Body         string
	Labels       []string // NOTE: Discussionの場合はID, Isssueの場合はlabel名が入っている
	URL          *string  // nil when befor create
	LastIssueURL string
	Meta         *map[string]string
}

// TODO 名前が適切でないので修正する。*github.Label -> string
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
