package config

import (
	"time"

	"github.com/Shopify/sarama"
	"gopkg.in/ini.v1"
)

var Consumer = &ConsumerConfig{}

type ConsumerConfig struct {
	ConsumerGroup   string
	BalanceStrategy sarama.BalanceStrategy
	TPSLimit        int64
	MaxRetries      int
	LogFile         string
	LogSampler      LogSamplerConfig
}

func (conf *ConsumerConfig) SectionName() string {
	return "consumer"
}

func (conf *ConsumerConfig) Load(section *ini.Section) error {
	conf.MaxRetries = section.Key("max_retries").MustInt(10)
	conf.TPSLimit = section.Key("tps_limit").MustInt64(100000)
	conf.ConsumerGroup = section.Key("consumer_group").MustString("")
	conf.LogFile = section.Key("log_file").MustString("kafka.log")
	conf.LogSampler.Enabled = section.Key("log_sampler_enabled").MustBool(false)
	conf.LogSampler.Tick = section.Key("log_sampler_tick").MustDuration(time.Second)
	conf.LogSampler.First = section.Key("log_sampler_first").MustInt(100)
	conf.LogSampler.ThereAfter = section.Key("log_sampler_thereafter").MustInt(10000)

	balanceStrategy := section.Key("balance_startegy").MustString("sticky")
	switch balanceStrategy {
	case "roundrobin":
		conf.BalanceStrategy = sarama.BalanceStrategyRoundRobin
	case "range":
		conf.BalanceStrategy = sarama.BalanceStrategyRange
	default:
		conf.BalanceStrategy = sarama.BalanceStrategySticky
	}
	return nil
}
