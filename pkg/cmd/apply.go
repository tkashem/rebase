package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/tkashem/rebase/pkg/apply"
	"github.com/tkashem/rebase/pkg/carry"
)

type ApplyOptions struct {
	Options
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

			reader, err := carry.NewCommitReaderFromFile(options.CarryCommitLogFilePath, options.OverrideFilePath)
			if err != nil {
				return err
			}
			override, err := carry.NewPromptsFromFile(options.OverrideFilePath)
			if err != nil {
				return err
			}

			var runner Runner
			if runner, err = apply.New(reader, override, options.Target); err != nil {
				return err
			}

			if err := runner.Run(); err != nil {
				klog.ErrorS(err, "apply failed")
				return err
			}

			return nil
		},
	}

	options.AddFlags(cmd.Flags())
	return cmd
}
