package verify

import (
	gitv5object "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/tkashem/rebase/pkg/git"
)

// it outputs the commit summaries as seen in the current branch
type actual struct {
	git     git.Git
	carries []*gitv5object.Commit
}

func (a *actual) Transform() ([]descriptor, error) {
	return nil, nil
}
