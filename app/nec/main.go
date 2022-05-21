package main

import (
	"fmt"
	"os"

	"github.com/stn81/kate/app"
	"github.com/spf13/cobra"

	"github.com/stn81/nec/app/nec/cmd"
)

var (
	rootCmd = &cobra.Command{
		Use:        app.GetName(),
		Short:      "Redis Proxy",
		SuggestFor: []string{app.GetName()},
	}
)

func main() {
	cobra.EnablePrefixMatching = true
	rootCmd.PersistentFlags().BoolVar(&cmd.GlobalFlags.Debug, "debug", false, "enable debug")
	rootCmd.PersistentFlags().StringVar(&cmd.GlobalFlags.ConfigFile, "config", app.GetDefaultConfigFile(), "config file path")

	rootCmd.AddCommand(
		cmd.NewVersionCmd(),
		cmd.NewStartCmd(),
		cmd.NewCliCmd(),
		cmd.NewFetchCmd(),
		cmd.NewOffsetCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}
