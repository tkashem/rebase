package apply

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/tkashem/rebase/pkg/carry"
	"github.com/tkashem/rebase/pkg/git"
	"k8s.io/klog/v2"
)

type DoFunc func(*carry.Commit) error

type Processor interface {
	Init() error
	Done() error
	Step(*carry.Commit) (DoFunc, error)
}

func New(reader carry.CommitReader, override carry.Prompt, target string, cherryPickFromSHA string) (*cmd, error) {
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

	githubAPI, err := git.NewGitHubClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create githubAPI client - %w", err)
	}

	marker := fmt.Sprintf("openshift-rebase(%s):marker", target)

	// let's find the rebase marker
	klog.InfoS("looking for rebase marker", "pattern", marker)
	stopAtCommit, err := gitAPI.FindRebaseMarkerCommit("", marker)
	if err != nil {
		return nil, err
	}
	klog.InfoS("found rebase marker", "commit", stopAtCommit.Message)

	var cherryStopAtSHA string
	if len(cherryPickFromSHA) > 0 {
		klog.InfoS("looking for rebase marker for cherry-pick branch", "pattern", marker)
		cherryPickStopAt, err := gitAPI.FindRebaseMarkerCommit(cherryPickFromSHA, marker)
		if err != nil {
			return nil, err
		}
		klog.InfoS("found rebase marker for cherry-pick branch", "commit", cherryPickStopAt.Message)
		cherryStopAtSHA = cherryPickStopAt.Hash.String()
	}

	return &cmd{
		reader: reader,
		processor: &processor{
			override:  override,
			git:       gitAPI,
			github:    githubAPI,
			target:    target,
			marker:    marker,
			metadata:  fmt.Sprintf("openshift-rebase(%s):source", target),
			stopAtSHA: stopAtCommit.Hash.String(),

			cherryPickFromSHA: cherryPickFromSHA,
			cherryStopAtSHA:   cherryStopAtSHA,
		},
	}, nil
}

type cmd struct {
	reader    carry.CommitReader
	processor Processor
}

func (c *cmd) Run() error {
	commits, err := c.reader.Read()
	if err != nil {
		return err
	}

	if err := c.processor.Init(); err != nil {
		return fmt.Errorf("initialization failed with: %w", err)
	}

	// the logs are in right order, the oldest commit should be applied first
	for i, _ := range commits {
		commit := commits[i]
		doFn, err := c.processor.Step(commit)
		if err != nil {
			return err
		}
		if err := doFn(commit); err != nil {
			return err
		}
	}

	if err := c.processor.Done(); err != nil {
		return fmt.Errorf("cleaup failed with: %w", err)
	}

	return nil
}

// open a browser window with the openshift commit
// 	if errors.Is(err, &CherryPickError{}) {
//		openBrowser(commit.OpenShiftCommit)
//	}
// TODO: the browser window does not open, maybe the parent go routine
//  can't quit immediately?
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		klog.ErrorS(err, "failed to open a browser window")
	}
}
