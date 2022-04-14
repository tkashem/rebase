package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	pkgcmd "github.com/tkashem/rebase/pkg/cmd"
)

func main() {
	flags := pflag.NewFlagSet("rebase-helper", pflag.ExitOnError)
	pflag.CommandLine = flags

	klog.InitFlags(nil)

	root := NewRootCommand()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "rebase",
		Short:        "rebase helper",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(pkgcmd.NewApplyCommand())
	return cmd
}
