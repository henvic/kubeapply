package kubeapply

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

var flagsTests = []struct {
	name string

	in Flags

	wantKeys       []string
	wantTimeout    time.Duration
	wantTimeoutErr error
}{
	{
		"basic",
		Flags{},
		[]string{},
		0,
		nil,
	},
	{
		"unsorted",
		Flags{
			"b":   "",
			"a":   "",
			"-x":  "",
			"c":   "",
			"--z": "",
			"d":   "",
		},
		[]string{"--z", "-x", "a", "b", "c", "d"},
		0,
		nil,
	},
	{
		"unsorted with timeout",
		Flags{
			"b":   "",
			"a":   "none",
			"-x":  "abc",
			"c":   "",
			"--z": "",
			"d":   "def",

			"timeout": "3m",
		},
		[]string{"--z", "-x", "a", "b", "c", "d", "timeout"},
		3 * time.Minute,
		nil,
	},
	{
		"unsorted with invalid timeout",
		Flags{
			"b":   "",
			"a":   "none",
			"-x":  "abc",
			"c":   "",
			"--z": "true",
			"d":   "def",

			"timeout": "invalid",
		},
		[]string{"--z", "-x", "a", "b", "c", "d", "timeout"},
		0,
		errors.New("time: invalid duration invalid"),
	},
}

func TestFlags(t *testing.T) {
	for _, tt := range flagsTests {
		t.Run(tt.name, func(t *testing.T) {
			gotKeys := tt.in.Keys()

			if !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("Expected %v, got %v instead", tt.wantKeys, gotKeys)
			}

			gotTimeout, gotTimeoutErr := tt.in.Timeout()

			if gotTimeout != tt.wantTimeout || fmt.Sprint(gotTimeoutErr) != fmt.Sprint(tt.wantTimeoutErr) {
				t.Errorf("Expected Flags{%+v}.Timeout() = (%v, %v), got (%v, %v) instead",
					tt.in, tt.wantTimeout, tt.wantTimeoutErr, gotTimeout, gotTimeoutErr)
			}
		})
	}
}

var applyCommandTests = []struct {
	name string

	in *Apply

	wantExecutable string
	wantArgs       []string
}{
	{
		"basic",
		&Apply{},
		"kubectl",
		[]string{"apply", "--output=json"},
	},
	{
		"apply -f file.yaml",
		&Apply{
			Flags: Flags{
				"f": "file.yaml",
			},
		},
		"kubectl",
		[]string{"apply", "-f=file.yaml", "--output=json"},
	},
	{
		"apply forced",
		&Apply{
			Flags: Flags{
				"--force": "",
			},
		},
		"kubectl",
		[]string{"apply", "--force", "--output=json"},
	},
	{
		"apply forced with timeout",
		&Apply{
			Flags: Flags{
				"timeout": "1m",
				"--force": "",
				"-f":      "file.yaml",
			},
		},
		"kubectl",
		[]string{"apply", "--force", "-f=file.yaml", "--timeout=1m", "--output=json"},
	},
	{
		"apply forced with timeout in bad order",
		&Apply{
			Flags: Flags{
				"timeout=1m": "",
				"--force":    "",
				"-f":         "file.yaml",
			},
		},
		"kubectl",
		[]string{"apply", "--force", "-f=file.yaml", "--timeout=1m", "--output=json"},
	},
}

func TestApplyCommand(t *testing.T) {
	for _, tt := range applyCommandTests {
		t.Run(tt.name, func(t *testing.T) {
			gotExecutable, gotArgs := tt.in.Command()

			if gotExecutable != tt.wantExecutable || !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("Expected &Apply(%+v).Command() = (%v, %v), got (%v, %v) instead",
					tt.in, tt.wantExecutable, tt.wantArgs, gotExecutable, gotArgs)
			}
		})
	}
}

var applyRunTests = []struct {
	name string
	ctx  context.Context

	in *Apply

	executable string

	wantResponse Response
	wantErr      error
}{
	{
		"basic",
		context.Background(),
		&Apply{},
		"echo",
		Response{
			Stdout: Output("xyz123 --output=json\n"),
		},
		nil,
	},
	{
		"apply -f file.yaml",
		context.Background(),
		&Apply{
			Flags: Flags{
				"f": "file.yaml",
			},
		},
		"echo",
		Response{
			Stdout: Output("xyz123 -f=file.yaml --output=json\n"),
		},
		nil,
	},
	{
		"not found program",
		context.Background(),
		&Apply{
			Flags: Flags{
				"f": "file.yaml",
			},
		},
		"echo-not-found-12395234",
		Response{
			ExitCode: -1,
		},
		errors.New("exec: \"echo-not-found-12395234\": executable file not found in $PATH"),
	},
	{
		"error invoking program",
		context.Background(),
		&Apply{
			Flags: Flags{
				"-": "",
			},
		},
		"go",
		Response{
			Stderr:   "go xyz123: unknown command\nRun 'go help' for usage.\n",
			ExitCode: 2,
		},
		errors.New("exit status 2"),
	},
}

func TestRunApply(t *testing.T) {
	Command = "xyz123"
	defer func() {
		Command = "apply"
	}()

	mockBlacklist()
	defer restoreBlacklist()

	for _, tt := range applyRunTests {
		t.Run(tt.name, func(t *testing.T) {
			tt.in.executable = tt.executable // replace kubectl with something easier to test
			tt.in.dontSave = true

			var gotResponse, gotErr = tt.in.Run(tt.ctx)

			if gotResponse.ExitCode != tt.wantResponse.ExitCode ||
				gotResponse.Stderr != tt.wantResponse.Stderr ||
				gotResponse.Stdout != tt.wantResponse.Stdout ||
				fmt.Sprint(gotErr) != fmt.Sprint(tt.wantErr) {
				t.Errorf("Expected &Apply(%+v).Run(ctx) = (%+v, %+v), got (%+v, %+v) instead",
					tt.name, tt.wantResponse, tt.wantErr, gotResponse, gotErr)
			}
		})
	}
}

var restoredBlacklist = blacklist

func mockBlacklist() {
	blacklist = map[string]struct{}{}
}

func restoreBlacklist() {
	blacklist = restoredBlacklist
}
