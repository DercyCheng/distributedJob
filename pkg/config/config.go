package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Logger    LoggerConfig    `mapstructure:"logger"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	AI        AIConfig        `mapstructure:"ai"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	HTTP HTTPConfig `mapstructure:"http"`
	GRPC GRPCConfig `mapstructure:"grpc"`
}

// HTTPConfig HTTP 服务配置
type HTTPConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// GRPCConfig gRPC 服务配置
type GRPCConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL MySQLConfig `mapstructure:"mysql"`
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	Charset         string        `mapstructure:"charset"`
	ParseTime       bool          `mapstructure:"parseTime"`
	Loc             string        `mapstructure:"loc"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"poolSize"`
	MinIdleConns int    `mapstructure:"minIdleConns"`
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	Timezone          string   `mapstructure:"timezone"`
	MaxWorkers        int      `mapstructure:"maxWorkers"`
	RetryAttempts     int      `mapstructure:"retryAttempts"`
	HeartbeatInterval int      `mapstructure:"heartbeatInterval"`
	AI                AIConfig `mapstructure:"ai"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
	Format string `mapstructure:"format"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret string        `mapstructure:"secret"`
	Expire time.Duration `mapstructure:"expire"`
}

// AIConfig AI 配置
type AIConfig struct {
	Enabled         bool    `mapstructure:"enabled"`
	DashScopeAPIKey string  `mapstructure:"dashscopeApiKey"`
	Model           string  `mapstructure:"model"`
	Temperature     float64 `mapstructure:"temperature"`
	MaxTokens       int     `mapstructure:"maxTokens"`
}

var globalConfig *Config

// Load 加载配置
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	// 支持环境变量
	viper.AutomaticEnv()

	// 显式绑定环境变量
	viper.BindEnv("scheduler.ai.dashscopeApiKey", "DASHSCOPE_API_KEY")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	globalConfig = &config
	return &config, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// setDefaults 设置默认值
func setDefaults() {
	// 服务器默认值
	viper.SetDefault("server.http.port", "8080")
	viper.SetDefault("server.http.host", "0.0.0.0")
	viper.SetDefault("server.grpc.port", "9090")
	viper.SetDefault("server.grpc.host", "0.0.0.0")

	// 数据库默认值
	viper.SetDefault("database.mysql.host", "localhost")
	viper.SetDefault("database.mysql.port", 3306)
	viper.SetDefault("database.mysql.username", "root")
	viper.SetDefault("database.mysql.database", "go_job")
	viper.SetDefault("database.mysql.charset", "utf8mb4")
	viper.SetDefault("database.mysql.parseTime", true)
	viper.SetDefault("database.mysql.loc", "Local")
	viper.SetDefault("database.mysql.maxIdleConns", 10)
	viper.SetDefault("database.mysql.maxOpenConns", 100)
	viper.SetDefault("database.mysql.connMaxLifetime", time.Minute*5)

	// Redis 默认值
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.poolSize", 10)
	viper.SetDefault("redis.minIdleConns", 2)

	// 调度器默认值
	viper.SetDefault("scheduler.timezone", "Asia/Shanghai")
	viper.SetDefault("scheduler.maxWorkers", 100)
	viper.SetDefault("scheduler.retryAttempts", 3)
	viper.SetDefault("scheduler.heartbeatInterval", 30)

	// AI 调度器默认值
	viper.SetDefault("scheduler.ai.enabled", true)
	viper.SetDefault("scheduler.ai.dashscopeApiKey", "")
	viper.SetDefault("scheduler.ai.model", "qwen-max")
	viper.SetDefault("scheduler.ai.temperature", 0.7)
	viper.SetDefault("scheduler.ai.maxTokens", 2000)

	// 日志默认值
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.output", "stdout")
	viper.SetDefault("logger.format", "json")

	// JWT 默认值
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expire", time.Hour*24)
}

// GetHTTPAddr 获取 HTTP 地址
func (c *Config) GetHTTPAddr() string {
	return fmt.Sprintf("%s:%s", c.Server.HTTP.Host, c.Server.HTTP.Port)
}

// GetGRPCAddr 获取 gRPC 地址
func (c *Config) GetGRPCAddr() string {
	return fmt.Sprintf("%s:%s", c.Server.GRPC.Host, c.Server.GRPC.Port)
}

// GetMySQLDSN 获取 MySQL DSN
func (c *Config) GetMySQLDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Database.MySQL.Username,
		c.Database.MySQL.Password,
		c.Database.MySQL.Host,
		c.Database.MySQL.Port,
		c.Database.MySQL.Database,
		c.Database.MySQL.Charset,
		c.Database.MySQL.ParseTime,
		c.Database.MySQL.Loc,
	)
}

// GetRedisAddr 获取 Redis 地址
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}
