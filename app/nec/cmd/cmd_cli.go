package cmd

import (
	"github.com/spf13/cobra"
)

var CliFlags = &cliFlags{}

type cliFlags struct {
	Addr     string
	DumpPath string
}

func NewCliCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "cli",
		Short:      "Client",
		SuggestFor: []string{"cli"},
	}

	cmd.AddCommand(
		NewCliSetCmd(),
		NewCliSetexCmd(),
		NewCliHsetCmd(),
		NewCliGetCmd(),
		NewCliHgetCmd(),
	)

	cmd.PersistentFlags().StringVarP(&CliFlags.Addr, "addr", "a", ":9090", "ip:port")
	cmd.PersistentFlags().StringVarP(&CliFlags.DumpPath, "dump_path", "d", "", "dump request to file")
	return cmd
}
