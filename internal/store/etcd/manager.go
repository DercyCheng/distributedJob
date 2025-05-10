package etcd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// Manager provides etcd operations
type Manager struct {
	client *clientv3.Client
	config Config
	mu     sync.RWMutex
}

// Config holds etcd configuration
type Config struct {
	Endpoints        []string
	DialTimeout      time.Duration
	OperationTimeout time.Duration
}

// ServiceRegistry provides service registration and discovery
type ServiceRegistry struct {
	manager *Manager
	leaseID clientv3.LeaseID
	prefix  string
	ttl     int64
}

// Lock provides distributed locking
type Lock struct {
	session *concurrency.Session
	mutex   *concurrency.Mutex
	key     string
}

// NewManager creates a new etcd manager
func NewManager(config Config) (*Manager, error) {
	if len(config.Endpoints) == 0 {
		return nil, errors.New("at least one endpoint is required")
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: config.DialTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &Manager{
		client: client,
		config: config,
	}, nil
}

// Put stores a key-value pair with optional lease
func (m *Manager) Put(ctx context.Context, key, value string, leaseID ...clientv3.LeaseID) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, m.config.OperationTimeout)
	defer cancel()

	var opts []clientv3.OpOption
	if len(leaseID) > 0 {
		opts = append(opts, clientv3.WithLease(leaseID[0]))
	}

	_, err := m.client.Put(timeoutCtx, key, value, opts...)
	if err != nil {
		return fmt.Errorf("failed to put value in etcd: %w", err)
	}

	return nil
}

// Get retrieves a value by key
func (m *Manager) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, m.config.OperationTimeout)
	defer cancel()

	resp, err := m.client.Get(timeoutCtx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get value from etcd: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("key not found: %s", key)
	}

	return string(resp.Kvs[0].Value), nil
}

// Delete removes a key
func (m *Manager) Delete(ctx context.Context, key string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, m.config.OperationTimeout)
	defer cancel()

	_, err := m.client.Delete(timeoutCtx, key)
	if err != nil {
		return fmt.Errorf("failed to delete key from etcd: %w", err)
	}

	return nil
}

// GetPrefix retrieves all keys with a specific prefix
func (m *Manager) GetPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, m.config.OperationTimeout)
	defer cancel()

	resp, err := m.client.Get(timeoutCtx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get values with prefix from etcd: %w", err)
	}

	result := make(map[string]string, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		result[string(kv.Key)] = string(kv.Value)
	}

	return result, nil
}

// Watch watches for changes to a key or prefix
func (m *Manager) Watch(ctx context.Context, key string, isPrefix bool) clientv3.WatchChan {
	m.mu.RLock()
	defer m.mu.RUnlock()

	opts := []clientv3.OpOption{}
	if isPrefix {
		opts = append(opts, clientv3.WithPrefix())
	}

	return m.client.Watch(ctx, key, opts...)
}

// GrantLease creates a new lease
func (m *Manager) GrantLease(ctx context.Context, ttl int64) (clientv3.LeaseID, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, m.config.OperationTimeout)
	defer cancel()

	resp, err := m.client.Lease.Grant(timeoutCtx, ttl)
	if err != nil {
		return 0, fmt.Errorf("failed to grant lease: %w", err)
	}

	return resp.ID, nil
}

// KeepAlive keeps a lease alive until context is cancelled
func (m *Manager) KeepAlive(ctx context.Context, leaseID clientv3.LeaseID) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keepAliveCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	keepAliveChan, err := m.client.Lease.KeepAlive(keepAliveCtx, leaseID)
	if err != nil {
		return fmt.Errorf("failed to keep lease alive: %w", err)
	}

	go func() {
		for {
			select {
			case ka, ok := <-keepAliveChan:
				if !ok {
					log.Printf("Keep alive channel closed for lease %d", leaseID)
					return
				}
				log.Printf("Lease %d renewed with TTL %d", ka.ID, ka.TTL)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Close closes the etcd client
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.client.Close()
}

// NewServiceRegistry creates a new service registry
func (m *Manager) NewServiceRegistry(prefix string, ttl int64) (*ServiceRegistry, error) {
	return &ServiceRegistry{
		manager: m,
		prefix:  prefix,
		ttl:     ttl,
	}, nil
}

// Register registers a service
func (r *ServiceRegistry) Register(ctx context.Context, serviceName, serviceAddr string) error {
	leaseID, err := r.manager.GrantLease(ctx, r.ttl)
	if err != nil {
		return fmt.Errorf("failed to grant lease for service registration: %w", err)
	}

	r.leaseID = leaseID

	key := fmt.Sprintf("%s/%s/%s", r.prefix, serviceName, serviceAddr)
	if err := r.manager.Put(ctx, key, serviceAddr, leaseID); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	if err := r.manager.KeepAlive(ctx, leaseID); err != nil {
		return fmt.Errorf("failed to keep service registration alive: %w", err)
	}

	return nil
}

// Deregister deregisters a service
func (r *ServiceRegistry) Deregister(ctx context.Context, serviceName, serviceAddr string) error {
	key := fmt.Sprintf("%s/%s/%s", r.prefix, serviceName, serviceAddr)
	return r.manager.Delete(ctx, key)
}

// Discover discovers services
func (r *ServiceRegistry) Discover(ctx context.Context, serviceName string) ([]string, error) {
	prefix := fmt.Sprintf("%s/%s/", r.prefix, serviceName)
	values, err := r.manager.GetPrefix(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	services := make([]string, 0, len(values))
	for _, addr := range values {
		services = append(services, addr)
	}

	return services, nil
}

// NewLock creates a new distributed lock
func (m *Manager) NewLock(ctx context.Context, key string, ttl int) (*Lock, error) {
	session, err := concurrency.NewSession(m.client, concurrency.WithTTL(ttl))
	if err != nil {
		return nil, fmt.Errorf("failed to create lock session: %w", err)
	}

	mutex := concurrency.NewMutex(session, key)

	return &Lock{
		session: session,
		mutex:   mutex,
		key:     key,
	}, nil
}

// Lock acquires the lock
func (l *Lock) Lock(ctx context.Context) error {
	return l.mutex.Lock(ctx)
}

// Unlock releases the lock
func (l *Lock) Unlock(ctx context.Context) error {
	return l.mutex.Unlock(ctx)
}

// Close closes the lock session
func (l *Lock) Close() error {
	return l.session.Close()
}
