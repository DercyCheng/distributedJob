package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"distributedJob/pkg/logger"

	"gopkg.in/yaml.v2"
)

// Config 应用配置结构体
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Log      LogConfig      `yaml:"log"`
	Auth     AuthConfig     `yaml:"auth"`
	Job      JobConfig      `yaml:"job"`
	Rpc      RpcConfig      `yaml:"rpc"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Etcd     EtcdConfig     `yaml:"etcd"`
	Tracing  TracingConfig  `yaml:"tracing"`
	Metrics  MetricsConfig  `yaml:"metrics"`
	Logging  LoggingConfig  `yaml:"logging"`
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

// RedisConfig Redis配置
type RedisConfig struct {
	URL            string `yaml:"url"`
	Password       string `yaml:"password"`
	DB             int    `yaml:"db"`
	MaxIdle        int    `yaml:"max_idle"`
	MaxActive      int    `yaml:"max_active"`
	IdleTimeout    int    `yaml:"idle_timeout"`
	ConnectTimeout int    `yaml:"connect_timeout"`
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
	JwtSecret               string `yaml:"jwt_secret"`                // 访问令牌密钥
	JwtRefreshSecret        string `yaml:"jwt_refresh_secret"`        // 刷新令牌密钥
	JwtExpireMinutes        int    `yaml:"jwt_expire_minutes"`        // 访问令牌过期时间(分钟)
	JwtRefreshExpireDays    int    `yaml:"jwt_refresh_expire_days"`   // 刷新令牌过期时间(天)
	EnableEncryption        bool   `yaml:"enable_encryption"`         // 是否启用加密
	TokenRevocationStrategy string `yaml:"token_revocation_strategy"` // 令牌撤销策略: memory 或 redis
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
	Port                 int `yaml:"port"`
	MaxConcurrentStreams int `yaml:"max_concurrent_streams"`
	KeepAliveTime        int `yaml:"keep_alive_time"`
	KeepAliveTimeout     int `yaml:"keep_alive_timeout"`
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

	// Redis默认设置
	if config.Redis.URL == "" {
		config.Redis.URL = "localhost:6379"
	}
	if config.Redis.MaxIdle == 0 {
		config.Redis.MaxIdle = 10
	}
	if config.Redis.MaxActive == 0 {
		config.Redis.MaxActive = 100
	}
	if config.Redis.IdleTimeout == 0 {
		config.Redis.IdleTimeout = 300
	}
	if config.Redis.ConnectTimeout == 0 {
		config.Redis.ConnectTimeout = 5
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
	if config.Auth.JwtRefreshSecret == "" {
		// 如果未提供刷新令牌密钥，使用访问令牌密钥作为默认值，但在日志中警告
		config.Auth.JwtRefreshSecret = config.Auth.JwtSecret
		logger.Warn("No JWT refresh secret provided, using the same secret for both tokens is not recommended for production")
	}
	if config.Auth.JwtExpireMinutes == 0 {
		config.Auth.JwtExpireMinutes = 30 // 默认30分钟
	}
	if config.Auth.JwtRefreshExpireDays == 0 {
		config.Auth.JwtRefreshExpireDays = 7 // 默认7天
	}
	if config.Auth.TokenRevocationStrategy == "" {
		config.Auth.TokenRevocationStrategy = "memory" // 默认使用内存存储
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

	// RPC服务器默认设置
	if config.Rpc.Port == 0 {
		config.Rpc.Port = 9090
	}
	if config.Rpc.MaxConcurrentStreams == 0 {
		config.Rpc.MaxConcurrentStreams = 100
	}
	if config.Rpc.KeepAliveTime == 0 {
		config.Rpc.KeepAliveTime = 60 // 60秒
	}
	if config.Rpc.KeepAliveTimeout == 0 {
		config.Rpc.KeepAliveTimeout = 10 // 10秒
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
	if config.Auth.JwtRefreshSecret == "" {
		return fmt.Errorf("JWT refresh secret is required")
	}

	// 验证Token撤销策略
	if config.Auth.TokenRevocationStrategy != "memory" && config.Auth.TokenRevocationStrategy != "redis" {
		return fmt.Errorf("invalid token revocation strategy: %s, must be 'memory' or 'redis'", config.Auth.TokenRevocationStrategy)
	}

	return nil
}
