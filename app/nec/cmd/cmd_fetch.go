package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/stn81/nec/proto/proxy"

	"github.com/stn81/nec/config"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"github.com/stn81/kate/app"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var FetchFlags = &fetchFlags{}

type fetchFlags struct {
	Partition int32
	Offset    int64
	MaxBytes  int32
	DumpPath  string
}

func NewFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "fetch kafka message",
		Run:   fetchCmdFunc,
	}

	cmd.Flags().Int32VarP(&FetchFlags.Partition, "partition", "p", 0, "kafka partition")
	cmd.Flags().Int64VarP(&FetchFlags.Offset, "offset", "o", -1, "kafka offset")
	cmd.Flags().Int32VarP(&FetchFlags.MaxBytes, "maxbytes", "m", 8*1024*1024, "max bytes")
	cmd.Flags().StringVarP(&FetchFlags.DumpPath, "dump_path", "d", "", "dump value to file")
	return cmd
}

func fetchCmdFunc(cmd *cobra.Command, args []string) {
	os.Chdir(app.GetHomeDir())

	logger, err := initStdLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "create std logger failed: %v", err)
		os.Exit(1)
	}

	if err = config.Load(GlobalFlags.ConfigFile); err != nil {
		logger.Fatal("load config failed", zap.String("file", GlobalFlags.ConfigFile), zap.Error(err))
	}

	conf := sarama.NewConfig()
	conf.Version = config.Kafka.Version
	conf.ClientID = config.Kafka.ClientID

	client, err := sarama.NewClient(config.Kafka.BrokerAddrs, conf)
	if err != nil {
		logger.Fatal("failed to create sarama client", zap.Error(err))
	}
	defer client.Close()

	leader, err := client.Leader(config.Kafka.Topic, FetchFlags.Partition)
	if err != nil {
		logger.Fatal("failed to get leader of partition",
			zap.String("topic", config.Kafka.Topic),
			zap.Int32("partition", FetchFlags.Partition),
			zap.Error(err),
		)
	}

	fetchRequest := &sarama.FetchRequest{Version: 10}
	fetchRequest.AddBlock(config.Kafka.Topic, FetchFlags.Partition, FetchFlags.Offset, FetchFlags.MaxBytes)

	fetchResponse, err := leader.Fetch(fetchRequest)
	if err != nil {
		logger.Fatal("failed to fetch message",
			zap.String("topic", config.Kafka.Topic),
			zap.Int32("partition", FetchFlags.Partition),
			zap.Int64("offset", FetchFlags.Offset),
			zap.Int32("max_bytes", FetchFlags.MaxBytes),
			zap.Error(err),
		)
	}

	block := fetchResponse.GetBlock(config.Kafka.Topic, FetchFlags.Partition)

	if block.Records == nil || block.Records.RecordBatch == nil || len(block.Records.RecordBatch.Records) == 0 {
		logger.Fatal("fetch got empty records")
	}

	record := block.Records.RecordBatch.Records[0]
	recordTimestamp := fetchResponse.Timestamp.Add(record.TimestampDelta)

	req := &proxy.Request{}
	if err := proto.Unmarshal(record.Value, req); err != nil {
		logger.Fatal("failed to unmarshal message", zap.Error(err))
	}

	fmt.Printf("===========%s/%v/%v===========\n",
		config.Kafka.Topic,
		FetchFlags.Partition,
		FetchFlags.Offset,
	)
	fmt.Printf("timestamp: %v\n", recordTimestamp.Format(time.RFC3339Nano))
	fmt.Printf("key: %v\n", record.Key)
	fmt.Println("value:")
	fmt.Println(string(bytes.Join(req.Args, []byte(" "))))

	if FetchFlags.DumpPath != "" {
		if err := ioutil.WriteFile(FetchFlags.DumpPath, record.Value, 0666); err != nil {
			logger.Fatal("dump to file failed",
				zap.String("file", FetchFlags.DumpPath),
				zap.Error(err),
			)
		}
		logger.Info("dump to file success",
			zap.String("file", FetchFlags.DumpPath),
			zap.Int("content_length", len(record.Value)),
		)
	}
}
