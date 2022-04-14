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

type DoFunc func(*carry.Record) error

type Processor interface {
	Init() error
	Done() error
	Step(*carry.Record) (DoFunc, error)
}

func New(reader carry.Reader, target string, overrideFilePath string) (*cmd, error) {
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

	overrider, err := newOverrider(overrideFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load commit overrides from %q - %w", overrideFilePath, err)
	}

	return &cmd{
		reader: reader,
		processor: &processor{
			git:    gitAPI,
			github: githubAPI,
			target: target,
		},
		overrider: overrider,
	}, nil
}

type cmd struct {
	reader    carry.Reader
	overrider Overrider
	processor Processor
}

func (c *cmd) Run() error {
	commits, err := c.reader.Read()
	if err != nil {
		return err
	}

	// apply override, before we start processing
	c.overrider.Override(commits)
	return process(c.processor, commits)
}

func process(p Processor, commits []*carry.Record) error {
	if err := p.Init(); err != nil {
		return fmt.Errorf("initialization failed with: %w", err)
	}

	// the logs are in right order, the oldest commit should be applied first
	for i, _ := range commits {
		commit := commits[i]
		doFn, err := p.Step(commit)
		if err != nil {
			return err
		}
		if err := doFn(commit); err != nil {
			return err
		}
	}

	if err := p.Done(); err != nil {
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
