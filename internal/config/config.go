package config

import (
    "fmt"
    "github.com/spf13/viper"
)

// Holds all the configuration fields for the indexing service.
type Config struct {
    // Basic server settings
    ServerPort      string `mapstructure:"SERVER_PORT"`
    
    // Queue capacity
    QueueCapacity   int    `mapstructure:"QUEUE_CAPACITY"`

    // Elasticsearch
    ElasticsearchURL string `mapstructure:"ELASTICSEARCH_URL"`
    IndexName        string `mapstructure:"INDEX_NAME"`

    // Bulk Indexer
    BulkThreshold   int    `mapstructure:"BULK_THRESHOLD"`

    // Logging
    LogLevel        string `mapstructure:"LOG_LEVEL"`
}

// Initializes Viper and unmarshals config into our Config struct.
// It can read from environment variables, config files, etc.
func LoadConfig() (*Config, error) {
    viper.SetDefault("SERVER_PORT", "8080")
    viper.SetDefault("QUEUE_CAPACITY", 1000)
    viper.SetDefault("ELASTICSEARCH_URL", "http://localhost:9200/_bulk")
    viper.SetDefault("INDEX_NAME", "search_engine_index")
    viper.SetDefault("BULK_THRESHOLD", 3)
    viper.SetDefault("LOG_LEVEL", "info")

    // Read environment variables
    viper.AutomaticEnv()

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }

    return &config, nil
}
