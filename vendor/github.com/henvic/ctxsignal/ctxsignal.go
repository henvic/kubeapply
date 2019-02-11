// Package ctxsignal can be used to create contexts cancelable by system signals.
//
// You can send a signal using kill -SIGNAL PID (e.g., kill -SIGHUP 170).
//
// kill -l gives you a list of signals available on your system.
// On Unix-like systems you can use "man signal" to learn about signals.
// SIGKILL and SIGSTOP signals cannot be intercepted or handled.
package ctxsignal

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// WithSignals returns a copy of the parent context cancelable by the given
// system signals. The signals are reset when the context's Done channel is
// closed.
func WithSignals(parent context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
	var v = &sig{}
	ctx, cancel := context.WithCancel(context.WithValue(parent, ctxSig{}, v))
	s := make(chan os.Signal, 1)

	// BUG(henvic): Be aware signal handling is vulnerable to race conditions.
	signal.Notify(s, signals...)
	go withSignalsHandler(ctx, cancel, s, v)

	return ctx, cancel
}

func withSignalsHandler(ctx context.Context, cancel context.CancelFunc, s chan os.Signal, v *sig) {
	select {
	case sig := <-s:
		v.m.Lock()
		v.t = sig
		v.m.Unlock()

		signal.Stop(s)
		cancel()
		return
	case <-ctx.Done():
		signal.Stop(s)
	}
}

type ctxSig struct{}

type sig struct {
	m sync.RWMutex
	t os.Signal
}

// WithTermination creates a context canceled on signals SIGINT or SIGTERM.
func WithTermination(ctx context.Context) (context.Context, context.CancelFunc) {
	return WithSignals(ctx, syscall.SIGINT, syscall.SIGTERM)
}

// Closed gets the signal that closed a context channel.
func Closed(ctx context.Context) (os.Signal, error) {
	if v := ctx.Value(ctxSig{}); v != nil {
		if sv, ok := v.(*sig); ok {
			sv.m.RLock()
			var t = sv.t
			sv.m.RUnlock()

			if t != nil {
				return t, nil
			}
		}
	}

	var s os.Signal
	return s, errors.New("context not closed by signal")
}
