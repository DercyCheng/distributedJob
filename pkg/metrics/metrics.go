package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config holds metrics configuration
type Config struct {
	Enabled bool
	Port    int
}

// Metrics is the wrapper around Prometheus metrics
type Metrics struct {
	registry   *prometheus.Registry
	config     Config
	server     *http.Server
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
	summaries  map[string]*prometheus.SummaryVec
}

// NewMetrics creates a new metrics handler
func NewMetrics(config Config) *Metrics {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		registry:   registry,
		config:     config,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		summaries:  make(map[string]*prometheus.SummaryVec),
	}

	// Register default metrics
	m.RegisterCounter("requests_total", "Total number of requests", []string{"method", "endpoint", "status"})
	m.RegisterHistogram("request_duration_seconds", "Request duration in seconds", []string{"method", "endpoint"}, prometheus.DefBuckets)
	m.RegisterGauge("active_workers", "Number of active workers", []string{"type"})
	m.RegisterGauge("job_queue_size", "Current size of the job queue", []string{"priority"})
	m.RegisterCounter("jobs_processed_total", "Total number of jobs processed", []string{"type", "status"})
	m.RegisterHistogram("job_execution_time_seconds", "Job execution time in seconds", []string{"type"}, prometheus.DefBuckets)

	return m
}

// RegisterCounter registers a counter metric
func (m *Metrics) RegisterCounter(name, help string, labels []string) {
	if !m.config.Enabled {
		return
	}

	counter := promauto.With(m.registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)

	m.counters[name] = counter
}

// RegisterGauge registers a gauge metric
func (m *Metrics) RegisterGauge(name, help string, labels []string) {
	if !m.config.Enabled {
		return
	}

	gauge := promauto.With(m.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)

	m.gauges[name] = gauge
}

// RegisterHistogram registers a histogram metric
func (m *Metrics) RegisterHistogram(name, help string, labels []string, buckets []float64) {
	if !m.config.Enabled {
		return
	}

	histogram := promauto.With(m.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)

	m.histograms[name] = histogram
}

// RegisterSummary registers a summary metric
func (m *Metrics) RegisterSummary(name, help string, labels []string, objectives map[float64]float64) {
	if !m.config.Enabled {
		return
	}

	summary := promauto.With(m.registry).NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		labels,
	)

	m.summaries[name] = summary
}

// IncrementCounter increments a counter metric
func (m *Metrics) IncrementCounter(name string, labels ...string) {
	if !m.config.Enabled {
		return
	}

	counter, exists := m.counters[name]
	if !exists {
		return
	}

	counter.WithLabelValues(labels...).Inc()
}

// AddToCounter adds a value to a counter metric
func (m *Metrics) AddToCounter(name string, value float64, labels ...string) {
	if !m.config.Enabled {
		return
	}

	counter, exists := m.counters[name]
	if !exists {
		return
	}

	counter.WithLabelValues(labels...).Add(value)
}

// SetGauge sets a gauge metric
func (m *Metrics) SetGauge(name string, value float64, labels ...string) {
	if !m.config.Enabled {
		return
	}

	gauge, exists := m.gauges[name]
	if !exists {
		return
	}

	gauge.WithLabelValues(labels...).Set(value)
}

// IncrementGauge increments a gauge metric
func (m *Metrics) IncrementGauge(name string, labels ...string) {
	if !m.config.Enabled {
		return
	}

	gauge, exists := m.gauges[name]
	if !exists {
		return
	}

	gauge.WithLabelValues(labels...).Inc()
}

// DecrementGauge decrements a gauge metric
func (m *Metrics) DecrementGauge(name string, labels ...string) {
	if !m.config.Enabled {
		return
	}

	gauge, exists := m.gauges[name]
	if !exists {
		return
	}

	gauge.WithLabelValues(labels...).Dec()
}

// Note: GetGaugeValue implementation moved to gauge_getter.go

// ObserveHistogram observes a value in a histogram metric
func (m *Metrics) ObserveHistogram(name string, value float64, labels ...string) {
	if !m.config.Enabled {
		return
	}

	histogram, exists := m.histograms[name]
	if !exists {
		return
	}

	histogram.WithLabelValues(labels...).Observe(value)
}

// ObserveSummary observes a value in a summary metric
func (m *Metrics) ObserveSummary(name string, value float64, labels ...string) {
	if !m.config.Enabled {
		return
	}

	summary, exists := m.summaries[name]
	if !exists {
		return
	}

	summary.WithLabelValues(labels...).Observe(value)
}

// MeasureRequestDuration measures request duration and records it in a histogram
func (m *Metrics) MeasureRequestDuration(method, endpoint string, start time.Time) {
	if !m.config.Enabled {
		return
	}

	duration := time.Since(start).Seconds()
	m.ObserveHistogram("request_duration_seconds", duration, method, endpoint)
}

// Start starts the metrics server
func (m *Metrics) Start() error {
	if !m.config.Enabled {
		return nil
	}

	handler := promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
	mux := http.NewServeMux()
	mux.Handle("/metrics", handler)

	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.config.Port),
		Handler: mux,
	}

	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start metrics server: %v", err)
		}
	}()

	return nil
}

// Stop stops the metrics server
func (m *Metrics) Stop(ctx context.Context) error {
	if !m.config.Enabled || m.server == nil {
		return nil
	}

	return m.server.Shutdown(ctx)
}
