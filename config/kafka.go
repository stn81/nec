package config

import (
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	"gopkg.in/ini.v1"
)

var Kafka = &KafkaConfig{}

type KafkaConfig struct {
	Version     sarama.KafkaVersion
	ClientID    string
	BrokerAddrs []string
	Topic       string
}

func (conf *KafkaConfig) SectionName() string {
	return "kafka"
}

func (conf *KafkaConfig) Load(section *ini.Section) error {
	versionStr := section.Key("version").MustString("2.1.1")
	version, err := sarama.ParseKafkaVersion(versionStr)
	if err != nil {
		return fmt.Errorf("invalid kafka version: %v, error=%w", version, err)
	}
	conf.Version = version

	brokerAddrs := section.Key("broker_addrs").MustString("")
	conf.BrokerAddrs = strings.Split(brokerAddrs, ",")

	conf.ClientID = section.Key("client_id").MustString("cpc_redis_proxy")
	conf.Topic = section.Key("topic").MustString("")

	return nil
}
