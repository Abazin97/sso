package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env            string        `yaml:"env" env-default:"local"`
	StoragePath    string        `yaml:"storage_path" env-required:"true"`
	TokenTTL       time.Duration `yaml:"token_ttl" env-default:"1h"`
	GRPC           GRPCConfig    `yaml:"grpc"`
	App            AppConfig     `yaml:"app"`
	SMTP           SMTPConfig    `yaml:"smtp"`
	Email          EmailConfig   `yaml:"email"`
	Redis          RedisConfig   `yaml:"redis"`
	MigrationsPath string
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type AppConfig struct {
	AppID     int    `yaml:"id"`
	AppName   string `yaml:"name"`
	AppSecret string `env:"secret"`
}

type SMTPConfig struct {
	Host                   string `yaml:"host"`
	Port                   int    `yaml:"port"`
	From                   string `yaml:"from"`
	Pass                   string `env:"pass"`
	Email                  string `yaml:"email"`
	VerificationCodeLength int    `yaml:"verification_code_length" env-default:"6"`
}

type EmailConfig struct {
	Templates EmailTemplate
	Subjects  EmailTemplate
}

type EmailTemplate struct {
	VerificationCode string `yaml:"verification_email"`
	VerificationName string `yaml:"verification_info"`
}

type RedisConfig struct {
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	VerTokenTTL time.Duration `yaml:"token_ttl" env-default:"5m"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadByPath(configPath)
}

func MustLoadByPath(configPath string) *Config {
	//checks if file is exist
	_ = godotenv.Load(".env")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

// fetches config path from command line flag or env variable.
func fetchConfigPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}
	return res
}
