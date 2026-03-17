/*
Copyright © 2025-present, Meta Platforms, Inc. and affiliates
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package backends

import (
	"fmt"
	"sync"

	"github.com/facebookincubator/ttpforge/pkg/logging"
)

// ConnectionPool manages reusable ExecutionBackend instances keyed
// by their connection parameters so that multiple steps targeting the
// same remote host share a single connection.  Named connections
// (registered via `connect` steps) are also stored here.
type ConnectionPool struct {
	mu           sync.Mutex
	backends     map[string]ExecutionBackend
	namedConfigs map[string]*RemoteConfig
}

// NewConnectionPool creates an empty ConnectionPool.
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		backends:     make(map[string]ExecutionBackend),
		namedConfigs: make(map[string]*RemoteConfig),
	}
}

// Register stores a RemoteConfig under the given name so that
// subsequent steps can reference it with `remote: <name>`.
// Returns an error if a connection with that name already exists.
func (p *ConnectionPool) Register(name string, cfg *RemoteConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if name == "local" {
		return fmt.Errorf("connection name %q is reserved", name)
	}
	if _, exists := p.namedConfigs[name]; exists {
		return fmt.Errorf("connection name %q is already registered", name)
	}
	p.namedConfigs[name] = cfg
	return nil
}

// RegisterWithBackend stores a RemoteConfig and a pre-built
// ExecutionBackend under the given name. This is useful for testing
// where you want to inject a mock backend without establishing a
// real connection.
func (p *ConnectionPool) RegisterWithBackend(name string, cfg *RemoteConfig, backend ExecutionBackend) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if name == "local" {
		return fmt.Errorf("connection name %q is reserved", name)
	}
	if _, exists := p.namedConfigs[name]; exists {
		return fmt.Errorf("connection name %q is already registered", name)
	}
	p.namedConfigs[name] = cfg
	key := poolKey(cfg)
	p.backends[key] = backend
	return nil
}

// GetByName resolves a named connection to its ExecutionBackend,
// creating it if necessary.
func (p *ConnectionPool) GetByName(name string) (ExecutionBackend, error) {
	p.mu.Lock()
	cfg, ok := p.namedConfigs[name]
	p.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("no connection registered with name %q", name)
	}
	return p.GetOrCreate(cfg)
}

// GetConfigByName returns the RemoteConfig for a named connection,
// or nil if no such name is registered.
func (p *ConnectionPool) GetConfigByName(name string) *RemoteConfig {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.namedConfigs[name]
}

// poolKey generates a unique key for a RemoteConfig.
func poolKey(cfg *RemoteConfig) string {
	port := cfg.Port
	if port == 0 {
		port = 22
	}
	return fmt.Sprintf("%s:%s:%d:%s", cfg.Protocol, cfg.Host, port, cfg.User)
}

// GetOrCreate returns a cached backend for the given config, or creates
// a new one if none exists yet.
func (p *ConnectionPool) GetOrCreate(cfg *RemoteConfig) (ExecutionBackend, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := poolKey(cfg)
	if backend, ok := p.backends[key]; ok {
		return backend, nil
	}

	backend, err := newBackendFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend for %s: %w", key, err)
	}

	p.backends[key] = backend
	return backend, nil
}

// CloseAll closes all backends in the pool and clears the cache.
func (p *ConnectionPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, backend := range p.backends {
		if err := backend.Close(); err != nil {
			logging.L().Errorf("error closing backend %s: %v", key, err)
		}
	}
	p.backends = make(map[string]ExecutionBackend)
}

// newBackendFromConfig creates an ExecutionBackend from a RemoteConfig.
func newBackendFromConfig(cfg *RemoteConfig) (ExecutionBackend, error) {
	protocol := cfg.Protocol
	if protocol == "" {
		protocol = "ssh"
	}

	switch protocol {
	case "ssh":
		return NewSSHBackend(cfg)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
