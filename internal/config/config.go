// Модуль конфигурации сервиса.
//
// Конфигурировать сервис можно при помощи флагов и переменных окружения, напимер:
//
//	PROFILER_ENABLED=true go run cmd/shortener/main.go -a :8080
package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
)

type Mode int

const (
	ModeInMemory Mode = iota + 1
	ModeFileStorage
	ModeDBStorage
)

type Config struct {
	ServerAddr      string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	DatabaseDriver  string
	CurrentMode     Mode
	AuditFile       string
	AuditURL        string
	ProfilerEnabled bool
	HTTPSEnabled    bool   `json:"enable_https"`
	CertFile        string `json:"cert_file"`
	KeyFile         string `json:"key_file"`
}

func NewConfig() *Config {
	conf := &Config{
		ServerAddr:      ":8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "./storage.json",
		DatabaseDSN:     "postgresql://postgres:postgres@localhost:5432/shortify_data?sslmode=disable",
		DatabaseDriver:  "pgx",
		CurrentMode:     ModeInMemory,
	}

	conf.ParseConfig()

	return conf
}

func (c *Config) ParseConfig() {
	configFile := c.getConfigFile()
	if configFile != "" {
		c.parseConfigFile(configFile)
	}
	c.ParseFlags()
	c.ParseEnv()
}

func (c *Config) getConfigFile() string {
	configFile := ""

	for i, arg := range os.Args {
		if arg == "-c" || arg == "-config" && len(os.Args) > i+1 {
			configFile = os.Args[i+1]
			break
		}
	}

	if envValue, ok := os.LookupEnv("CONFIG"); ok {
		configFile = envValue
	}

	return configFile
}

func (c *Config) parseConfigFile(configFile string) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return
	}

	json.Unmarshal(data, c)
}

func (c *Config) ParseFlags() {
	flag.StringVar(&c.ServerAddr, "a", c.ServerAddr, "server address")
	flag.Func("f", "file storage path (default: \"./storage.json\")", func(path string) error {
		if len(path) == 0 {
			return nil
		}

		c.FileStoragePath = path

		if c.CurrentMode <= ModeFileStorage {
			c.CurrentMode = ModeFileStorage
		}

		return nil
	})
	flag.Func("b", "result base url (default: \"http://localhost:8080\")", func(url string) error {
		if len(url) == 0 {
			return nil
		}

		if url[len(url)-1] == '/' {
			url = url[:len(url)-1]
		}

		c.BaseURL = url

		return nil
	})
	flag.Func("d", "database DSN (default: \"postgresql://postgres:postgres@localhost:5432/shortify_data?sslmode=disable\")", func(dsn string) error {
		if len(dsn) == 0 {
			return nil
		}

		c.DatabaseDSN = dsn

		if c.CurrentMode <= ModeDBStorage {
			c.CurrentMode = ModeDBStorage
		}

		return nil
	})
	flag.StringVar(&c.AuditFile, "audit-file", c.AuditFile, "audit file")
	flag.StringVar(&c.AuditURL, "audit-url", c.AuditURL, "audit url")
	flag.BoolVar(&c.ProfilerEnabled, "profiler", c.ProfilerEnabled, "run profiler")
	flag.BoolVar(&c.HTTPSEnabled, "s", c.HTTPSEnabled, "enable https")
	flag.StringVar(&c.CertFile, "cert-file", c.CertFile, "cert file for https")
	flag.StringVar(&c.KeyFile, "key-file", c.KeyFile, "key file for https")

	flag.Parse()
}

func (c *Config) ParseEnv() {
	if addr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		c.ServerAddr = addr
	}

	if url, ok := os.LookupEnv("BASE_URL"); ok {
		c.BaseURL = url
	}

	if filePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		c.FileStoragePath = filePath

		if c.CurrentMode <= ModeFileStorage {
			c.CurrentMode = ModeFileStorage
		}
	}

	if dsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		c.DatabaseDSN = dsn

		if c.CurrentMode <= ModeDBStorage {
			c.CurrentMode = ModeDBStorage
		}
	}

	if auditFile, ok := os.LookupEnv("AUDIT_FILE"); ok {
		c.AuditFile = auditFile
	}

	if auditURL, ok := os.LookupEnv("AUDIT_URL"); ok {
		c.AuditURL = auditURL
	}

	if profilerEnabledStr, ok := os.LookupEnv("PROFILER"); ok {
		if profilerEnabled, err := strconv.ParseBool(profilerEnabledStr); err == nil {
			c.ProfilerEnabled = profilerEnabled
		}
	}

	if httpsEnabledStr, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		if httpsEnabled, err := strconv.ParseBool(httpsEnabledStr); err == nil {
			c.HTTPSEnabled = httpsEnabled

			if certFile, ok := os.LookupEnv("CERT_FILE"); ok {
				c.CertFile = certFile
			}

			if keyFile, ok := os.LookupEnv("KEY_FILE"); ok {
				c.KeyFile = keyFile
			}
		}
	}
}
