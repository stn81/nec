package proxysrv

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	"github.com/juju/ratelimit"
	"github.com/stn81/kate/rdb"
	"github.com/stn81/kate/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stn81/nec/config"
	"github.com/stn81/nec/proto/proxy"
)

const MaxReqSize = 1024 * 1024 // 1M

var (
	errRateLimitReached = errors.New("ratelimit reached")
)

type proxyImpl struct {
	cmdInfoMap   map[string]*redis.CommandInfo
	client       sarama.SyncProducer
	tokenBucket  *ratelimit.Bucket
	logger       *zap.Logger
	accessLogger *zap.Logger
	total        prometheus.Counter
	succ         prometheus.Counter
	fail         prometheus.Counter
	processTime  prometheus.Histogram
}

func newProxyImpl(tokenBucket *ratelimit.Bucket, logger, accessLogger *zap.Logger) *proxyImpl {
	return &proxyImpl{
		tokenBucket:  tokenBucket,
		logger:       logger,
		accessLogger: accessLogger,
		total: promauto.NewCounter(prometheus.CounterOpts{
			Name: "req_processed_total",
			Help: "The total number of processed requests by proxy",
		}),
		succ: promauto.NewCounter(prometheus.CounterOpts{
			Name: "req_processed_succ",
			Help: "The succ number of processed requests by proxy",
		}),
		fail: promauto.NewCounter(prometheus.CounterOpts{
			Name: "req_processed_fail",
			Help: "The fail number of processed requests by proxy",
		}),
		processTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name: "req_process_time_ms",
			Help: "The process time of proxy request in ms",
		}),
	}
}

func (s *proxyImpl) Init() error {
	clientConf := sarama.NewConfig()
	clientConf.Producer.Retry.Max = config.Proxy.MaxRetries
	clientConf.Producer.RequiredAcks = sarama.WaitForAll
	clientConf.Producer.Return.Successes = true
	clientConf.Metadata.Full = true
	clientConf.Version = config.Kafka.Version
	clientConf.ClientID = config.Kafka.ClientID

	client, err := sarama.NewSyncProducer(config.Kafka.BrokerAddrs, clientConf)
	if err != nil {
		s.logger.Error("failed to create kafka producer client", zap.Error(err))
		return err
	}

	s.client = client

	rdb := rdb.Get()
	cmdInfoMap, err := rdb.Command().Result()
	if err != nil {
		s.logger.Error("failed to get redis command infos", zap.Error(err))
		return err
	}

	for cmd := range cmdInfoMap {
		if allowed := config.Proxy.Commands[cmd]; !allowed {
			delete(cmdInfoMap, cmd)
		}
	}

	s.cmdInfoMap = cmdInfoMap

	return nil
}

func (s *proxyImpl) Uninit() error {
	if s.client != nil {
		if err := s.client.Close(); err != nil {
			s.logger.Error("failed to close kafka producer client", zap.Error(err))
			return err
		}
	}

	return nil
}

func (s *proxyImpl) Do(ctx context.Context, req *proxy.Request) (resp *proxy.Response, err error) {
	s.total.Inc()

	resp, err = s.do(ctx, req)

	if err != nil {
		s.fail.Inc()
	} else {
		s.succ.Inc()
	}

	return resp, err
}

func (s *proxyImpl) do(ctx context.Context, req *proxy.Request) (resp *proxy.Response, err error) {
	begin := time.Now()

	cmd, firstKey, err := s.check(req)
	switch {
	case err == errRateLimitReached:
		return &proxy.Response{Errno: proxy.Error_RATELIMIT, Message: "ratelimit reached"}, nil
	case err != nil:
		s.logger.Error("proxy request check failed",
			zap.String("request", utils.ToJSON(req)),
			zap.Error(err))
		return nil, err
	}

	value, err := proto.Marshal(req)
	if err != nil {
		s.logger.Error("proxy marshal request to pb failed",
			zap.String("request", utils.ToJSON(req)),
			zap.Error(err),
		)
		return nil, err
	}

	if len(value) > MaxReqSize {
		s.logger.Error("proxy request size too large",
			zap.String("command", req.Cmd),
			zap.String("key", string(firstKey)),
			zap.Int("size", len(value)),
		)
		return &proxy.Response{Errno: proxy.Error_SIZE_TOO_LARGE, Message: "size too large"}, nil
	}

	message := &sarama.ProducerMessage{
		Topic: config.Kafka.Topic,
		Key:   sarama.ByteEncoder(firstKey),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := s.client.SendMessage(message)
	if err != nil {
		s.logger.Error("proxy send message to kafka failed",
			zap.String("command", cmd),
			zap.String("key", string(firstKey)),
			zap.Error(err),
		)
		return nil, err
	}

	elapsed := time.Since(begin).Milliseconds()
	s.accessLogger.Info("send request to kakfa success",
		zap.String("command", cmd),
		zap.String("key", string(firstKey)),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
		zap.Int64("elapsed_ms", elapsed),
	)

	s.processTime.Observe(float64(elapsed))

	return &proxy.Response{}, nil
}

func (s *proxyImpl) check(req *proxy.Request) (cmd string, firstKey []byte, err error) {
	if !s.tokenBucket.WaitMaxDuration(1, time.Millisecond*100) {
		return "", nil, errRateLimitReached
	}

	if len(req.Args) < 1 {
		return "", nil, status.Error(codes.InvalidArgument, "len(args) < 2")
	}

	cmd = strings.ToLower(req.Cmd)

	cmdInfo, ok := s.cmdInfoMap[cmd]
	if !ok {
		return "", nil, status.Error(codes.InvalidArgument, "redis command not supported")
	}

	arityOk := true
	if cmdInfo.Arity > 0 {
		if len(req.Args)+1 != int(cmdInfo.Arity) {
			arityOk = false
		}
	} else {
		if len(req.Args)+1 < -int(cmdInfo.Arity) {
			arityOk = false
		}
	}

	if !arityOk {
		return "", nil, status.Error(codes.InvalidArgument, "invalid redis command arity")
	}

	return cmd, req.Args[cmdInfo.FirstKeyPos-1], nil
}
