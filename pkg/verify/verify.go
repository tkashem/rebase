package verify

import (
	"bufio"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"os"
	"strings"

	gitv5object "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/tkashem/rebase/pkg/carry"
	"github.com/tkashem/rebase/pkg/git"
	"k8s.io/klog/v2"
)

func New(reader carry.CommitReader, target string) (*cmd, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	klog.InfoS("working directory is set", "working-directory", workingDir)

	gitAPI, err := git.OpenGit(workingDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open gitAPI workspace at %q - %w", workingDir, err)
	}
	klog.InfoS("opened gitAPI repository successfully", "working-directory", workingDir)
	klog.InfoS("rebase target", "version", target)

	return &cmd{
		reader:   reader,
		git:      gitAPI,
		target:   target,
		marker:   fmt.Sprintf("openshift-rebase(%s):marker", target),
		metadata: fmt.Sprintf("openshift-rebase(%s):source", target),
	}, nil
}

type cmd struct {
	reader                   carry.CommitReader
	git                      git.Git
	target, marker, metadata string
}

func (c *cmd) Run() error {
	if err := c.git.CheckRemotes(); err != nil {
		return fmt.Errorf("git repo not setup properly: %v", err)
	}

	klog.InfoS("rebase marker", "target", c.target, "pattern", c.marker, "metadata", c.metadata)

	markerCommit, err := c.git.FindRebaseMarkerCommit(c.marker)
	if err != nil {
		return err
	}
	klog.InfoS("verify in progress", "target", c.target,
		"rebase-marker-sha", markerCommit.Hash.String(), "message", markerCommit.Message)

	// this is our source, carry commits we want to pick in new rebase target
	carries, err := c.reader.Read()
	if err != nil {
		return err
	}

	// this is the list of commits picked in this branch
	picked, err := c.git.Log(markerCommit.Hash.String())
	if err != nil {
		return err
	}
	// the last commit is the marker commit, we can exclude it
	if len(picked) > 0 {
		picked = picked[0 : len(picked)-1]
	}

	newCarries, drops := sanitize(carries)
	klog.Infof("stats: total(%d), carries(%d), drops(%d), picked(%d)", len(carries), len(newCarries), len(drops), len(picked))
	klog.Infof("diff: \n%s", cmp.Diff(c.expected(newCarries), c.got(picked)))
	return nil
}

func sanitize(all []*carry.Commit) ([]*carry.Commit, []*carry.Commit) {
	drops := make([]*carry.Commit, 0)
	carries := make([]*carry.Commit, 0)

	for i := range all {
		if all[i].CommitType == "drop" {
			drops = append(drops, all[i])
			continue
		}
		carries = append(carries, all[i])
	}

	return carries, drops
}

func (c *cmd) got(picked []*gitv5object.Commit) []descriptor {
	ex := make([]descriptor, 0)
	for i := len(picked) - 1; i >= 0; i-- {
		split := strings.SplitN(picked[i].Message, "\n", 2)
		ex = append(ex, descriptor{Commit: c.getSourceCommitHash(picked[i].Message), Message: split[0]})
	}
	return ex
}

func (c *cmd) expected(carries []*carry.Commit) []descriptor {
	ex := make([]descriptor, 0)
	for _, carry := range carries {
		ex = append(ex, descriptor{Commit: carry.SHA, Message: carry.MessageWithPrefix})
	}
	return ex
}

func (c *cmd) getSourceCommitHash(msg string) string {
	reader := strings.NewReader(msg)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), c.metadata) {
			split := strings.Split(scanner.Text(), "=")
			if len(split) >= 2 {
				return split[1]
			}
		}
	}

	return ""
}

type descriptor struct {
	Commit  string
	Message string
}

func (d descriptor) String() string { return fmt.Sprintf("(%s): %s", d.Commit, d.Message) }
