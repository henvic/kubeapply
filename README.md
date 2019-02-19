# kubeapply
[![GoDoc](https://godoc.org/github.com/henvic/kubeapply?status.svg)](https://godoc.org/github.com/henvic/kubeapply) [![Build Status](https://travis-ci.org/henvic/kubeapply.svg?branch=master)](https://travis-ci.org/henvic/kubeapply) [![Coverage Status](https://coveralls.io/repos/henvic/kubeapply/badge.svg)](https://coveralls.io/r/henvic/kubeapply) [![codebeat badge](https://codebeat.co/badges/9bea91b8-e09c-43da-96c0-ac8aaa967c24)](https://codebeat.co/projects/github-com-henvic-kubeapply-master) [![Go Report Card](https://goreportcard.com/badge/github.com/henvic/kubeapply)](https://goreportcard.com/report/github.com/henvic/kubeapply)

kubeapply is a microservice for running `kubectl apply` through a web API.

kubeapply makes it easier to use Kubernetes in a declarative manner while Kubernetes API still doesn't have an endpoint similar to `kubectl apply`.

## Server-side Apply enhancement workaround
`kubectl apply` is a core part of the Kubernetes config workflow. However, its implementation is in the client-side of a CLI tool. As of February 2019, there is work in progress to migrate the functionality to the server-side.

This middleware is a workaround useful for using Kubernetes `kubectl apply` over an HTTP connection while work on this integration is still in progress.

* [Kubernetes Enhancement Proposals: Apply](https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/0006-apply.md)
* [v2 API proposal "desired vs actual #17333"](https://issues.k8s.io/17333)
* [Server-side Apply #555](https://github.com/kubernetes/enhancements/issues/555)
* [Umbrella Issue for Server Side Apply #73723](https://issues.k8s.io/73723)

## Dependencies

* Go ≥ 1.11 to generate the server binary.
* [Kubernetes](https://www.kubernetes.io) 1.10 or greater.

## Commands
* Run `make server`

You might want to run `cmd/server --help` to list the available options.

The environment variable `DEBUG` sets the logging to debug mode.

`kubectl` must be available on the machine.

## Security
It is unsafe to run this software unless you protect this service appropriately.
You must run it on an isolated machine with limited network connectivity.

Port 8080 (API) is only accessible from localhost.
Port 8081 (debugging tool) is enabled by default (also only for localhost).

For your safety, you must assume that anyone who can reach this middleware endpoints has total control over the machine it is running on. Reasons: cluster options, file-system access, etc.

To communicate with other machines outside of a trusted network use a secure layer and proper client and server authentication protocols.

## Endpoints

### /version
`curl http://localhost:8080/version -v` returns the local `kubectl` version.

### /apply

You can use all flags available on `kubectl apply` (including global ones).

```json
{
	"flags": {
		"dry-run": true,
		"output": "wide"
	},
	"files": {
		"relative/path/cert.yaml": "apiVersion: ...",
		"server.json": {"apiVersion": "..."}
	}
}
```

A JSON object is returned containing the explanation of the executed command and its result.

Some details:

* `cmd_line` is the corresponding command you can copy and paste on a shell to execute the command yourself.
* `exit_code` is the process exit code.
* `dir` is the relative path to the stored configuration and logs.
* `stderr` is always a string.
* `stdout` is JSON body by default. For other output formats, it is returned as a string value.

#### Recordings and logs
Configurations requested are recorded on a directory inside `configurations` named by the id of the request and organized by date. No rotation policy is in place.

You don't need to pass the `--filename` flag as if no file is found on your YAML, `--filename=./` and `--recursive` are automatically set.

Run example with --dry-run:
`curl -d @example.json -v -XPUT http://localhost:8080/apply -H "Content-Type: application/json" | jq`

#### Flags
You don't need to prefix flags or shortcuts with `--` or `-`. You also can use numbers or booleans directly.

Example:

```json
{
	"flags": {
		"--dry-run": true,
		"timeout": "1m",
		"R": true,
		"f": "service.yaml"
	}
}
```

This configuration is similar to `kubectl apply --dry-run=true --timeout=1m -R -f=service.yaml`.

## Contributing
You can get the latest source code with `go get -u github.com/henvic/kubeapply`

The following commands are available and require no arguments:

* **make test**: run tests

In lieu of a formal style guide, take care to maintain the existing coding style. Add unit tests for any new or changed functionality. Integration tests should be written as well.

## Committing and pushing changes
The master branch of this repository on GitHub is protected:
* force-push is disabled
* tests MUST pass on Travis before merging changes to master
* branches MUST be up to date with master before merging

Keep your commits neat and [well documented](https://wiki.openstack.org/wiki/GitCommitMessages). Try to always rebase your changes before publishing them.

## Maintaining code quality
[goreportcard](https://goreportcard.com/report/github.com/henvic/kubeapply) can be used online or locally to detect defects and static analysis results from tools with a great overview.

Using go test and go cover are essential to make sure your code is covered with unit tests.

Always run `make test` before submitting changes.
