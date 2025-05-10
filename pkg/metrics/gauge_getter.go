package metrics

// GetGaugeValue returns a gauge value by name and labels
func (m *Metrics) GetGaugeValue(name string, labels ...string) float64 {
	if !m.config.Enabled {
		return -1
	}
	_, exists := m.gauges[name]
	if !exists {
		return -1
	}

	// For simplicity, we assume empty labels or specific label values
	// In a real implementation, you would match the exact label values
	// We can't directly access gauge values from Prometheus client library
	// This is just a placeholder that would normally access a local cache or
	// query the Prometheus API to get the current value
	if len(labels) == 0 || (name == "active_workers" && len(labels) > 0) {
		// In an actual implementation, we would get the value here
	}

	// Since we can't directly get the value from a gauge, we'll return -1 as a signal
	// that the value isn't available
	// In a real implementation, you might want to expose the gauge value through metrics registry
	return -1
}

// GetActiveWorkers returns the number of active workers
func (m *Metrics) GetActiveWorkers() float64 {
	return m.GetGaugeValue("active_workers", "worker")
}
