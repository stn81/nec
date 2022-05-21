package proxysrv

import (
	"net"
	"path"
	"sync"

	"github.com/stn81/kate/log"
	"github.com/cloudflare/tableflip"
	"github.com/juju/ratelimit"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	"github.com/stn81/nec/config"
	"github.com/stn81/nec/proto/proxy"
)

var gService *proxyService

type proxyService struct {
	conf         config.ProxyConfig
	upgrader     *tableflip.Upgrader
	listener     net.Listener
	server       *grpc.Server
	proxy        *proxyImpl
	wg           sync.WaitGroup
	logger       *zap.Logger
	accessLogger *zap.Logger
}

// Start start the grpc service
func Start(upgrader *tableflip.Upgrader, logger *zap.Logger) {
	if gService != nil {
		panic("grpcsrv start twice")
	}

	gService = &proxyService{
		conf:     *config.Proxy,
		upgrader: upgrader,
		logger:   logger.Named("grpcsrv"),
	}
	gService.start()
}

// Stop stop the grpc service
func Stop() {
	if gService != nil {
		gService.stop()
	}
}

func (s *proxyService) start() {
	loggerCfg := zap.NewProductionEncoderConfig()
	loggerCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	enc := zapcore.NewJSONEncoder(loggerCfg)
	core := log.MustNewCoreWithLevelAbove(zapcore.InfoLevel, path.Join(config.Main.LogDir, s.conf.LogFile), enc)

	if s.conf.LogSampler.Enabled {
		core = zapcore.NewSampler(
			core,
			s.conf.LogSampler.Tick,
			s.conf.LogSampler.First,
			s.conf.LogSampler.ThereAfter,
		)
	}

	opts := []zap.Option{
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCaller(),
	}

	s.accessLogger = zap.New(core, opts...)

	var err error
	if s.listener, err = s.upgrader.Listen("tcp", s.conf.Addr); err != nil {
		s.logger.Fatal("grpc listen failed",
			zap.String("addr", s.conf.Addr),
			zap.Error(err),
		)
	}

	tokenBucket := ratelimit.NewBucketWithRate(float64(s.conf.TPSLimit), s.conf.TPSLimit)

	s.proxy = newProxyImpl(tokenBucket, s.logger, s.accessLogger)
	if err = s.proxy.Init(); err != nil {
		s.logger.Fatal("proxysrv init failed", zap.Error(err))
	}

	s.server = grpc.NewServer()
	proxy.RegisterProxyServer(s.server, s.proxy)

	gService.wg.Add(1)
	go gService.serve()
}

func (s *proxyService) serve() {
	defer func() {
		s.wg.Done()
		s.logger.Info("grpc service stopped")
	}()

	s.logger.Info("grpc service started listening", zap.String("addr", s.conf.Addr))

	if err := s.server.Serve(s.listener); err != nil {
		s.logger.Fatal("failed to serve grpc service", zap.Error(err))
	}
}

func (s *proxyService) stop() {
	s.server.GracefulStop()
	s.wg.Wait()

	if s.proxy != nil {
		s.proxy.Uninit()
	}
}
