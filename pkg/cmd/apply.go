package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"os"

	"github.com/tkashem/rebase/pkg/apply"
	"github.com/tkashem/rebase/pkg/carry"
)

type ApplyOptions struct {
	CarryCommitLogFilePath string
	OverrideFilePath       string
	Target                 string
}

func NewApplyCommand() *cobra.Command {
	options := &ApplyOptions{}

	cmd := &cobra.Command{
		Use:          "apply --target=v.1.24 --carry-commit-file={carry-commit-log-file-path} --overrides={override file path}",
		Short:        "Iterates through the specified commit log file and applies each commit.",
		Example:      "",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := options.Validate(); err != nil {
				return err
			}

			reader, err := carry.NewCommitReaderFromFile(options.CarryCommitLogFilePath)
			if err != nil {
				return err
			}

			var runner Runner
			if runner, err = apply.New(reader, options.Target, options.OverrideFilePath); err != nil {
				return err
			}

			if err := runner.Run(); err != nil {
				klog.ErrorS(err, "apply failed")
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&options.CarryCommitLogFilePath, "carry-commit-file", options.CarryCommitLogFilePath, "file containing all commit logs")
	cmd.Flags().StringVar(&options.OverrideFilePath, "overrides", options.OverrideFilePath, "path to file that contains overrides")
	cmd.Flags().StringVar(&options.Target, "target", options.Target, "rebase target, ie. v1.24")

	return cmd
}

func (o *ApplyOptions) Validate() error {
	stat, err := os.Stat(o.CarryCommitLogFilePath)
	if err != nil {
		return fmt.Errorf("invalid path: %q - %w", o.CarryCommitLogFilePath, err)
	}
	if stat.IsDir() {
		return fmt.Errorf("must be a file: %q", err)
	}
	if len(o.Target) == 0 {
		return fmt.Errorf("must be a valid value ie. v1.24")
	}

	return nil
}
