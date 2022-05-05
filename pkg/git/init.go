package git

import (
	"fmt"
	"k8s.io/klog/v2"
	"os"
)

type Accessor struct {
	Git    Git
	GitHub GitHub

	Target          string
	Marker          string
	MetadataSource  string
	StopAtCommitSHA string
}

func Initialize(target string) (*Accessor, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	klog.InfoS("working directory is set", "working-directory", workingDir)

	gitAPI, err := OpenGit(workingDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open gitAPI workspace at %q - %w", workingDir, err)
	}

	klog.InfoS("opened gitAPI repository successfully", "working-directory", workingDir)
	klog.InfoS("rebase target", "version", target)

	if err := gitAPI.CheckRemotes(); err != nil {
		return nil, fmt.Errorf("git repo not setup properly: %v", err)
	}

	githubAPI, err := NewGitHubClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create githubAPI client - %w", err)
	}

	marker := fmt.Sprintf("openshift-rebase(%s):marker", target)

	// let's find the rebase marker
	klog.InfoS("looking for rebase marker", "pattern", marker)
	stopAtCommit, err := gitAPI.FindRebaseMarkerCommit("", marker)
	if err != nil {
		return nil, fmt.Errorf("rebase marker not found, this branch is not properly setup for rebase - %w", err)
	}
	klog.InfoS("found rebase marker", "commit", stopAtCommit.Message)

	return &Accessor{
		Git:             gitAPI,
		GitHub:          githubAPI,
		Target:          target,
		Marker:          marker,
		MetadataSource:  fmt.Sprintf("openshift-rebase(%s):source", target),
		StopAtCommitSHA: stopAtCommit.Hash.String(),
	}, nil
}
