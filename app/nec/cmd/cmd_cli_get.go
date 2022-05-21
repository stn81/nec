package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/stn81/kate/rdb"

	"go.uber.org/zap"

	"github.com/stn81/nec/config"
	"github.com/spf13/cobra"
)

var CliGetFlags = &cliGetFlags{}

type cliGetFlags struct {
	SaveFile string
}

func NewCliGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get KEY",
		Short: "get redis key",
		Run:   cliGetCmdFunc,
	}
	cmd.Flags().StringVarP(&CliGetFlags.SaveFile, "save_file", "s", "", "save redis value to file")
	return cmd
}

func cliGetCmdFunc(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: get KEY")
		os.Exit(1)
	}

	key := args[0]

	logger, err := initStdLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "create std logger failed: %v", err)
		os.Exit(1)
	}

	if err := config.Load(GlobalFlags.ConfigFile); err != nil {
		logger.Fatal("load config failed", zap.String("file", GlobalFlags.ConfigFile), zap.Error(err))
	}

	rdb.Init(config.Redis.Config)
	defer rdb.Uninit()

	client := rdb.Get()
	value, err := client.Get(key).Bytes()
	if err != nil {
		logger.Fatal("failed to get redis key",
			zap.String("key", key),
			zap.Error(err),
		)
	}
	fmt.Printf("=============%s=============\n", key)
	fmt.Println(string(value))

	if CliGetFlags.SaveFile != "" {
		if err = ioutil.WriteFile(CliGetFlags.SaveFile, value, 0666); err != nil {
			logger.Fatal("failed to save value to file",
				zap.String("file", CliGetFlags.SaveFile),
				zap.Error(err),
			)
		}
		logger.Info("save value to file success",
			zap.String("file", CliGetFlags.SaveFile),
			zap.Int("content_length", len(value)),
		)
	}
}
