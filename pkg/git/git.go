package git

import (
	"fmt"
	"io"
	"k8s.io/klog/v2"
	"os/exec"
	"strings"

	// "k8s.io/klog/v2"
	gitv5 "github.com/go-git/go-git/v5"
	gitv5object "github.com/go-git/go-git/v5/plumbing/object"
)

type Git interface {
	FindRebaseMarkerCommitSHA(marker string) (string, error)
	FetchURLForRemote(remoteName string) (string, error)
	Log(stopAtHash string) ([]*gitv5object.Commit, error)
	CherryPick(sha string) error
}

func OpenGit(path string) (Git, error) {
	repository, err := gitv5.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return &git{repository: repository}, nil
}

type git struct {
	repository *gitv5.Repository
}

func (git *git) FindRebaseMarkerCommitSHA(marker string) (string, error) {
	iter, err := git.repository.Log(&gitv5.LogOptions{})
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}

	defer iter.Close()
	for {
		commit, err := iter.Next()
		if err != nil {
			return "", fmt.Errorf("failed to find commit with marker: %s - %w", marker, err)
		}

		if strings.Contains(commit.Message, marker) {
			return commit.Hash.String(), nil
		}
	}

	return "", fmt.Errorf("failed to find commit with marker: %s", marker)
}

func (git *git) Log(stopAtHash string) ([]*gitv5object.Commit, error) {
	iter, err := git.repository.Log(&gitv5.LogOptions{})
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	defer iter.Close()
	commits := make([]*gitv5object.Commit, 0)
	for {
		commit, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				return commits, nil
			}
			return nil, fmt.Errorf("iterating through commit log failed: %w", err)
		}

		commits = append(commits, commit)
		if commit.Hash.String() == stopAtHash {
			break
		}
	}

	return commits, nil
}

func (git *git) FetchURLForRemote(remoteName string) (string, error) {
	remote, err := git.repository.Remote(remoteName)
	if err != nil {
		return "", err
	}
	config := remote.Config()
	// URLs the URLs of a remote repository. It must be non-empty. Fetch will
	// always use the first URL, while push will use all of them.
	if len(config.URLs) == 0 {
		return "", fmt.Errorf("no fetch URLs, remote=%s", remoteName)
	}
	return config.URLs[0], nil
}

func (git *git) CherryPick(sha string) error {
	// skipping --strategy-option=ours
	cmd := exec.Command("git", "cherry-pick", "--allow-empty", sha)

	var stdoutStderr []byte
	var err error

	klog.InfoS("executing cherry-pick", "command", cmd.String())
	defer func() {
		if len(stdoutStderr) > 0 {
			defer klog.Infof(">>>>>>>>>>>>>>>>>>>> OUTPUT: END >>>>>>>>>>>>>>>>>>>>>>")
			klog.Infof("<<<<<<<<<<<<<<<<<<<< OUTPUT: START <<<<<<<<<<<<<<<<<<<<\n%s", stdoutStderr)
		}
	}()

	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git cherry-pick failed: %w", err)
	}
	return nil
}
