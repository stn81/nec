package cmd

import "go.uber.org/zap"

type globalFlags struct {
	Debug      bool
	ConfigFile string
}

var (
	GlobalFlags = &globalFlags{}
)

func initStdLogger() (logger *zap.Logger, err error) {
	opts := []zap.Option{
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCaller(),
	}
	return zap.NewDevelopment(opts...)
}
