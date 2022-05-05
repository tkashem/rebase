package verify

import (
	"github.com/tkashem/rebase/pkg/carry"
	"github.com/tkashem/rebase/pkg/git"
)

// it outputs the original commit summaries
type original struct {
	git     git.Git
	carries []*carry.CommitSummary
}

func (o *original) Transform() ([]descriptor, error) {
	carries := make([]descriptor, 0)
	for i, carry := range o.carries {
		carries = append(carries, descriptor{
			Order:   i,
			Action:  carry.OriginalType,
			Commit:  carry.SHA,
			Message: carry.MessageWithPrefix,
		})
	}
	return nil, nil
}

func (o *original) action(summary *carry.CommitSummary) (string, error) {
	switch {
	case summary.OriginalType == "drop":
	}
	return summary.OriginalType, nil
}
