package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/tkashem/rebase/pkg/apply"
	"github.com/tkashem/rebase/pkg/carry"
)

type PickOptions struct {
	CarryCommitLogFilePath string
	OverrideFilePath       string
	Target                 string
}

func NewPickCommand() *cobra.Command {
	options := &PickOptions{}

	cmd := &cobra.Command{
		Use:          "apply --target=v.1.24 --carry-commit-file={carry-commit-log-file-path} --overrides={override file path}",
		Short:        "Iterates through the specified commit log file and applies each commit.",
		Example:      "",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			reader, err := carry.NewCommitReaderFromFile(options.CarryCommitLogFilePath)
			if err != nil {
				return err
			}

			var runner Runner
			if runner, err = apply.New(reader, options.Target, options.OverrideFilePath); err != nil {
				return err
			}

			if err := runner.Run(); err != nil {
				klog.ErrorS(err, "exiting due to error")
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
