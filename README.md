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

Below is an example of how the `gerrittest` command could be used. This 
output is pulled from ``./test.sh``

```
++ gerrittest --log-level debug run
2017-01-21 18:00:56,367 DEBUG docker run --detach --label gerrittest=1 --publish 8080 --publish 29418 opalmer/gerrittest:latest 
+ container_id=999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36
+ gerrittest --log-level debug wait 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36
2017-01-21 18:00:56,922 DEBUG Waiting for 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36 to become active
2017-01-21 18:00:56,922 DEBUG docker inspect --type container 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36 
2017-01-21 18:00:56,932 DEBUG docker inspect --type container 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36 
2017-01-21 18:00:56,942 DEBUG docker inspect --type container 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36 
2017-01-21 18:00:56,951 DEBUG Waiting on 172.30.0.1:32893
2017-01-21 18:01:03,269 DEBUG docker inspect --type container 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36 
2017-01-21 18:01:03,283 DEBUG Waiting on 172.30.0.1:32892
+ gerrittest --log-level debug get-port ssh 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36
2017-01-21 18:01:03,454 DEBUG docker inspect --type container 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36 
32892
+ gerrittest --log-level debug get-port http 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36
2017-01-21 18:01:03,616 DEBUG docker inspect --type container 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36 
32893
+ gerrittest --log-level debug create-admin 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36
2017-01-21 18:01:03,768 DEBUG docker inspect --type container 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36 
2017-01-21 18:01:03,777 DEBUG docker inspect --type container 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36 
2017-01-21 18:01:03,786 DEBUG Creating admin account.
2017-01-21 18:01:03,873 DEBUG GET http://172.30.0.1:32893/login/%23%2F?account_id=1000000 (response: 200)
2017-01-21 18:01:03,908 DEBUG GET http://172.30.0.1:32893/a/accounts/self (response: 200)
2017-01-21 18:01:03,909 DEBUG Generating RSA key.
2017-01-21 18:01:03,911 DEBUG ssh-keygen -b 2048 -t rsa -f /tmp/tmpY8ljAL/id_rsa -q -N ""
2017-01-21 18:01:04,039 DEBUG docker inspect --type container 999ed2823460ab340eaf52a912d29c2ab3daccbf38e2474d7d4cb96aeb0cde36 
2017-01-21 18:01:04,049 DEBUG Adding RSA key /tmp/tmpY8ljAL/id_rsa to admin
2017-01-21 18:01:04,049 DEBUG POST http://172.30.0.1:32893/a/accounts/self/sshkeys
2017-01-21 18:01:04,158 DEBUG ssh -o LogLevel=quiet -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i /tmp/tmpY8ljAL/id_rsa -p 32892 admin@172.30.0.1 gerrit version 
/tmp/tmpY8ljAL/id_rsa

```