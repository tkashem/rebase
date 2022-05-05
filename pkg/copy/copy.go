package copy

import (
	"fmt"
	"github.com/tkashem/rebase/pkg/git"
	"k8s.io/klog/v2"
)

func New(target string, sourceHeadSHA string, sourceMarker string) (*cmd, error) {
	accessor, err := git.Initialize(target)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git - %w", err)
	}

	marker := accessor.Marker
	if len(sourceMarker) > 0 {
		marker = fmt.Sprintf("openshift-rebase(%s):marker", sourceMarker)
	}
	klog.InfoS("looking for rebase marker for the source branch", "pattern", marker)
	sourceStopAt, err := accessor.Git.FindRebaseMarkerCommit(sourceHeadSHA, accessor.Marker)
	if err != nil {
		return nil, err
	}
	klog.InfoS("found rebase marker for the source branch", "commit", sourceStopAt.Message)

	return &cmd{
		copier: &copier{
			accessor:        accessor,
			sourceHeadSHA:   sourceHeadSHA,
			sourceStopAtSHA: sourceStopAt.Hash.String(),
		},
	}, nil
}

type cmd struct {
	copier *copier
}

func (c *cmd) Run() error {
	return c.copier.copyAll()
}
