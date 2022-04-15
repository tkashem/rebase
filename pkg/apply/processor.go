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
	git                      git.Git
	github                   git.GitHub
	prompt                   carry.Prompt
	target, marker, metadata string

	// filled by Init
	stopAtSHA string
}

func (s *processor) Init() error {
	if err := s.git.CheckRemotes(); err != nil {
		return fmt.Errorf("git repo not setup properly: %v", err)
	}

	klog.InfoS("apply in progress", "target", s.target, "marker", s.marker, "rebase-marker-sha",
		s.stopAtSHA, "commit-amend-metadata", s.metadata)

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

func (s *processor) picked(r *carry.Commit) (bool, error) {
	commits, err := s.git.Log(s.stopAtSHA)
	if err != nil {
		return false, fmt.Errorf("git log failed with error: %w", err)
	}

	for _, commit := range commits {
		if strings.Contains(commit.Message, fmt.Sprintf("%s=%s", s.metadata, r.SHA)) &&
			strings.Contains(commit.Message, r.MessageWithPrefix) {
			return true, nil
		}
	}

	return false, nil
}

func (s *processor) cherrypicked(r *carry.Commit) (bool, error) {
	head, err := s.git.Head()
	if err != nil {
		return false, fmt.Errorf("failed to get HEAD: %w", err)
	}

	if strings.Contains(head.Message, r.MessageWithPrefix) {
		// looks like conflict was resolved abd cherry-pick done
		return true, nil
	}

	return false, nil
}

func (s *processor) apply(r *carry.Commit, cherrypick bool) error {
	if cherrypick {
		if err := s.git.CherryPick(r.SHA); err != nil {
			return &CherryPickError{
				gitErr:  err,
				message: r.ShortString(),
			}
		}
	}

	// chery-pick succeeded, now we need to append rebase metadata
	// to the commit message
	if err := s.git.AmendCommitMessage(func(current string) []string {
		return []string{
			removePreviousRebaseMetadata(current),
			fmt.Sprintf("%s=%s", s.metadata, r.SHA),
		}
	}); err != nil {
		return fmt.Errorf("failed to amend commit message with rebase metadata - %w", err)
	}

	return nil
}

func (s *processor) carry(r *carry.Commit) error {
	picked, err := s.picked(r)
	if err != nil {
		return err
	}
	if picked {
		klog.Infof("status=picked-in-branch do=noop - %s", r.ShortString())
		return nil
	}

	// did cherry pick abort last time due to conflict?
	cherrypicked, err := s.cherrypicked(r)
	if err != nil {
		return err
	}

	if cherrypicked {
		klog.Infof("status=cherry-pick-completed do=apply-metadata - %s", r.ShortString())
		return s.apply(r, false)
	}

	klog.Infof("status=not-picked-in-branch do=cherry-pick - %s", r.ShortString())
	if err := s.apply(r, true); err != nil {
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
	klog.Infof("upstream PR(%s) status=not-merged - %s", r.UpstreamPR, r.MessageWithPrefix)

	return s.carry(r)
}

func (s *processor) drop(r *carry.Commit) error {
	klog.Infof("status= do=? - %s", r.ShortString())
	// do we al

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

func removePreviousRebaseMetadata(msg string) string {
	// TODO: remove existing rebase metadata tag if any
	return msg
}
