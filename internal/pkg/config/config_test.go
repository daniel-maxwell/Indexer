package config

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Clear environment variables that might interfere.
	os.Clearenv()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check a few default values.
	if config.ServerPort != "8080" {
		t.Errorf("expected ServerPort to be '8080', got %s", config.ServerPort)
	}
	if config.QueueCapacity != 1000 {
		t.Errorf("expected QueueCapacity to be 1000, got %d", config.QueueCapacity)
	}
	if config.ElasticsearchURL != "http://localhost:9200/_bulk" {
		t.Errorf("expected ElasticsearchURL to be 'http://localhost:9200/_bulk', got %s", config.ElasticsearchURL)
	}
	if config.LogLevel != "info" {
		t.Errorf("expected LogLevel to be 'info', got %s", config.LogLevel)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables.
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("QUEUE_CAPACITY", "500")
	os.Setenv("LOG_LEVEL", "debug")
	// You can set additional variables here to test other fields.

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if config.ServerPort != "9090" {
		t.Errorf("expected ServerPort to be '9090', got %s", config.ServerPort)
	}
	if config.QueueCapacity != 500 {
		t.Errorf("expected QueueCapacity to be 500, got %d", config.QueueCapacity)
	}
	if config.LogLevel != "debug" {
		t.Errorf("expected LogLevel to be 'debug', got %s", config.LogLevel)
	}

	// Clean up environment variables after test.
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("QUEUE_CAPACITY")
	os.Unsetenv("LOG_LEVEL")
}
