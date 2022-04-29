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
	override                 carry.Prompt
	git                      git.Git
	github                   git.GitHub
	prompt                   carry.Prompt
	target, marker, metadata string

	// filled by Init
	stopAtSHA string

	cherryPickFromSHA, cherryStopAtSHA string
}

func (s *processor) Init() error {
	if err := s.git.CheckRemotes(); err != nil {
		return fmt.Errorf("git repo not setup properly: %v", err)
	}

	klog.InfoS("apply in progress", "target", s.target, "marker", s.marker, "rebase-marker-sha",
		s.stopAtSHA, "commit-amend-metadata", s.metadata, "pick-cherry-picks-from", s.cherryPickFromSHA)

	return nil
}

func (s *processor) Done() error {
	klog.InfoS("apply has completed")
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
	commits, err := s.git.Log("", s.stopAtSHA)
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

func (s *processor) findCherryPickedCommit(r *carry.Commit) (string, error) {
	if len(s.cherryPickFromSHA) == 0 {
		return "", nil
	}

	commits, err := s.git.Log(s.cherryPickFromSHA, s.cherryStopAtSHA)
	if err != nil {
		return "", fmt.Errorf("git log failed with error: %w", err)
	}

	marker := fmt.Sprintf("%s=%s", s.metadata, r.SHA)
	for _, commit := range commits {
		if strings.Contains(commit.Message, marker) {
			return commit.Hash.String(), nil
		}
	}

	return "", nil
}

func (s *processor) apply(r *carry.Commit, cherrypick bool) error {
	if cherrypick {
		if err := s.git.CherryPick(r.SHA); err != nil {
			// the cherry pick failed, possibly due to a conflict
			// is there a branch from where we can pick it up?
			var success bool
			var cherryPickCommitSHA string
			if cherryPickCommitSHA, err = s.findCherryPickedCommit(r); err != nil {
				klog.Infof("did not find cherry-picked commit - %v", err)
				return &CherryPickError{gitErr: err, message: r.ShortString()}
			}

			if len(cherryPickCommitSHA) > 0 {
				klog.InfoS("found a resolved commit, going to cherry pick", "sha", cherryPickCommitSHA)
				s.git.AbortCherryPick()
				if err := s.git.CherryPick(cherryPickCommitSHA); err != nil {
					return &CherryPickError{gitErr: err, message: r.ShortString()}
				}
				// successfully picked.
				success = true
			}

			if !success {
				return &CherryPickError{gitErr: err, message: r.ShortString()}
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
	if drop := s.override.ShouldDrop(r.SHA); drop {
		klog.Infof("status=drop(override) do=skip - %s", r.ShortString())
		return nil
	}

	klog.Infof("type=%s do=? - %s", r.CommitType, r.ShortString())
	drop, err := prompt(fmt.Sprintf("do you want to drop(%s)?[Yes/No]:", r.SHA))
	if err != nil {
		return err
	}

	if drop {
		klog.Infof("status=drop(prompt) do=skip - %s", r.ShortString())
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
	case answer == "no" || answer == "n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid answer: %s", answer)
	}
}

func removePreviousRebaseMetadata(msg string) string {
	// TODO: remove existing rebase metadata tag if any
	return msg
}
