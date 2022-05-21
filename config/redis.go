package config

import (
	"strings"
	"time"

	"github.com/stn81/kate/rdb"
	"gopkg.in/ini.v1"
)

// Redis is the redis config instance
var Redis = &RedisConfig{Config: &rdb.Config{}}

// Config defines the redis config
type RedisConfig struct {
	*rdb.Config
	CacheTimeout   time.Duration
	CacheSizeLimit int
}

// SectionName implements the `Config.SectionName()` method
func (conf *RedisConfig) SectionName() string {
	return "redis"
}

// Load implements the `Config.Load()` method
func (conf *RedisConfig) Load(section *ini.Section) error {
	addrs := section.Key("addrs").MustString("127.0.0.1:6379")
	conf.Addrs = strings.Split(addrs, ",")
	conf.ClusterEnabled = section.Key("cluster_enabled").MustBool(true)
	conf.RouteMode = section.Key("route_mode").MustString("master_slave_random")
	conf.MaxRedirects = section.Key("max_redirects").MustInt(8)
	conf.MaxRetries = section.Key("max_retries").MustInt(0)
	conf.MinRetryBackoff = section.Key("min_retry_backoff").MustDuration(0)
	conf.MaxRetryBackoff = section.Key("max_retry_backoff").MustDuration(0)
	conf.ConnectTimeout = section.Key("connect_timeout").MustDuration(20 * time.Millisecond)
	conf.ReadTimeout = section.Key("read_timeout").MustDuration(20 * time.Millisecond)
	conf.WriteTimeout = section.Key("write_timeout").MustDuration(20 * time.Millisecond)
	conf.PoolSize = section.Key("pool_size").MustInt(100)
	conf.MinIdleConns = section.Key("min_idle_conns").MustInt(20)
	conf.MaxConnAge = section.Key("max_conn_age").MustDuration(0)
	conf.PoolTimeout = section.Key("pool_timeout").MustDuration(20 * time.Millisecond)
	conf.IdleTimeout = section.Key("idle_timeout").MustDuration(30 * time.Second)
	conf.IdleCheckFrequency = section.Key("idle_check_frequency").MustDuration(0)
	conf.CacheTimeout = section.Key("cache_timeout").MustDuration(time.Minute * 10)
	conf.CacheSizeLimit = section.Key("cache_size_limit").MustInt(1024 * 1024 * 1024)
	conf.Password = section.Key("password").MustString("")
	conf.DB = section.Key("db").MustInt(0)

	return nil
}
