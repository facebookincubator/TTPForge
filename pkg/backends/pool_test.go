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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAndGetConfigByName(t *testing.T) {
	pool := NewConnectionPool()
	cfg := &RemoteConfig{
		Host:     "example.com",
		Port:     22,
		Protocol: "ssh",
		User:     "root",
	}

	err := pool.Register("myconn", cfg)
	require.NoError(t, err)

	got := pool.GetConfigByName("myconn")
	require.NotNil(t, got)
	assert.Equal(t, "example.com", got.Host)
	assert.Equal(t, "root", got.User)
}

func TestRegisterDuplicateNameErrors(t *testing.T) {
	pool := NewConnectionPool()
	cfg := &RemoteConfig{Host: "example.com"}

	err := pool.Register("dup", cfg)
	require.NoError(t, err)

	err = pool.Register("dup", cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestGetByNameUnregisteredErrors(t *testing.T) {
	pool := NewConnectionPool()

	_, err := pool.GetByName("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no connection registered")
}

func TestGetConfigByNameReturnsNilForUnknown(t *testing.T) {
	pool := NewConnectionPool()
	assert.Nil(t, pool.GetConfigByName("unknown"))
}
