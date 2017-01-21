# Gerrit Testing With  Docker

[![Build Status](https://travis-ci.org/opalmer/gerrittest.svg?branch=master)](https://travis-ci.org/opalmer/gerrittest)


This project is meant to assist in testing Gerrit. It provides a docker
container to run Gerrit and a Makefile with some useful helpers.

## Setup

* Install docker
* Install gerrittest, typically inside a virtualenv, one of two ways:
  * `pip install gerrittest`
  * Clone down CreateClone down the repository, `pip install -e .` 

## Command

The gerrittest package provides a `gerrittest` command. This command has
a few commands and options:
 
```
usage: gerrittest [-h] [--log-level {debug,info,warn,warning,error,critical}]
                  {run,get-port,self-test} ...

Wraps the the `docker` command to run gerrittests

optional arguments:
  -h, --help            show this help message and exit
  --log-level {debug,info,warn,warning,error,critical}
                        Sets the logging level for gerrittest. This does not
                        impact command line output.

Subcommands:
  {run,get-port,self-test}
    run                 Runs Gerrit in the docker container.
    get-port            Returns the requested port for the provided container.
    self-test           Runs a sequence of sub-commands intended to 'self
                        test' the gerrittest command.
```

## Run

**Running Gerrit**
```bash
> container_id=$(gerrittest --log-level debug run)
2017-01-21 14:14:02,528 DEBUG docker version
2017-01-21 14:14:02,537 DEBUG docker run --detach --publish 8080 --publish 29418 opalmer/gerrittest:latest
> echo $container_id
e3e7d684faa0110a6243186d0ff9b7379cf1dc068f731a3a60822901e002fa71
> ssh_port=$(gerrittest --log-level debug get-port ssh $container_id)
2017-01-21 14:16:46,948 DEBUG docker inspect --type container e3e7d684faa0110a6243186d0ff9b7379cf1dc068f731a3a60822901e002fa71
> http_port=$(gerrittest --log-level debug get-port http $container_id)
> echo $ssh_port $http_port
32774 32775
```
