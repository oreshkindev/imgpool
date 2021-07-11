package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config struct for app config
type Config struct {
	Server struct {
		// Local machine IP Address to bind the HTTP Server to
		Host string `yaml:"host"`

		// Local machine TCP Port to bind the HTTP Server to
		Port string `yaml:"port"`

		// Channel capacity limit
		Queue int `yaml:"queue"`

		// How much tasks workers can picked from the channel
		Workers int `yaml:"workers"`

		// Time to emulate highload process
		Duration int `yaml:"duration"`

		// Primary route uri
		Api string `yaml:"api"`

		// Life time of temporary file
		Timeout int `yaml:"timeout"`

		// Hash is the length of temporary link
		Hash int `yaml:"hash"`

		// Default temporary dirrectory
		Path string `yaml:"path"`
	} `yaml:"server"`
	Database struct {
		// Database owner. Use postgres by default
		User string `yaml:"user"`

		// Database owner password
		Pass string `yaml:"pass"`

		// Database IP Address to bind the HTTP Server to
		Host string `yaml:"host"`

		// Database TCP Port to bind the HTTP Server to
		Port string `yaml:"port"`

		// Database owner table
		Table string `yaml:"table"`
	} `yaml:"database"`
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	configFile, e := os.Open(configPath)
	if e != nil {
		return nil, e
	}

	defer configFile.Close()

	// Init new YAML decode
	decode := yaml.NewDecoder(configFile)

	// Start YAML decoding from file
	if e := decode.Decode(&config); e != nil {
		return nil, e
	}

	return config, nil
}
