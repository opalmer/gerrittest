# Gerrit Testing With  Docker

[![Build Status](https://travis-ci.org/opalmer/gerrittest.svg?branch=master)](https://travis-ci.org/opalmer/gerrittest)


This project is meant to assist in testing Gerrit. It provides a docker
container to run Gerrit and a Makefile with some useful helpers.

## Setup

* Install docker
* Install gerrittest, typically inside a virtualenv, one of two ways:
  * `pip install gerrittest`
  * Clone down down the repository, `pip install -e .`
   
## Testing

The gerrittest project can be tested locally. To build the container and
the gerrittest command run:

```
make check
```

## Command

The gerrittest package provides a `gerrittest` command. This command has
a few different sub-commands, use `--help` to see them.

## Run

Below is an example of how the `gerrittest` command could be used:

```
$ container_id=$(gerrittest run)
$ gerrittest wait $container_id
$ ssh_port=$(gerrittest get-port ssh $container_id)
$ http_port=$(gerrittest get-port http $container_id)
$ rsa_key=$(gerrittest create-admin $container_id)
```

When the above finishes Gerrit will be up and running with an admin user whose
password is 'secret' which can be used with digest auth to query the REST
API:

```
$ curl --digest -u admin:secret http://localhost:$http_port/a/accounts/self
)]}'
{
  "_account_id": 1000000,
  "name": "Administrator",
  "email": "admin@example.com",
  "username": "admin"
}
```

The outputs above can also be used with ssh:
```
$ ssh -o LogLevel=quiet -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i $rsa_key -p $ssh_port admin@localhost gerrit version
gerrit version 2.13.5
```

Or git, with some extra options:

```
$ GIT_SSH="ssh -o LogLevel=quiet -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i $rsa_key -p $ssh_port"
$ GIT_SSH=GIT_SSH git ...
```
