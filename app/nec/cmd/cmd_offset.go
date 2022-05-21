package cmd

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/stn81/nec/config"
	"github.com/Shopify/sarama"
	"github.com/stn81/kate/app"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var OffsetFlags = &offsetFlags{}

type offsetFlags struct {
	Action    string
	Partition int32
	Offset    int64
}

func NewOffsetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "offset",
		Short: "offset get/set tool",
		Run:   offsetCmdFunc,
	}

	cmd.Flags().StringVarP(&OffsetFlags.Action, "action", "a", "get", "action: get/set")
	cmd.Flags().Int32VarP(&OffsetFlags.Partition, "partition", "p", -1, "kafka partition")
	cmd.Flags().Int64VarP(&OffsetFlags.Offset, "offset", "o", -1, "kafka offset")

	return cmd
}

func offsetCmdFunc(cmd *cobra.Command, args []string) {
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

	partitions, err := client.Partitions(config.Kafka.Topic)
	if err != nil {
		logger.Fatal("failed to get partition list",
			zap.String("topic", config.Kafka.Topic),
			zap.Error(err),
		)
	}

	if OffsetFlags.Partition != -1 {
		partitions = []int32{OffsetFlags.Partition}
	}

	offsetManager, err := sarama.NewOffsetManagerFromClient(config.Consumer.ConsumerGroup, client)
	if err != nil {
		logger.Fatal("failed to create offset manager", zap.Error(err))
	}
	defer offsetManager.Close()

	switch OffsetFlags.Action {
	case "set":
		doOffsetSet(offsetManager, OffsetFlags.Partition, OffsetFlags.Offset, logger)
	default:
		doOffsetGet(offsetManager, partitions, logger)
	}

}

func doOffsetSet(offsetManager sarama.OffsetManager, partition int32, offset int64, logger *zap.Logger) {
	pom, err := offsetManager.ManagePartition(config.Kafka.Topic, partition)
	if err != nil {
		logger.Fatal("failed to get partition offset manager",
			zap.String("topic", config.Kafka.Topic),
			zap.Int32("partition", partition),
			zap.Error(err),
		)
	}

	nextOffset, _ := pom.NextOffset()

	if offset <= nextOffset {
		pom.ResetOffset(offset, "")
	} else {
		pom.MarkOffset(offset, "")
	}

	if err := pom.Close(); err != nil {
		logger.Fatal("failed to reset offset",
			zap.String("topic", config.Kafka.Topic),
			zap.Int32("partition", partition),
			zap.Int64("offset", offset),
			zap.Error(err))
	}

	logger.Info("offset reset successfully",
		zap.String("topic", config.Kafka.Topic),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
	)
}

type PartitionOffset struct {
	Partition int32
	Offset    int64
}

func doOffsetGet(offsetManager sarama.OffsetManager, partitions []int32, logger *zap.Logger) {
	var (
		wg     sync.WaitGroup
		poList = make([]PartitionOffset, len(partitions))
	)

	for i, partition := range partitions {
		wg.Add(1)
		go func(partition int32, po *PartitionOffset) {
			defer wg.Done()
			pom, err := offsetManager.ManagePartition(config.Kafka.Topic, partition)
			if err != nil {
				logger.Fatal("failed to get partition offset manager",
					zap.String("topic", config.Kafka.Topic),
					zap.Int32("partition", partition),
				)
			}
			po.Partition = partition
			po.Offset, _ = pom.NextOffset()
			pom.Close()
		}(partition, &poList[i])
	}
	wg.Wait()

	sort.Slice(poList, func(i, j int) bool { return poList[i].Partition < poList[j].Partition })

	for i := range poList {
		logger.Info("partion next offset",
			zap.String("topic", config.Kafka.Topic),
			zap.Int32("partition", poList[i].Partition),
			zap.Int64("offset", poList[i].Offset),
		)
	}
}
