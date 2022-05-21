package consumer

import (
	"context"
	"path"
	"sync"
	"time"

	"github.com/stn81/nec/config"
	"github.com/stn81/nec/proto/proxy"
	"github.com/stn81/kate/log"
	"github.com/stn81/kate/rdb"
	"github.com/stn81/retry"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"github.com/juju/ratelimit"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var gService *consumerService

type consumerService struct {
	conf         config.ConsumerConfig
	ready        chan bool
	client       sarama.ConsumerGroup
	redis        rdb.Client
	tokenBucket  *ratelimit.Bucket
	logger       *zap.Logger
	accessLogger *zap.Logger
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	total        prometheus.Counter
	succ         prometheus.Counter
	fail         prometheus.Counter
	processTime  prometheus.Histogram
}

func Start(logger *zap.Logger) {
	if gService != nil {
		panic("consumer start twice")
	}

	gService = &consumerService{
		conf:   *config.Consumer,
		ready:  make(chan bool),
		logger: logger.Named("consumer"),
		total: promauto.NewCounter(prometheus.CounterOpts{
			Name: "consumer_processed_total",
			Help: "The total number of processed messages by consumer",
		}),
		succ: promauto.NewCounter(prometheus.CounterOpts{
			Name: "consumer_processed_succ",
			Help: "The succ number of processed messages by consumer",
		}),
		fail: promauto.NewCounter(prometheus.CounterOpts{
			Name: "consumer_processed_fail",
			Help: "The fail number of processed messages by consumer",
		}),
		processTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name: "consumer_process_time_ms",
			Help: "The process time of consumer message in ms",
		}),
	}
	gService.start()
}

func Stop() {
	if gService != nil {
		gService.stop()
	}
}

func (s *consumerService) start() {
	loggerCfg := zap.NewProductionEncoderConfig()
	loggerCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	enc := zapcore.NewJSONEncoder(loggerCfg)
	core := zapcore.NewSampler(
		log.MustNewCoreWithLevelAbove(zapcore.InfoLevel, path.Join(config.Main.LogDir, s.conf.LogFile), enc),
		s.conf.LogSampler.Tick,
		s.conf.LogSampler.First,
		s.conf.LogSampler.ThereAfter,
	)

	opts := []zap.Option{
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCaller(),
	}

	s.tokenBucket = ratelimit.NewBucketWithRate(float64(s.conf.TPSLimit), s.conf.TPSLimit)

	s.accessLogger = zap.New(core, opts...)

	s.redis = rdb.Get()

	clientConf := sarama.NewConfig()
	clientConf.Version = config.Kafka.Version
	clientConf.ClientID = config.Kafka.ClientID
	clientConf.Consumer.Group.Rebalance.Strategy = s.conf.BalanceStrategy

	var err error
	s.client, err = sarama.NewConsumerGroup(
		config.Kafka.BrokerAddrs,
		s.conf.ConsumerGroup,
		clientConf,
	)
	if err != nil {
		s.logger.Fatal("failed to create consumer group", zap.Error(err))
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())

	gService.wg.Add(1)
	go gService.serve()
	<-gService.ready
}

func (s *consumerService) serve() {
	defer func() {
		s.wg.Done()
		s.logger.Info("consumer client stopped")
	}()

	s.logger.Info("consumer client started to serve")

	for {
		if err := s.client.Consume(s.ctx, []string{config.Kafka.Topic}, s); err != nil {
			s.logger.Fatal("failed to consume message",
				zap.String("topic", config.Kafka.Topic),
				zap.Error(err))
		}

		if s.ctx.Err() != nil {
			return
		}
		s.ready = make(chan bool)
	}
}

func (s *consumerService) stop() {
	s.cancel()
	s.wg.Wait()
	if err := s.client.Close(); err != nil {
		s.logger.Fatal("failed to close consumer client", zap.Error(err))
	}
}

func (s *consumerService) Setup(session sarama.ConsumerGroupSession) error {
	close(s.ready)
	return nil
}

func (s *consumerService) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (s *consumerService) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		begin := time.Now()

		s.total.Inc()

		logger := s.logger.With(
			zap.String("topic", msg.Topic),
			zap.Int32("partition", msg.Partition),
			zap.Int64("offset", msg.Offset),
			zap.String("key", string(msg.Key)),
		)

		req := &proxy.Request{}
		if err := proto.Unmarshal(msg.Value, req); err != nil {
			logger.Error("failed to parse request", zap.Error(err))
			session.MarkMessage(msg, "")
			s.fail.Inc()
			continue
		}

		if len(req.Args) < 1 {
			logger.Error("too few args")
			session.MarkMessage(msg, "")
			s.fail.Inc()
			continue
		}

		args := make([]interface{}, 0, len(req.Args)+1)
		args = append(args, req.Cmd)
		for i := range req.Args {
			args = append(args, req.Args[i])
		}

		s.tokenBucket.Wait(1)

		strategy := s.getRetryStrategy()
		success := retry.Do(s.ctx, strategy, func() bool {
			if _, err := s.redis.Do(args...).Result(); err != nil {
				logger.Error("failed to proxy redis command",
					zap.String("command", string(req.Cmd)),
					zap.Error(err),
					zap.Bool("will_retry", strategy.HasNext()),
				)
				return false
			}
			return true
		})

		if success {
			s.succ.Inc()
		} else {
			s.fail.Inc()
		}

		session.MarkMessage(msg, "")

		elapsed := time.Since(begin).Milliseconds()

		s.processTime.Observe(float64(elapsed))

		s.accessLogger.Info("message claimed",
			zap.String("topic", msg.Topic),
			zap.Int32("partition", msg.Partition),
			zap.Int64("offset", msg.Offset),
			zap.String("key", string(msg.Key)),
			zap.String("command", req.Cmd),
			zap.Time("timestamp", msg.Timestamp),
			zap.Bool("success", success),
			zap.Int64("wait_ms", begin.Sub(msg.Timestamp).Milliseconds()),
			zap.Int64("elapsed_ms", elapsed),
		)
	}

	return nil
}

func (s *consumerService) getRetryStrategy() retry.Strategy {
	return &retry.All{
		&retry.ExponentialBackoffStrategy{
			InitialDelay: time.Millisecond * 20,
			MaxDelay:     time.Millisecond * 500,
		},
		&retry.CountStrategy{
			Tries: s.conf.MaxRetries,
		},
	}
}
