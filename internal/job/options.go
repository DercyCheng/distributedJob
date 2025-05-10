package job

import (
	"distributedJob/internal/store/etcd"
	"distributedJob/internal/store/kafka"
	"distributedJob/pkg/metrics"
	"distributedJob/pkg/tracing"
)

// SchedulerOption defines an option for configuring the scheduler
type SchedulerOption func(*Scheduler)

// WithKafka configures the scheduler to use Kafka for job distribution
func WithKafka(manager *kafka.Manager) SchedulerOption {
	return func(s *Scheduler) {
		s.kafkaManager = manager
		s.useKafka = true
	}
}

// WithEtcd configures the scheduler to use etcd for distributed locking and coordination
func WithEtcd(manager *etcd.Manager) SchedulerOption {
	return func(s *Scheduler) {
		s.etcdManager = manager
		s.useEtcd = true
	}
}

// WithMetrics configures the scheduler to report metrics
func WithMetrics(metricsManager *metrics.Metrics) SchedulerOption {
	return func(s *Scheduler) {
		s.metrics = metricsManager
	}
}

// WithTracer configures the scheduler to use distributed tracing
func WithTracer(tracer *tracing.Tracer) SchedulerOption {
	return func(s *Scheduler) {
		s.tracer = tracer
	}
}
