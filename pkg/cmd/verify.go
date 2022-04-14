package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"os"

	"github.com/tkashem/rebase/pkg/carry"
	"github.com/tkashem/rebase/pkg/verify"
)

type VerifyOptions struct {
	Target                 string
	CarryCommitLogFilePath string
}

func NewVerifyCommand() *cobra.Command {
	options := &VerifyOptions{}

	cmd := &cobra.Command{
		Use:          "verify --target=v.1.24 --carry-commit-file={carry-commit-log-file-path}",
		Short:        "Iterates through the carry commits picked in the branch and compares",
		Example:      "",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := options.Validate(); err != nil {
				return err
			}

			// TODO: today the carries are obtained from a CSV file, but in
			//  the following rebase we can generate them on the fly using
			//  the openshift rebase marker
			reader, err := carry.NewCommitReaderFromFile(options.CarryCommitLogFilePath)
			if err != nil {
				return err
			}

			var runner Runner
			if runner, err = verify.New(reader, options.Target); err != nil {
				return err
			}

			if err := runner.Run(); err != nil {
				klog.ErrorS(err, "verify failed")
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&options.CarryCommitLogFilePath, "carry-commit-file", options.CarryCommitLogFilePath, "file containing all commit logs")
	cmd.Flags().StringVar(&options.Target, "target", options.Target, "rebase target, ie. v1.24")

	return cmd
}

func (o *VerifyOptions) Validate() error {
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
