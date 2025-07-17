package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port         string        `yaml:"port"`
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration `yaml:"write_timeout"`
	} `yaml:"server"`

	Limits struct {
		MaxConcurrentTasks int `yaml:"max_concurrent_tasks"`
		MaxFilesPerTask    int `yaml:"max_files_per_task"`
		MaxFileSizeMB      int `yaml:"max_file_size_mb"`
	} `yaml:"limits"`

	AllowedTypes []string `yaml:"allowed_types"`
}

func Load(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}