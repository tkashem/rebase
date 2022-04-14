package verify

import (
	"fmt"
	"os"

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
		reader: reader,
		git:    gitAPI,
		target: target,
	}, nil
}

type cmd struct {
	reader carry.CommitReader
	git    git.Git
	target string
}

func (c *cmd) Run() error {
	if err := c.git.CheckRemotes(); err != nil {
		return fmt.Errorf("git repo not setup properly: %v", err)
	}

	marker := fmt.Sprintf("openshift-rebase-marker:%s", c.target)
	klog.InfoS("rebase marker", "pattern", marker)

	markerCommit, err := c.git.FindRebaseMarkerCommit(marker)
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

	return verify(carries, picked)
}

func verify(carries []*carry.Commit, picked []*gitv5object.Commit) error {
	return nil
}
