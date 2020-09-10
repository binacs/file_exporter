package config

import (
	"reflect"
	"sync"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	cfg1, err := LoadFromFile("config_test.toml")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	cfg2 := testConfig()
	if configEqual(cfg1, cfg2) {
		t.Fatalf("toml marshal and unmarshal fail")
		return
	}
}

func testConfig() *Config {
	cfg := Config{
		file: "config_test.toml",
		ExporterConfig: &ExporterConfig{
			RootPath: "./",
			HttpPort: "8012",
			Gateway:  "http://127.0.0.1:9091",
		},
		rwmtx: &sync.RWMutex{},
	}
	return &cfg
}

func configEqual(cfg1, cfg2 *Config) bool {
	if cfg1.file != cfg2.file {
		return false
	}
	if !testExporterConfigEqual(cfg1.ExporterConfig, cfg2.ExporterConfig) {
		return false
	}
	return true
}

func testExporterConfigEqual(cfg1, cfg2 *ExporterConfig) bool {
	if cfg1.RootPath != cfg2.RootPath {
		return false
	}
	if cfg1.HttpPort != cfg2.HttpPort {
		return false
	}
	if cfg1.Gateway != cfg2.Gateway {
		return false
	}
	if !reflect.DeepEqual(cfg1.Dir_Keyword, cfg2.Dir_Keyword) {
		return false
	}
	return true
}

func TestConfigReload(t *testing.T) {
	oldConfig := testConfig()
	newConfig := testConfig()
	err := newConfig.Reload()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if configEqual(newConfig, oldConfig) {
		t.Fatalf("newconfig not reload")
	}
}
