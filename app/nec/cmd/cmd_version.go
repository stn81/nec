package cmd

import (
	"github.com/stn81/kate/app"
	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run:   versionCmdFunc,
	}
	return cmd
}

func versionCmdFunc(cmd *cobra.Command, args []string) {
	app.PrintVersion()
}
