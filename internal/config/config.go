package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/distributedJob/pkg/logger"
	"gopkg.in/yaml.v2"
)

// Config 应用配置结构体
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
	Auth     AuthConfig     `yaml:"auth"`
	Job      JobConfig      `yaml:"job"`
	Rpc      RpcConfig      `yaml:"rpc"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	ContextPath     string `yaml:"context_path"`
	ShutdownTimeout int    `yaml:"shutdown_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Schema   string `yaml:"schema"`
	MaxIdle  int    `yaml:"max_idle"`
	MaxConn  int    `yaml:"max_conn"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JwtSecret        string `yaml:"jwt_secret"`
	JwtExpireHours   int    `yaml:"jwt_expire_hours"`
	EnableEncryption bool   `yaml:"enable_encryption"`
}

// JobConfig 任务配置
type JobConfig struct {
	HttpWorkers  int `yaml:"http_workers"`
	GrpcWorkers  int `yaml:"grpc_workers"`
	QueueSize    int `yaml:"queue_size"`
	MaxRetry     int `yaml:"max_retry"`
	RetryBackoff int `yaml:"retry_backoff"`
}

// RpcConfig RPC服务器配置
type RpcConfig struct {
	Port               int `yaml:"port"`
	MaxConcurrentStreams int `yaml:"max_concurrent_streams"`
	KeepAliveTime      int `yaml:"keep_alive_time"`
	KeepAliveTimeout   int `yaml:"keep_alive_timeout"`
}

// LoadConfig 从文件加载配置
func LoadConfig(file string) (*Config, error) {
	// 读取配置文件
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析YAML配置
	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置默认值
	setDefaults(&config)

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	logger.Infof("Loaded configuration from %s", filepath.Base(file))
	return &config, nil
}

// setDefaults 设置配置默认值
func setDefaults(config *Config) {
	// 服务器默认设置
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.ContextPath == "" {
		config.Server.ContextPath = "/v1"
	}
	if config.Server.ShutdownTimeout == 0 {
		config.Server.ShutdownTimeout = 30
	}

	// 数据库默认设置
	if config.Database.MaxIdle == 0 {
		config.Database.MaxIdle = 10
	}
	if config.Database.MaxConn == 0 {
		config.Database.MaxConn = 50
	}

	// 日志默认设置
	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	if config.Log.MaxSize == 0 {
		config.Log.MaxSize = 100
	}
	if config.Log.MaxBackups == 0 {
		config.Log.MaxBackups = 10
	}
	if config.Log.MaxAge == 0 {
		config.Log.MaxAge = 30
	}

	// 认证默认设置
	if config.Auth.JwtExpireHours == 0 {
		config.Auth.JwtExpireHours = 24
	}

	// 任务默认设置
	if config.Job.HttpWorkers == 0 {
		config.Job.HttpWorkers = 10
	}
	if config.Job.GrpcWorkers == 0 {
		config.Job.GrpcWorkers = 10
	}
	if config.Job.QueueSize == 0 {
		config.Job.QueueSize = 100
	}
	if config.Job.MaxRetry == 0 {
		config.Job.MaxRetry = 3
	}
	if config.Job.RetryBackoff == 0 {
		config.Job.RetryBackoff = 5
	}
}

// validateConfig 验证配置有效性
func validateConfig(config *Config) error {
	// 验证数据库配置
	if config.Database.URL == "" {
		return fmt.Errorf("database URL is required")
	}
	if config.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}
	if config.Database.Schema == "" {
		return fmt.Errorf("database schema is required")
	}

	// 验证JWT配置
	if config.Auth.JwtSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	return nil
}
