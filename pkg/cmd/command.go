package cmd

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

type Runner interface {
	Run() error
}

type Options struct {
	CarryCommitLogFilePath string
	OverrideFilePath       string
	Target                 string
}

func (o *Options) AddFlags(flags *flag.FlagSet) {
	flags.StringVar(&o.CarryCommitLogFilePath, "carry-commit-file", o.CarryCommitLogFilePath, "file containing all commit logs")
	flags.StringVar(&o.OverrideFilePath, "overrides", o.OverrideFilePath, "path to file that contains overrides")
	flags.StringVar(&o.Target, "target", o.Target, "rebase target, ie. v1.24")
}

func (o *Options) Validate() error {
	if err := isFile(o.CarryCommitLogFilePath); err != nil {
		return err
	}
	if len(o.Target) == 0 {
		return fmt.Errorf("must be a valid value ie. v1.24")
	}

	if len(o.OverrideFilePath) > 0 {
		return isFile(o.OverrideFilePath)
	}

	return nil
}

func isFile(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("invalid path: %q - %w", path, err)
	}
	if stat.IsDir() {
		return fmt.Errorf("must be a file: %q", err)
	}

	return nil
}
