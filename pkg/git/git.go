package git

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	gitv5 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitv5object "github.com/go-git/go-git/v5/plumbing/object"
	"k8s.io/klog/v2"
)

type Git interface {
	CheckRemotes() error
	FindRebaseMarkerCommit(from string, marker string) (*gitv5object.Commit, error)
	Head() (*gitv5object.Commit, error)
	Log(from string, stopAtHash string) ([]*gitv5object.Commit, error)
	CherryPick(sha string) error
	AmendCommitMessage(f func(string) []string) error
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

func (git *git) CheckRemotes() error {
	for _, remote := range []struct {
		name string
		path string
	}{
		{
			name: "openshift",
			path: "github.com:openshift/kubernetes.git",
		},
		{
			name: "upstream",
			path: "github.com:kubernetes/kubernetes.git",
		},
	} {
		fetchURL, err := git.fetchURLForRemote(remote.name)
		if err != nil {
			return err
		}
		if !strings.Contains(fetchURL, remote.path) {
			return fmt.Errorf("fetch URL does not match, remote=%s path=%s", remote.name, remote.path)
		}
		klog.InfoS("git remote setup properly", "remote", remote.name, "fetch-url", fetchURL)
	}
	return nil
}

func (git *git) FindRebaseMarkerCommit(from string, marker string) (*gitv5object.Commit, error) {
	o := &gitv5.LogOptions{}
	if len(from) > 0 {
		o.From = plumbing.NewHash(from)
	}
	iter, err := git.repository.Log(o)
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	defer iter.Close()
	for {
		commit, err := iter.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to find commit with marker: %s - %w", marker, err)
		}

		if strings.Contains(commit.Message, marker) {
			return commit, nil
		}
	}

	return nil, fmt.Errorf("failed to find commit with marker: %s", marker)
}

func (git *git) Log(from, stopAtHash string) ([]*gitv5object.Commit, error) {
	o := &gitv5.LogOptions{}
	if len(from) > 0 {
		o.From = plumbing.NewHash(from)
	}

	iter, err := git.repository.Log(o)
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

func (git *git) CherryPick(sha string) error {
	// skipping --strategy-option=ours
	cmd := exec.Command("git", "cherry-pick", "--allow-empty", sha)

	var stdoutStderr []byte
	var err error

	klog.InfoS("executing cherry-pick", "command", cmd.String())
	defer func() {
		if len(stdoutStderr) > 0 {
			defer klog.Infof(">>>>>>>>>>>>>>>>>>>> OUTPUT: END >>>>>>>>>>>>>>>>>>>>>>\n")
			klog.Infof("<<<<<<<<<<<<<<<<<<<< OUTPUT: START <<<<<<<<<<<<<<<<<<<<\n%s", stdoutStderr)
		}
	}()

	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git cherry-pick failed: %w", err)
	}
	return nil
}

func (git *git) AmendCommitMessage(f func(string) []string) error {
	var err error
	current, err := git.getCommitMessageAtHead()
	if err != nil {
		return err
	}

	args := []string{"commit", "--allow-empty", "--amend"}
	for _, msg := range f(current) {
		args = append(args, "-m", msg)
	}

	cmd := exec.Command("git", args...)
	klog.InfoS("amend commit message", "command", cmd.String())

	var stdoutStderr []byte
	defer func() {
		if len(stdoutStderr) > 0 {
			defer klog.Infof(">>>>>>>>>>>>>>>>>>>> OUTPUT: END >>>>>>>>>>>>>>>>>>>>>>\n")
			klog.Infof("<<<<<<<<<<<<<<<<<<<< OUTPUT: START <<<<<<<<<<<<<<<<<<<<\n%s", stdoutStderr)
		}
	}()

	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git cherry-pick failed: %w", err)
	}
	return nil
}

func (git *git) Head() (*gitv5object.Commit, error) {
	reference, err := git.repository.Head()
	if err != nil {
		return nil, err
	}

	commit, err := git.repository.CommitObject(reference.Hash())
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func (git *git) getCommitMessageAtHead() (string, error) {
	reference, err := git.repository.Head()
	if err != nil {
		return "", err
	}

	commit, err := git.repository.CommitObject(reference.Hash())
	if err != nil {
		return "", err
	}

	return commit.Message, nil
}

func (git *git) fetchURLForRemote(remoteName string) (string, error) {
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
