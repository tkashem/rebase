package copy

import (
	"bufio"
	"fmt"
	"strings"

	gitv5object "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/tkashem/rebase/pkg/git"
	"k8s.io/klog/v2"
)

type copier struct {
	accessor        *git.Accessor
	sourceHeadSHA   string
	sourceStopAtSHA string
}

func (c *copier) copyAll() error {
	klog.InfoS("copy in progress", "target", c.accessor.Target, "marker", c.accessor.Marker, "rebase-marker-sha",
		c.accessor.StopAtCommitSHA, "commit-amend-metadata", c.accessor.MetadataSource, "pick-cherry-picks-from", c.sourceStopAtSHA)

	// this is the list of commits picked in the source branch
	sourceCommits, err := c.accessor.Git.Log(c.sourceHeadSHA, c.sourceStopAtSHA)
	if err != nil {
		return err
	}
	// the last commit is the marker commit, we can exclude it
	if len(sourceCommits) > 0 {
		sourceCommits = sourceCommits[0 : len(sourceCommits)-1]
	}

	klog.InfoS("copying commits", "count", len(sourceCommits))

	for i := len(sourceCommits) - 1; i >= 0; i-- {
		commit := sourceCommits[i]
		copied, err := c.copied(commit)
		if err != nil {
			return err
		}
		if copied {
			continue
		}

		if err := c.copy(commit); err != nil {
			return err
		}
	}

	return nil
}

func (c *copier) getMetadataSource(msg string) string {
	reader := strings.NewReader(msg)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, c.accessor.MetadataSource) {
			return line
		}
	}

	return ""
}

func (c *copier) copied(source *gitv5object.Commit) (bool, error) {
	commits, err := c.accessor.Git.Log("", c.accessor.StopAtCommitSHA)
	if err != nil {
		return false, fmt.Errorf("git log failed with error: %w", err)
	}

	// is the source commit a carry from the previous version?
	carry := c.getMetadataSource(source.Message)
	for _, commit := range commits {
		if (len(carry) > 0 && strings.Contains(commit.Message, carry)) ||
			strings.Contains(commit.Message, source.Message) {
			return true, nil
		}
	}

	return false, nil
}

func (c *copier) copy(source *gitv5object.Commit) error {
	if err := c.accessor.Git.CherryPick(source.Hash.String()); err != nil {
		return err
	}

	return nil
}
