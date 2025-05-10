package infrastructure

import (
	"context"
	"fmt"
	"log"
	"time"

	"distributedJob/internal/config"
	"distributedJob/internal/store"
	"distributedJob/internal/store/etcd"
	"distributedJob/internal/store/kafka"
	"distributedJob/internal/store/mysql"
	redisManager "distributedJob/internal/store/redis"
	"distributedJob/pkg/logger"
	"distributedJob/pkg/metrics"
	"distributedJob/pkg/tracing"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

// Infrastructure holds all external services
type Infrastructure struct {
	DB          store.RepositoryManager
	Redis       *redisManager.Manager
	Kafka       *kafka.Manager
	Etcd        *etcd.Manager
	Tracer      *tracing.Tracer
	Metrics     *metrics.Metrics
	Logger      *logger.Logger
	initialized bool
}

// New creates a new infrastructure instance
func New() *Infrastructure {
	return &Infrastructure{}
}

// Initialize initializes all infrastructure components
func (i *Infrastructure) Initialize(ctx context.Context, cfg *config.Config) error {
	if i.initialized {
		return errors.New("infrastructure already initialized")
	}
	// Initialize logger
	logger.Init(
		cfg.Log.Level,
		cfg.Log.Filename,
		cfg.Log.MaxSize,
		cfg.Log.MaxBackups,
		cfg.Log.MaxAge,
		cfg.Log.Compress,
	)
	i.Logger = logger.GetLogger()
	// Initialize database
	db, err := mysql.NewMySQLManager(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to initialize database")
	}
	i.DB = db
	// Initialize Redis
	redis, err := redisManager.NewManager(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to initialize Redis")
	}
	i.Redis = redis

	// Initialize Kafka if configured
	if len(cfg.Kafka.Brokers) > 0 {
		i.Kafka = kafka.NewManager(cfg.Kafka.Brokers)
		if err := i.Kafka.InitializeProducer(); err != nil {
			return errors.Wrap(err, "failed to initialize Kafka producer")
		}
		// Initialize consumer with default handler (will be replaced later)
		defaultHandler := func(msg *sarama.ConsumerMessage) error {
			return nil
		}

		if err := i.Kafka.InitializeConsumer(
			[]string{cfg.Kafka.TopicPrefix + "jobs"},
			cfg.Kafka.ConsumerGroup,
			defaultHandler,
		); err != nil {
			return errors.Wrap(err, "failed to initialize Kafka consumer")
		}
	}

	// Initialize etcd if configured
	if len(cfg.Etcd.Endpoints) > 0 {
		etcdManager, err := etcd.NewManager(etcd.Config{
			Endpoints:        cfg.Etcd.Endpoints,
			DialTimeout:      time.Duration(cfg.Etcd.DialTimeout) * time.Second,
			OperationTimeout: time.Duration(cfg.Etcd.OperationTimeout) * time.Second,
		})
		if err != nil {
			return errors.Wrap(err, "failed to initialize etcd")
		}
		i.Etcd = etcdManager
	}

	// Initialize tracer
	tracer, err := tracing.NewTracer(tracing.Config{
		ServiceName:    cfg.Tracing.ServiceName,
		JaegerEndpoint: cfg.Tracing.JaegerEndpoint,
		SamplingRate:   cfg.Tracing.SamplingRate,
		Enabled:        cfg.Tracing.Enabled,
	})
	if err != nil {
		return errors.Wrap(err, "failed to initialize tracer")
	}
	i.Tracer = tracer

	// Initialize metrics
	i.Metrics = metrics.NewMetrics(metrics.Config{
		Enabled: cfg.Metrics.Enabled,
		Port:    cfg.Metrics.PrometheusPort,
	})

	if err := i.Metrics.Start(); err != nil {
		return errors.Wrap(err, "failed to start metrics server")
	}

	i.initialized = true
	log.Println("Infrastructure initialized successfully")
	return nil
}

// Shutdown gracefully shuts down all infrastructure components
func (i *Infrastructure) Shutdown(ctx context.Context) error {
	if !i.initialized {
		return nil
	}

	var errs []error

	// Shutdown metrics server
	if i.Metrics != nil {
		if err := i.Metrics.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown metrics server: %w", err))
		}
	}

	// Shutdown tracer
	if i.Tracer != nil {
		if err := i.Tracer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown tracer: %w", err))
		}
	}

	// Close etcd connection
	if i.Etcd != nil {
		if err := i.Etcd.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close etcd connection: %w", err))
		}
	}

	// Close Kafka connection
	if i.Kafka != nil {
		if err := i.Kafka.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Kafka connection: %w", err))
		}
	}

	// Close Redis connection
	if i.Redis != nil {
		if err := i.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis connection: %w", err))
		}
	}
	// Close DB connection if it supports Close method
	if i.DB != nil {
		if closer, ok := i.DB.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close DB connection: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to shutdown infrastructure: %v", errs)
	}

	i.initialized = false
	log.Println("Infrastructure shutdown successfully")
	return nil
}
