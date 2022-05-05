package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	flag "github.com/spf13/pflag"
	"github.com/tkashem/rebase/pkg/copy"
)

type CopyOptions struct {
	Target        string
	SourceHeadSHA string
	SourceMarker  string
}

func NewCopyCommand() *cobra.Command {
	options := &CopyOptions{}

	cmd := &cobra.Command{
		Use:          "copy --target=v.1.24 --source={SHA of the head of the source branch}",
		Short:        "Iterates through the specified commit log file and applies each commit.",
		Example:      "",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			var runner Runner
			var err error
			if runner, err = copy.New(options.Target, options.SourceHeadSHA, options.SourceMarker); err != nil {
				return err
			}

			if err := runner.Run(); err != nil {
				klog.ErrorS(err, "apply failed")
				return err
			}

			return nil
		},
	}

	flag.StringVar(&options.Target, "target", options.Target, "rebase target, ie. v1.24")
	flag.StringVar(&options.SourceMarker, "source-marker", options.SourceMarker, "source marker, ie. v1.24.0-rc.1")
	flag.StringVar(&options.SourceHeadSHA, "source", options.SourceHeadSHA, "SHA pointing to the HEAD of the branch from where to copy")

	return cmd
}
