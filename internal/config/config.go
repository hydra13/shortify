// Модуль конфигурации сервиса.
//
// Конфигурировать сервис можно при помощи флагов и переменных окружения, напимер:
//
//	PROFILER_ENABLED=true go run cmd/shortener/main.go -a :8080
package config

import (
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
	ServerAddr      string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	DatabaseDriver  string
	CurrentMode     Mode
	AuditFile       string
	AuditURL        string
	ProfilerEnabled bool
	HTTPSEnabled    bool
}

func NewConfig() *Config {
	return &Config{
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "./storage.json",
		DatabaseDSN:     "postgresql://postgres:postgres@localhost:5432/shortify_data?sslmode=disable",
		DatabaseDriver:  "pgx",
		CurrentMode:     ModeInMemory,
		AuditFile:       "",
		AuditURL:        "",
		ProfilerEnabled: false,
		HTTPSEnabled:    false,
	}
}

func (c *Config) ParseConfig() {
	flag.StringVar(&c.ServerAddr, "a", ":8080", "server address")
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
	flag.StringVar(&c.AuditFile, "audit-file", "", "audit file")
	flag.StringVar(&c.AuditURL, "audit-url", "", "audit url")
	flag.BoolVar(&c.ProfilerEnabled, "profiler", false, "run profiler")
	flag.BoolVar(&c.HTTPSEnabled, "s", false, "enable https")

	flag.Parse()

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
		}
	}
}
