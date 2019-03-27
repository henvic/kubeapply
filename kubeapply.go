package kubeapply

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Executable to call.
const Executable = "kubectl"

// Command is the first argument of "kubectl".
var Command = "apply"

// configurations directory
const configurations = "configurations"

const fileMode = os.FileMode(0644)
const dirFileMode = os.FileMode(0755)

// Flags for kubectl.
type Flags map[string]string

// Keys of the used flags.
func (f Flags) Keys() []string {
	var keys = []string{}

	for k := range f {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

// Timeout for the task.
func (f Flags) Timeout() (time.Duration, error) {
	timeout, ok := f["timeout"]

	if !ok {
		return 0, nil
	}

	return time.ParseDuration(timeout)
}

// Apply command.
type Apply struct {
	Flags Flags
	Files map[string][]byte

	IP string

	RequestDump []byte

	name string
	args []string

	executable string
	Subcommand string

	id        string
	timestamp time.Time
	dir       string

	dontSave bool // useful for disabling creating configuration directories

	m sync.RWMutex
}

func (a *Apply) listFiles() []string {
	var list = []string{}

	for f := range a.Files {
		list = append(list, f)
	}

	sort.Strings(list)
	return list
}

// Command prepared to be invoked by Run.
func (a *Apply) Command() (string, []string) {
	a.m.Lock()
	defer a.m.Unlock()
	return a.unsafeCommand()
}

func (a *Apply) unsafeCommand() (string, []string) {
	if a.executable == "" {
		a.executable = Executable
	}

	if a.Subcommand == "" {
		a.Subcommand = Command
	}

	if a.Flags == nil {
		a.Flags = Flags{}
	}

	args := []string{a.Subcommand}

	var flags = a.Flags

	var filenameFlag bool
	var outputFlag bool

	for _, f := range flags.Keys() {
		af := addFlag(f)

		switch af {
		case "-af", "--filename":
			filenameFlag = true
		case "-o", "--output":
			outputFlag = true
		}

		switch v := flags[f]; {
		case v == "":
			args = append(args, af)
		default:
			args = append(args, fmt.Sprintf("%s=%s", af, v))
		}
	}

	if !filenameFlag {
		args = append(args, a.addFilenameFlag()...)
	}

	if !outputFlag {
		args = append(args, "--output=json")
	}

	return a.executable, args
}

func (a *Apply) addFilenameFlag() []string {
	for filename := range a.Files {
		// expect Kubernetes configuration objects
		if strings.HasSuffix(filename, ".json") ||
			strings.HasSuffix(filename, ".yaml") ||
			strings.HasSuffix(filename, ".yml") {
			return []string{"--filename=./", "--recursive"}
		}
	}

	return nil
}

func addFlag(f string) string {
	if strings.HasPrefix(f, "-") {
		return f
	}

	if len(f) == 1 {
		return "-" + f
	}

	return "--" + f
}

// Response for the apply command.
type Response struct {
	ID string `json:"id,omitempty"`

	Command string   `json:"cmd"`
	Args    []string `json:"args"`
	CmdLine string   `json:"cmdline"`

	Stderr string `json:"stderr"`
	Stdout Output `json:"stdout,omitempty"`

	ExitCode int `json:"exit_code"`

	Dir string `json:"dir,omitempty"`
}

func (r *Response) embedError(err error) {
	// skip embeding error messages if exit code != 0
	if _, ok := err.(*exec.ExitError); ok {
		return
	}

	switch {
	case err == nil:
		return
	case r.Stderr == "":
		r.Stderr = err.Error()
	default:
		r.Stderr += fmt.Sprintf("\nkubeapply exec.Command error: %v", err)
	}
}

// Run command.
func (a *Apply) Run(ctx context.Context) (Response, error) {
	a.init()

	if err := a.maybeConfigure(); err != nil {
		return Response{
			Stderr:   err.Error(),
			ExitCode: -1,
		}, err
	}

	var stderr, stdout, err = a.cmdRun(ctx)

	var r = Response{
		ID: a.id,

		Command: a.executable,
		Args:    a.args,
		CmdLine: strings.Join(append([]string{a.executable}, a.args...), " "),

		Stderr: stderr,
		Stdout: Output(stdout),

		ExitCode: getExitStatus(err),

		Dir: a.dir,
	}

	if err != nil {
		r.embedError(err)
	}

	if esr := a.maybeSaveResponse(r); esr != nil {
		log.Errorf("cannot save response for request %v: %v", a.id, esr)
	}

	return r, err
}

func (a *Apply) cmdRun(ctx context.Context) (stderr, stdout string, err error) {
	var cmd = exec.CommandContext(ctx, a.name, a.args...) // #nosec

	var (
		buf    bytes.Buffer
		bufErr bytes.Buffer
	)

	if a.checkStateful() {
		cmd.Dir = a.dir
	}

	cmd.Stderr = &bufErr
	cmd.Stdout = &buf

	err = cmd.Run()
	return bufErr.String(), buf.String(), err
}

// checkStateful checks if it is needed to save anything or you can just safely run the command
func (a *Apply) checkStateful() bool {
	if a.dontSave {
		return false
	}

	if len(a.args) == 0 || len(a.Files) == 0 {
		return false
	}

	for _, a := range a.args {
		switch a {
		case "-h", "--help":
			return false
		}
	}

	return true
}

func (a *Apply) maybeSaveResponse(r Response) error {
	if !a.checkStateful() {
		return nil
	}

	return a.saveResponse(r)
}

func (a *Apply) saveResponse(r Response) error {
	var b, err = json.MarshalIndent(r, "", "    ")

	if err != nil {
		return err
	}

	b = append(b, '\n')

	return a.saveFile("response", b)
}

func (a *Apply) maybeConfigure() error {
	if !a.checkStateful() {
		return nil
	}

	return a.configure()
}

func (a *Apply) configure() error {
	if err := a.checkUploads(); err != nil {
		return err
	}

	if err := a.initConfigurationDir(); err != nil {
		return err
	}

	if err := a.saveDescription(); err != nil {
		return err
	}

	if err := a.copyConfigurationFiles(); err != nil {
		return err
	}

	return a.saveFile("request", a.RequestDump)
}

var blacklist = map[string]struct{}{
	"description": {},
	"request":     {},
	"response":    {},
}

func (a *Apply) checkUploads() error {
	for f := range a.Files {
		_, blocked := blacklist[f]

		if blocked || strings.Contains(f, "..") || filepath.IsAbs(f) {
			return fmt.Errorf(`refusing to apply: unsafe filepath "%s"`, f)
		}
	}

	return nil
}

func (a *Apply) initConfigurationDir() error {
	if err := os.MkdirAll(a.dir, dirFileMode); err != nil {
		return fmt.Errorf("can't create files configuration directory: %v", err)
	}

	return nil
}

func (a *Apply) copyConfigurationFiles() error {
	for f, v := range a.Files {
		file := filepath.Join(a.dir, f)

		if err := os.MkdirAll(filepath.Dir(file), dirFileMode); err != nil {
			return fmt.Errorf("can't create files configuration directory: %v", err)
		}

		if err := ioutil.WriteFile(file, v, fileMode); err != nil {
			return fmt.Errorf("error writing %s (%s): %v", f, file, err)
		}
	}

	return nil
}

const descriptionTemplate = `ID: %s
Date: %v
IP: %v

Command:
%s %s

List of files:
%v
`

func (a *Apply) saveDescription() error {
	var files = a.listFiles()

	var description = []byte(fmt.Sprintf(descriptionTemplate,
		a.id,
		a.timestamp.Format(time.RubyDate),
		a.IP,
		a.name,
		strings.Join(a.args, " "),
		strings.Join(files, "\n"),
	))

	return a.saveFile("description", description)
}

func (a *Apply) saveFile(name string, b []byte) error {
	file := filepath.Join(a.dir, name)

	if err := ioutil.WriteFile(file, b, fileMode); err != nil {
		return fmt.Errorf("cannot write %s: %v", file, err)
	}

	return nil
}

func getExitStatus(err error) int {
	if err == nil {
		return 0
	}

	if exit, ok := err.(*exec.ExitError); ok {
		if process, ok := exit.Sys().(syscall.WaitStatus); ok {
			return process.ExitStatus()
		}
	}

	return -1
}

func (a *Apply) init() {
	a.m.Lock()
	defer a.m.Unlock()

	if a.id == "" {
		a.id = uuid.NewV4().String()

		a.timestamp = time.Now()
		a.dir = filepath.Join(configurations,
			a.timestamp.Format("2006-01-02"),
			fmt.Sprintf("%v", a.timestamp.Unix())+"-"+a.id)
	}

	a.name, a.args = a.unsafeCommand()
}
