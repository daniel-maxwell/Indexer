package config

import (
    "fmt"
    "github.com/spf13/viper"
)

type Config struct {
    ServerPort       string `mapstructure:"SERVER_PORT"`
    QueueCapacity    int    `mapstructure:"QUEUE_CAPACITY"`
    NumWorkers       int    `mapstructure:"NUM_WORKERS"`

    // Existing fields remain unchanged
    ElasticsearchURL string `mapstructure:"ELASTICSEARCH_URL"`
    IndexName        string `mapstructure:"INDEX_NAME"`
    BulkThreshold    int    `mapstructure:"BULK_THRESHOLD"`
    FlushInterval    int    `mapstructure:"FLUSH_INTERVAL"`
    MaxRetries       int    `mapstructure:"MAX_RETRIES"`
    
    // Redis config
    RedisHost     string `mapstructure:"REDIS_HOST"`
    RedisPort     string `mapstructure:"REDIS_PORT"`
    RedisPassword string `mapstructure:"REDIS_PASSWORD"`
    RedisDB       int    `mapstructure:"REDIS_DB"`

    // NLP service config
    NlpServiceURL string `mapstructure:"NLP_SERVICE_URL"`
    
    LogLevel string `mapstructure:"LOG_LEVEL"`
}

func LoadConfig() (*Config, error) {
    // Set defaults for configuration values
    viper.SetDefault("SERVER_PORT", "8080")
    viper.SetDefault("QUEUE_CAPACITY", 1000)
    viper.SetDefault("NUM_WORKERS", 4) // Default to 4 workers
    viper.SetDefault("ELASTICSEARCH_URL", "http://localhost:9200/_bulk")
    viper.SetDefault("INDEX_NAME", "search_engine_index")
    viper.SetDefault("BULK_THRESHOLD", 3)
    viper.SetDefault("FLUSH_INTERVAL", 30)
    viper.SetDefault("MAX_RETRIES", 3)

    // Redis defaults
    viper.SetDefault("REDIS_HOST", "localhost")
    viper.SetDefault("REDIS_PORT", "6379")
    viper.SetDefault("REDIS_PASSWORD", "")
    viper.SetDefault("REDIS_DB", 0)
    viper.SetDefault("LOG_LEVEL", "info")

    // NLP service defaults
    viper.SetDefault("NLP_SERVICE_URL", "http://localhost:5000/nlp")

    viper.AutomaticEnv()

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    return &config, nil
}