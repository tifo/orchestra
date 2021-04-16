Orchestra [![wercker status](https://app.wercker.com/status/16ba07e3d295feb5c3874207a9f3fe36/s "wercker status")](https://app.wercker.com/project/bykey/16ba07e3d295feb5c3874207a9f3fe36) [![GoDoc](https://godoc.org/github.com/vinceprignano/orchestra?status.svg)](https://godoc.org/github.com/vinceprignano/orchestra)
======================================================
Orchestra is a toolkit to manage a fleet of Go binaries/services. A unique place where you can run, stop, aggregate logs and config your Go binaries.

![](https://cloud.githubusercontent.com/assets/3118335/6255612/4811c940-b7a9-11e4-8d06-966981de3926.png)

> You can find an application design/proposal document [here](https://github.com/vinceprignano/orchestra/blob/master/DESIGN.md)

Build & Install
---------------
`go get -u github.com/tifo/orchestra`

Start an Orchestra Project
--------------------------
You should have an `orchestra.yml` file in your root directory and a `service.yml` file in every service directory.

```
.
├── first-service
│   ├── main.go
│   └── service.yml			<- Service file
├── second-service
│   ├── second.go
│   ├── main.go
│   └── service.yml			<- Service file
└── orchestra.yml           <- Main project file
```

You can specify a custom configuration file using the `--config` flag or setting the `ORCHESTRA_CONFIG` env variable.

By default orchestra will use `go install` to install your binaries in `GOPATH/bin`.

## Example
```yaml
env:
    ABC: "somethingGlobal"
before:
    - "echo I am a global command before"
after:
    - "echo I am a global after"
```



Commands
--------
- **start** `--option [<service>...]` Starts every service
> _Options:_
>
> `--attach, -a` Attach to services output after start
>
> `--logs, -l`	Start logging after start

- **stop** `--option [<service>...]` Stops every service
- **restart** `--option [<service>...]` Restarts every service
> _Options:_
>
> `--attach, -a` Attach to services output after start
>
> `--logs, -l`	Start logging after start

- **logs** `--option [<service>...]` Aggregates the output from the services
- **test** `--option [<service>...]` Runs `go test ./...` for every service
> _Options:_
>
> `-v` `--verbose` Run tests in verbose mode
>
> `-r` `--race` Run tests with race condition

- **ps** Displays the _status_ of every service, _process id_ and the _ports_ in use.

A service name can be prefixed with `~` to run a command in exclusion mode.
For example `orchestra start ~second-service` will start everything expect the second-service.

> When using `-a` or `--attach` with start/restart, the services will be spawned in the same ochestra's process group.

## Configuring commands
Every command can be configured separately with special environment variables or with before/after commands.

For example, in `orchestra.yml` you can configure to `echo AFTER START` before running `orchestra start` command.

```yaml
env:
    ABC: "A global env variable"
before:
    - "echo I am a global command before"
after:
    - "echo I am a global after"
start:
    env:
    	ABC: "Override in start"
    after:
    	- "echo AFTER START"
```

## Configuring services
You can use your `service.yml` to override the environment variables in a specific service. Variables specified on a service will always have precedence over the global ones.

```yaml
env:
    ABC: "Override in service"
```

Autocomplete
------------
Orchestra supports bash autocomplete.
```sh
source $GOPATH/src/github.com/vinceprignano/orchestra/autocomplete/orchestra
```

Licensing
---------
Orchestra is licensed under the Apache License, Version 2.0.
