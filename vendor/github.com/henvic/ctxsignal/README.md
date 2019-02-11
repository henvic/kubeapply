# ctxsignal

[![GoDoc](https://godoc.org/github.com/henvic/ctxsignal?status.svg)](https://godoc.org/github.com/henvic/ctxsignal) [![Build Status](https://travis-ci.org/henvic/ctxsignal.svg?branch=master)](https://travis-ci.org/henvic/ctxsignal) [![Coverage Status](https://coveralls.io/repos/henvic/ctxsignal/badge.svg)](https://coveralls.io/r/henvic/ctxsignal) [![codebeat badge](https://codebeat.co/badges/b1aeaafc-7c28-4230-9867-8f88f10404fe)](https://codebeat.co/projects/github-com-henvic-ctxsignal-master) [![Go Report Card](https://goreportcard.com/badge/github.com/henvic/ctxsignal)](https://goreportcard.com/report/github.com/henvic/ctxsignal)

Package ctxsignal can be used to create contexts cancelable by system signals.

## Example

Creating a context copy cancelable when intercepting a SIGINT, SIGTERM, or SIGHUP signal:

```go
ctx, cancel := ctxsignal.WithSignals(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
defer cancel()

<-ctx.Done()

fmt.Println("Received signal!")
```

You can check what type of signal was received with:

```go
sig, err := ctxsignal.Closed(ctx)

if err != nil {
        return err
}

fmt.Println(sig) // sig type is os.Signal
```

You can send a signal using `kill -SIGNAL PID`. Example: `kill -SIGHUP 170`.

On Unix-like systems you can read the manual about signals with

```bash
$ man signal
```

## Signals available on a typical Linux system

```bash
$ kill -l
 1) SIGHUP         2) SIGINT         3) SIGQUIT         4) SIGILL
 5) SIGTRAP        6) SIGABRT        7) SIGBUS          8) SIGFPE
 9) SIGKILL       10) SIGUSR1       11) SIGSEGV        12) SIGUSR2
13) SIGPIPE       14) SIGALRM       15) SIGTERM        16) SIGSTKFLT
17) SIGCHLD       18) SIGCONT       19) SIGSTOP        20) SIGTSTP
21) SIGTTIN       22) SIGTTOU       23) SIGURG         24) SIGXCPU
25) SIGXFSZ       26) SIGVTALRM     27) SIGPROF        28) SIGWINCH
29) SIGPOLL       30) SIGPWR        31) SIGSYS         32) SIGRTMIN
64) SIGRTMAX
```

**SIGKILL and SIGSTOP signals cannot be intercepted or handled.**

See the [docs](https://godoc.org/github.com/henvic/ctxsignal) for more examples and information.
