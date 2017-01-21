# Gerrit Testing With  Docker

[![Build Status](https://travis-ci.org/opalmer/gerrittest.svg?branch=master)](https://travis-ci.org/opalmer/gerrittest)


This project is meant to assist in testing Gerrit. It provides a docker
container to run Gerrit and a Makefile with some useful helpers.

## Setup

* Install docker
* Install gerrittest, typically inside a virtualenv, one of two ways:
  * `pip install gerrittest`
  * Clone down down the repository, `pip install -e .` 

## Command

The gerrittest package provides a `gerrittest` command. This command has
a few different sub-commands, use `--help` to see them.

## Run

Below is an example of how the `gerrittest` command could be used. In this
case the following takes places:

* Run a container with Gerrit
* Obtain the http/ssh ports
* Create the admin account, add an ssh key and return it.

```
> container_id=$(gerrittest run)
> ssh_port=$(gerrittest get-port ssh $container_id)
> http_port=$(gerrittest get-port http $container_id)
> rsa_key=$(gerrittest --log-level debug create-admin $container_id)
2017-01-21 17:18:10,802 DEBUG docker inspect --type container d3bd5a5205a6cd10da96446a1d14d96de11081c792e1d25cc1476e3dc441602b
2017-01-21 17:18:10,812 DEBUG docker inspect --type container d3bd5a5205a6cd10da96446a1d14d96de11081c792e1d25cc1476e3dc441602b
2017-01-21 17:18:10,820 DEBUG Creating admin account.
2017-01-21 17:18:10,927 DEBUG GET http://172.30.0.1:32873/login/%23%2F?account_id=1000000 (response: 200)
2017-01-21 17:18:10,961 DEBUG GET http://172.30.0.1:32873/a/accounts/self (response: 200)
2017-01-21 17:18:10,961 DEBUG Generating RSA key.
2017-01-21 17:18:10,964 DEBUG ssh-keygen -b 2048 -t rsa -f /tmp/tmphDJOEc/id_rsa -q -N ""
2017-01-21 17:18:10,998 DEBUG docker inspect --type container d3bd5a5205a6cd10da96446a1d14d96de11081c792e1d25cc1476e3dc441602b
2017-01-21 17:18:11,007 DEBUG Adding RSA key /tmp/tmphDJOEc/id_rsa to admin
2017-01-21 17:18:11,007 DEBUG POST http://172.30.0.1:32873/a/accounts/self/sshkeys
2017-01-21 17:18:11,109 DEBUG ssh -o LogLevel=quiet -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i /tmp/tmphDJOEc/id_rsa -p 32872 admin@172.30.0.1 gerrit version
> echo $container_id $ssh_port $http_port $rsa_key
d3bd5a5205a6cd10da96446a1d14d96de11081c792e1d25cc1476e3dc441602b 32774 32775 /tmp/tmphDJOEc/id_rsa
```
