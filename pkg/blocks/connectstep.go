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

package blocks

import (
	"fmt"

	"github.com/facebookincubator/ttpforge/pkg/backends"
	"github.com/facebookincubator/ttpforge/pkg/logging"
)

// ConnectStep establishes a named SSH connection that subsequent steps
// can reference via `remote: <connection_name>`.
type ConnectStep struct {
	actionDefaults `yaml:",inline"`
	Protocol       string `yaml:"protocol,omitempty"`
	Host           string `yaml:"host"`
	Port           int    `yaml:"port,omitempty"`
	User           string `yaml:"user,omitempty"`
	Auth           string `yaml:"auth,omitempty"`
	KeyFile        string `yaml:"key_file,omitempty"`
	Password       string `yaml:"password,omitempty"`
	PasswordEnv    string `yaml:"password_env,omitempty"`
	KnownHosts     string `yaml:"known_hosts,omitempty"`
	JumpHost       string `yaml:"jump_host,omitempty"`
	Shell          string `yaml:"shell,omitempty"`
	ConnectionName string `yaml:"connection_name"`
}

// NewConnectStep creates a new ConnectStep.
func NewConnectStep() *ConnectStep {
	return &ConnectStep{}
}

// IsNil returns true when the step has no connection_name set,
// meaning no connect block was actually provided.
func (s *ConnectStep) IsNil() bool {
	return s.ConnectionName == ""
}

// Validate checks that required fields are present.
func (s *ConnectStep) Validate(_ TTPExecutionContext) error {
	if s.Host == "" {
		return fmt.Errorf("connect step %q: host is required", s.ConnectionName)
	}
	if s.ConnectionName == "" {
		return fmt.Errorf("connect step: connection_name is required")
	}
	return nil
}

// Template replaces template variables in all string fields.
func (s *ConnectStep) Template(execCtx TTPExecutionContext) error {
	fields := []*string{
		&s.Protocol,
		&s.Host,
		&s.User,
		&s.Auth,
		&s.KeyFile,
		&s.Password,
		&s.PasswordEnv,
		&s.KnownHosts,
		&s.JumpHost,
		&s.Shell,
		&s.ConnectionName,
	}
	for _, f := range fields {
		if *f == "" || !execCtx.containsStepTemplating(*f) {
			continue
		}
		val, err := execCtx.templateStep(*f)
		if err != nil {
			return err
		}
		*f = val
	}
	return nil
}

// Execute registers the connection in the pool and eagerly connects
// to fail fast if the host is unreachable.
func (s *ConnectStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	if execCtx.ConnPool == nil {
		return nil, fmt.Errorf("connect step %q: connection pool is not initialized", s.ConnectionName)
	}

	cfg := &backends.RemoteConfig{
		Host:        s.Host,
		Port:        s.Port,
		Protocol:    s.Protocol,
		User:        s.User,
		Auth:        s.Auth,
		KeyFile:     s.KeyFile,
		Password:    s.Password,
		PasswordEnv: s.PasswordEnv,
		KnownHosts:  s.KnownHosts,
		JumpHost:    s.JumpHost,
		Shell:       s.Shell,
	}

	logging.L().Debugf("ConnectStep config: Host=%q Port=%d Protocol=%q User=%q Auth=%q KeyFile=%q ConnectionName=%q",
		cfg.Host, cfg.Port, cfg.Protocol, cfg.User, cfg.Auth, cfg.KeyFile, s.ConnectionName)

	if err := execCtx.ConnPool.Register(s.ConnectionName, cfg); err != nil {
		return nil, fmt.Errorf("connect step %q: %w", s.ConnectionName, err)
	}

	// Eagerly connect so we fail fast if the host is unreachable.
	if _, err := execCtx.ConnPool.GetOrCreate(cfg); err != nil {
		return nil, fmt.Errorf("connect step %q: failed to connect: %w", s.ConnectionName, err)
	}

	logging.L().Infof("Connection %q established to %s", s.ConnectionName, s.Host)
	return &ActResult{Stdout: "connected"}, nil
}

// CanBeUsedInCompositeAction returns false — connect steps must be
// top-level steps.
func (s *ConnectStep) CanBeUsedInCompositeAction() bool {
	return false
}
