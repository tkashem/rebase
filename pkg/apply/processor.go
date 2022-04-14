package apply

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/tkashem/rebase/pkg/carry"
	"github.com/tkashem/rebase/pkg/git"
	"k8s.io/klog/v2"
)

type CherryPickError struct {
	message string
	gitErr  error
}

func (e *CherryPickError) Unwrap() error { return e.gitErr }
func (e *CherryPickError) Error() string {
	return fmt.Sprintf("%s - %v", e.message, e.gitErr)
}

type processor struct {
	git    git.Git
	github git.GitHub
	target string

	// filled by Init
	stopAtSHA string
}

func (s *processor) Init() error {
	if err := s.git.CheckRemotes(); err != nil {
		return fmt.Errorf("git repo not setup properly: %v", err)
	}

	// let's find the rebase marker
	marker := fmt.Sprintf("openshift-rebase-marker:%s", s.target)
	klog.InfoS("looking for rebase marker", "pattern", marker)

	stopAtCommit, err := s.git.FindRebaseMarkerCommit(marker)
	if err != nil {
		return err
	}

	s.stopAtSHA = stopAtCommit.Hash.String()
	klog.InfoS("apply in progress", "target", s.target, "rebase-marker-sha",
		s.stopAtSHA, "message", stopAtCommit.Message)

	return nil
}

func (s *processor) Done() error {
	return nil
}

func (s *processor) Step(r *carry.Commit) (DoFunc, error) {
	switch {
	case r.CommitType == "drop":
		return s.drop, nil
	case r.CommitType == "revert":
		return s.revert, nil
	case r.CommitType == "carry":
		return s.carry, nil
	case len(r.CommitType) > 0:
		return s.pick, nil
	}

	return nil, fmt.Errorf("invalid commit type: %s", r.CommitType)
}

func (s *processor) exists(r *carry.Commit) (bool, error) {
	commits, err := s.git.Log(s.stopAtSHA)
	if err != nil {
		return false, fmt.Errorf("git log failed with error: %w", err)
	}

	var found bool
	for _, commit := range commits {
		if strings.Contains(commit.Message, fmt.Sprintf("openshift-rebase:source=%s", r.SHA)) &&
			strings.Contains(commit.Message, r.MessageWithPrefix) {
			found = true
		}
	}

	return found, nil
}

func (s *processor) apply(r *carry.Commit) error {
	if err := s.git.CherryPick(r.SHA); err != nil {
		return &CherryPickError{
			gitErr:  err,
			message: r.ShortString(),
		}
	}

	// chery-pick succeeded, now we need to append rebase metadata
	// to the commit message
	metadata := fmt.Sprintf("openshift-rebase:source=%s", r.SHA)
	if err := s.git.AmendCommitMessage(metadata); err != nil {
		return fmt.Errorf("failed to amend commit message with rebase metadata - %w", err)
	}

	return nil
}

func (s *processor) carry(r *carry.Commit) error {
	picked, err := s.exists(r)
	if err != nil {
		return err
	}
	if picked {
		klog.Infof("status=picked-in-branch do=noop - %s", r.ShortString())
		return nil
	}

	klog.Infof("status=not-picked-in-branch do=cherry-pick - %s", r.ShortString())
	if err := s.apply(r); err != nil {
		return err
	}
	return nil
}

func (s *processor) pick(r *carry.Commit) error {
	merged, err := s.github.IsPRMerged(r.UpstreamPR)
	if err != nil {
		return err
	}
	if merged {
		klog.Infof("status=merged(upstream) do=skip - %s", r.ShortString())
		return nil
	}

	picked, err := s.exists(r)
	if err != nil {
		return err
	}
	if picked {
		klog.Infof("status=picked-in-branch do=noop - %s", r.ShortString())
		return nil
	}

	klog.Infof("status=not-merged(upstream) do=cherry-pick - %s", r.ShortString())
	if err := s.apply(r); err != nil {
		return err
	}
	return nil
}

func (s *processor) drop(r *carry.Commit) error {
	klog.Infof("status= do=? - %s", r.ShortString())
	drop, err := prompt(fmt.Sprintf("do you want to drop(%s)?[Yes/No]:", r.SHA))
	if err != nil {
		return err
	}

	if drop {
		klog.Infof("status= do=skip - %s", r.ShortString())
		return nil
	}

	return s.carry(r)
}

func (s *processor) revert(r *carry.Commit) error {
	return s.carry(r)
}

func prompt(msg string) (bool, error) {
	fmt.Print(msg)
	reader := bufio.NewReader(os.Stdin)

	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	answer = strings.ToLower(strings.TrimSuffix(answer, "\n"))
	switch {
	case answer == "yes" || answer == "y":
		return true, nil
	default:
		return false, nil
	}
}
