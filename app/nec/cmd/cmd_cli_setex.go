package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/stn81/nec/proto/proxy"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewCliSetexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setex KEY SECONDS VALUE",
		Short: "setex KEY SECONDS VALUE",
		Run:   cliSetexCmdFunc,
	}

	return cmd
}

func cliSetexCmdFunc(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: setex KEY SECONDS VALUE")
		os.Exit(1)
	}

	logger, err := initStdLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "create std logger failed: %v", err)
		os.Exit(1)
	}

	var (
		key     = []byte(args[0])
		seconds = []byte(args[1])
		value   = []byte(args[2])
		content = []byte(args[2])
	)

	if value[0] == '@' {
		if content, err = ioutil.ReadFile(string(value[1:])); err != nil {
			logger.Fatal("failed to read value from file",
				zap.String("file", string(value[1:])),
				zap.Error(err),
			)
		}
	}

	req := &proxy.Request{
		Cmd:  "setex",
		Args: [][]byte{key, seconds, content},
	}

	if CliFlags.DumpPath != "" {
		data, err := proto.Marshal(req)
		if err != nil {
			logger.Fatal("failed to marshal request", zap.Error(err))
		}
		if err := ioutil.WriteFile(CliFlags.DumpPath, data, 0644); err != nil {
			logger.Fatal("failed to write dump file", zap.String("path", CliFlags.DumpPath), zap.Error(err))
		}
	}

	conn, err := grpc.Dial(CliFlags.Addr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("failed to connect", zap.Error(err))
	}
	defer conn.Close()

	client := proxy.NewProxyClient(conn)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*time.Duration(30))
	defer cancel()

	begin := time.Now()
	if _, err = client.Do(ctx, req); err != nil {
		logger.Fatal("failed to do request", zap.Error(err))
	}

	logger.Info("proxy do success", zap.Int64("elapsed_ms", int64(time.Since(begin)/time.Millisecond)))
}
