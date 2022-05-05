package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/tkashem/rebase/pkg/carry"
	"github.com/tkashem/rebase/pkg/verify"
)

type VerifyOptions struct {
	Options
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
			carries, err := carry.NewReaderFromFile(options.CarryCommitLogFilePath, options.OverrideFilePath)
			if err != nil {
				return err
			}

			var runner Runner
			if runner, err = verify.New(carries, options.Target); err != nil {
				return err
			}

			if err := runner.Run(); err != nil {
				klog.ErrorS(err, "verify failed")
				return err
			}

			return nil
		},
	}

	options.AddFlags(cmd.Flags())
	return cmd
}
