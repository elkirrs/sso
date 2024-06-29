package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Host      string     `yaml:"host"`
	Env       string     `yaml:"env" env-default:"local"`
	GRPC      GRPCConfig `yaml:"grpc"`
	DB        DB         `yaml:"db"`
	HTTP      HTTPConfig `yaml:"http"`
	Token     Token      `yaml:"token"`
	AppConfig AppConfig  `yaml:"appConfig"`
	Metrics   Metrics    `yaml:"metrics"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type HTTPConfig struct {
	Port         int           `yaml:"port"`
	CORS         CORS          `yaml:"cors"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"10s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"10s"`
}

type CORS struct {
	AllowedMethods     []string `yaml:"allowed_methods"`
	AllowedOrigins     []string `yaml:"allowed_origins"`
	AllowCredentials   bool     `yaml:"allow_credentials"`
	AllowedHeaders     []string `yaml:"allowed_headers"`
	OptionsPassthrough bool     `yaml:"options_passthrough"`
	ExposedHeaders     []string `yaml:"exposed_headers"`
	Debug              bool     `yaml:"debug"`
}

type Token struct {
	TTL           time.Duration `yaml:"ttl" env-default:"1h"`
	Refresh       time.Duration `yaml:"refresh" env-default:"1d"`
	Secret        string        `yaml:"secret" env-default:"secret"`
	RefreshSecret string        `yaml:"secret_refresh" env-default:"refresh_secret"`
}
type DB struct {
	MigrationsPath string        `yaml:"migration_path" env-required:"true"`
	SQLITE         SQLITE        `yaml:"sqlite"`
	PGSQL          PGSQL         `yaml:"pgsql"`
	MaxAttempts    int           `yaml:"max_attempts" env-default:"3"`
	MaxDelay       time.Duration `yaml:"max_delay" env-default:"6s"`
}

type SQLITE struct {
	StoragePath     string `yaml:"storage_path" env-required:"true"`
	MigrationsTable string `yaml:"migration_table" env-default:"migrations"`
}

type PGSQL struct {
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Database string `yaml:"database" env-required:"true"`
	Port     string `yaml:"port" env-required:"true"`
	Host     string `yaml:"host" env-required:"true"`
	Attempts int    `yaml:"attempts" env-default:"5"`
	SSLMode  string `yaml:"ssl_mode"  env-default:"disable"`
}

type AppConfig struct {
	LogLevel string `yaml:"log_level" env-default:"trace"`
	LogJSON  bool   `yaml:"log_json" env-default:"false"`
}

type Metrics struct {
	Port int `yaml:"port"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}
	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	// --config="path/to/config.yaml"
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = "./config/config.yaml"
	}

	return res
}
