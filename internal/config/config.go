package config

import (
    "fmt"
    "github.com/spf13/viper"
)

type Config struct {
    ServerPort       string `mapstructure:"SERVER_PORT"`
    QueueCapacity    int    `mapstructure:"QUEUE_CAPACITY"`
    ElasticsearchURL string `mapstructure:"ELASTICSEARCH_URL"`
    IndexName        string `mapstructure:"INDEX_NAME"`
    BulkThreshold    int    `mapstructure:"BULK_THRESHOLD"`
    FlushInterval    int    `mapstructure:"FLUSH_INTERVAL"`  // in seconds
    MaxRetries       int    `mapstructure:"MAX_RETRIES"`     // for bulk requests
    LogLevel         string `mapstructure:"LOG_LEVEL"`
}

func LoadConfig() (*Config, error) {
    viper.SetDefault("SERVER_PORT", "8080")
    viper.SetDefault("QUEUE_CAPACITY", 1000)
    viper.SetDefault("ELASTICSEARCH_URL", "http://localhost:9200/_bulk")
    viper.SetDefault("INDEX_NAME", "search_engine_index")
    viper.SetDefault("BULK_THRESHOLD", 3)
    viper.SetDefault("FLUSH_INTERVAL", 30) // 30 seconds
    viper.SetDefault("MAX_RETRIES", 3)
    viper.SetDefault("LOG_LEVEL", "info")

    viper.AutomaticEnv()

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    return &config, nil
}
