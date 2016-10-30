#!/bin/bash -e

# This file is responsible for testing gerrittest on Travis. The docker image
# must be built before this script can be run.

DOCKER_IMAGE=opalmer/gerrittest:latest

. ./docker/helper-functions.sh

RunContainer
CreateAdminAccount
private_key=$(GenerateSSHPrivateKey)
AddAdminSSHKey $private_key
