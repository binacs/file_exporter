package config

import (
	"sync"

	"github.com/BurntSushi/toml"
)

type Config struct {
	file           string
	ExporterConfig *ExporterConfig `toml:"ExporterConfig"`
	rwmtx          *sync.RWMutex
}

type ExporterConfig struct {
	RootPath    string            `toml:"RootPath"`
	HttpPort    string            `toml:"HttpPort"`
	Gateway     string            `toml:"Gateway"`
	Jobname     string            `toml:"Jobname"`
	Dir_Keyword map[string]string `toml:"Dir_Keyword"`
}

// defaultConfig return a default Config
func defaultConfig() Config {
	ret := Config{
		ExporterConfig: &ExporterConfig{
			RootPath: "./",
			HttpPort: "8012",
			Jobname:  "file_exporter",
		},
		rwmtx: &sync.RWMutex{},
	}
	return ret
}

// LoadFromFile return the loaded config from filepath
func LoadFromFile(file string) (*Config, error) {
	cfg := defaultConfig()
	if _, err := toml.DecodeFile(file, &cfg); err != nil {
		return &cfg, err
	}
	cfg.file = file
	return &cfg, nil
}

// Reload reload config dynamic, not used for now
func (cf *Config) Reload() error {
	rwmtx := cf.rwmtx
	rwmtx.Lock()
	defer rwmtx.Unlock()
	file := cf.file
	newconfig, err := LoadFromFile(file)
	if err != nil {
		return err
	}
	cf.BeforeReloadNotify()
	*cf = *newconfig
	cf.file = file
	cf.rwmtx = rwmtx
	cf.AfterReloadNotify()
	return nil
}

func (cf *Config) BeforeReloadNotify() {
	// TODO
}

func (cf *Config) AfterReloadNotify() {
	// TODO
}
