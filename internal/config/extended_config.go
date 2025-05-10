package config

// KafkaConfig 消息队列配置
type KafkaConfig struct {
	Brokers       []string `yaml:"brokers"`
	TopicPrefix   string   `yaml:"topic_prefix"`
	ConsumerGroup string   `yaml:"consumer_group"`
}

// EtcdConfig ETCD配置
type EtcdConfig struct {
	Endpoints        []string `yaml:"endpoints"`
	DialTimeout      int      `yaml:"dial_timeout"`
	OperationTimeout int      `yaml:"operation_timeout"`
}

// TracingConfig 分布式追踪配置
type TracingConfig struct {
	Enabled        bool    `yaml:"enabled"`
	JaegerEndpoint string  `yaml:"jaeger_endpoint"`
	ServiceName    string  `yaml:"service_name"`
	SamplingRate   float64 `yaml:"sampling_rate"`
}

// MetricsConfig 指标监控配置
type MetricsConfig struct {
	Enabled        bool `yaml:"enabled"`
	PrometheusPort int  `yaml:"prometheus_port"`
}

// LoggingConfig 高级日志配置
type LoggingConfig struct {
	OutputPaths      []string            `yaml:"output_paths"`
	ErrorOutputPaths []string            `yaml:"error_output_paths"`
	Elasticsearch    ElasticsearchConfig `yaml:"elasticsearch"`
}

// ElasticsearchConfig Elasticsearch配置
type ElasticsearchConfig struct {
	Enabled  bool   `yaml:"enabled"`
	URL      string `yaml:"url"`
	Index    string `yaml:"index"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
