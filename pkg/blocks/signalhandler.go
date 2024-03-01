/*
Copyright © 2023-present, Meta Platforms, Inc. and affiliates
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
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/facebookincubator/ttpforge/pkg/logging"
)

var signalHandlerInstalled bool
var signalHandlerLock = sync.Mutex{}
var shutdownChan chan bool

// SetupSignalHandler sets up SIGINT and SIGTERM handlers for graceful shutdown
func SetupSignalHandler() chan bool {
	// setup signal handling only once
	signalHandlerLock.Lock()
	if signalHandlerInstalled {
		signalHandlerLock.Unlock()
		return shutdownChan
	}
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	shutdownChan = make(chan bool, 1)
	signalHandlerInstalled = true
	signalHandlerLock.Unlock()

	go func() {
		var sig os.Signal
		var counter int
		for {
			sig = <-sigs
			logging.L().Infof("[%v] Received signal %v, shutting down now", counter, sig)
			shutdownChan <- true
			counter++
		}
	}()

	return shutdownChan
}
