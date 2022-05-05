package apply

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/tkashem/rebase/pkg/carry"
	"github.com/tkashem/rebase/pkg/git"
	"k8s.io/klog/v2"
)

type DoFunc func(*carry.CommitSummary) error

type Processor interface {
	Init() error
	Done() error
	Step(*carry.CommitSummary) (DoFunc, error)
}

func New(reader carry.CommitReader, override carry.Prompt, target string, cherryPickFromSHA string) (*cmd, error) {
	accessor, err := git.Initialize(target)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git - %w", err)
	}

	var cherryStopAtSHA string
	if len(cherryPickFromSHA) > 0 {
		klog.InfoS("looking for rebase marker for cherry-pick branch", "pattern", accessor.Marker)
		cherryPickStopAt, err := accessor.Git.FindRebaseMarkerCommit(cherryPickFromSHA, accessor.Marker)
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
			git:       accessor.Git,
			github:    accessor.GitHub,
			target:    target,
			marker:    accessor.Marker,
			metadata:  fmt.Sprintf("openshift-rebase(%s):source", target),
			stopAtSHA: accessor.StopAtCommitSHA,

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
