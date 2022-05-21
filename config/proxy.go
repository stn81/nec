package config

import (
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

var Proxy = &ProxyConfig{}

type ProxyConfig struct {
	Addr       string
	TPSLimit   int64
	MaxRetries int
	LogFile    string
	LogSampler LogSamplerConfig
	Commands   map[string]bool
}

func (conf *ProxyConfig) SectionName() string {
	return "proxy"
}

func (conf *ProxyConfig) Load(section *ini.Section) error {
	conf.Commands = make(map[string]bool)

	cmdList := section.Key("commands").MustString("setex")
	for _, cmd := range strings.Split(cmdList, ",") {
		cmd = strings.ToLower(cmd)
		conf.Commands[cmd] = true
	}

	conf.Addr = section.Key("addr").MustString(":9090")
	conf.TPSLimit = section.Key("tps_limit").MustInt64(500000)
	conf.MaxRetries = section.Key("max_retries").MustInt(3)
	conf.LogFile = section.Key("log_file").MustString("grpc.log")
	conf.LogSampler.Enabled = section.Key("log_sampler_enabled").MustBool(false)
	conf.LogSampler.Tick = section.Key("log_sampler_tick").MustDuration(time.Second)
	conf.LogSampler.First = section.Key("log_sampler_first").MustInt(100)
	conf.LogSampler.ThereAfter = section.Key("log_sampler_thereafter").MustInt(10000)
	return nil
}
