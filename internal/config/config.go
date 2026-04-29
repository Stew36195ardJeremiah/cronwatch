package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job represents a single cron job to monitor.
type Job struct {
	Name          string        `yaml:"name"`
	Schedule      string        `yaml:"schedule"`
	MaxDuration   time.Duration `yaml:"max_duration"`
	DriftThreshold time.Duration `yaml:"drift_threshold"`
	AlertEmail    string        `yaml:"alert_email"`
}

// Config holds the full cronwatch configuration.
type Config struct {
	LogLevel  string        `yaml:"log_level"`
	StateFile string        `yaml:"state_file"`
	SMTP      SMTPConfig    `yaml:"smtp"`
	Jobs      []Job         `yaml:"jobs"`
}

// SMTPConfig holds mail alert settings.
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

// Load reads and parses the YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Jobs) == 0 {
		return fmt.Errorf("no jobs defined")
	}
	for i, job := range c.Jobs {
		if job.Name == "" {
			return fmt.Errorf("job[%d]: name is required", i)
		}
		if job.Schedule == "" {
			return fmt.Errorf("job %q: schedule is required", job.Name)
		}
	}
	return nil
}
