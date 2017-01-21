# Gerrit Testing With  Docker

[![Build Status](https://travis-ci.org/opalmer/gerrittest.svg?branch=master)](https://travis-ci.org/opalmer/gerrittest)


This project is meant to assist in testing Gerrit. It provides a docker
container to run Gerrit and a Makefile with some useful helpers.

## Setup

* Install docker
* Install gerrittest, typically inside a virtualenv, one of two days:
  * `pip install gerrittest`
  * Clone down CreateClone down the repository, `pip install -e .` 


## Run

**Using Default Ports**
```bash
> gerrittest run
39bce5010cd0bb34889cc1f20dd6251a54c41aa717342d4e2a2fe8fc9ac91102 8080 29418
```

**Using Randomly Mapped Ports**

```bash
> gerrittest run --http 0 --ssh 0
d9b38348d075c96af7691abe0e9a4b74fd293062bd8a329a6185116769a80fff
```

**Retrieve Ports After Launch**

```bash
> gerrittest get-port ssh d9b
32776
> gerrittest get-port http d9b
32777
```