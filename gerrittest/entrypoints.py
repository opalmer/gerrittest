"""
This module provides the `gerritest` command line entrypoint and other
utilities for handling command line input.
"""

from __future__ import print_function

import argparse
import logging
import sys
from subprocess import CalledProcessError

from gerrittest.command import check_output
from gerrittest.docker import (
    DEFAULT_IMAGE, DEFAULT_HTTP, DEFAULT_SSH,
    get_run_command, get_port, list_containers, remove_container,
    get_network_gateway)
from gerrittest.helpers import create_admin, generate_rsa_key, add_rsa_key, wait
from gerrittest.logger import logger


def subcommand_run(args):
    """Implements subcommand `run`"""
    command = get_run_command(
        image=args.image, ip=args.ip, http_port=args.http, ssh_port=args.ssh)
    container = check_output(command).strip()
    print(container)
    return container


def subcommand_get_port(args):
    """Implements subcommand `get-port`"""
    internal_port = 0
    if args.port == "http":
        internal_port = DEFAULT_HTTP
    if args.port == "ssh":
        internal_port = DEFAULT_SSH

    port = get_port(internal_port, args.container)
    if port is None:
        sys.exit(1)

    print(port)
    return port


def subcommand_ps(args):
    """Implements the subcommand: `ps`"""
    outputs = []
    for container_id in list_containers(show_all=args.all):
        outputs.append(container_id)
        print(container_id)
    return outputs


def subcommand_kill(args):
    """Implements the subcommand: `kill-all`"""
    if not args.all and not args.containers:
        logger.error("Please specify containers(s) to remove or use --all")
        sys.exit(1)

    outputs = []
    for container_id in list_containers(show_all=True):
        remove = True
        if args.containers:

            for arg_container in args.containers:
                if arg_container.startswith(container_id):
                    break
            else:
                remove = False

        if not remove:
            logger.debug("Not removing %s (filtered)", container_id)
            continue

        output = remove_container(container_id).strip()
        outputs.append(output)
        print(output)

    return outputs


def subcommand_wait(args):
    """Implements the subcommand: `wait`"""
    wait(args.container)


def subcommand_create_admin(args):
    """Implements the subcommand: `create-admin`"""
    network_addr = get_network_gateway(args.container)
    http_port = get_port("http", args.container)

    username, password = create_admin(network_addr, http_port)
    key = args.key_file
    if not key:
        key = generate_rsa_key()

    ssh_port = get_port("ssh", args.container)
    add_rsa_key(network_addr, http_port, ssh_port, username, password, key)
    print(key)


def make_parser():
    """Creates and returns a parser for handling command line input"""
    parser = argparse.ArgumentParser(
        description="Wraps the the `docker` command to run gerrittests")
    parser.add_argument(
        "--log-level", default="info",
        choices=("debug", "info", "warn", "warning", "error", "critical"),
        help="Sets the logging level for gerrittest. This does not impact "
             "command line output.")
    subparsers = parser.add_subparsers(title="Subcommands")

    # subcommand: run
    run = subparsers.add_parser(
        "run", help="Runs Gerrit in the docker container.")
    run.set_defaults(func=subcommand_run)
    run.add_argument(
        "--image", default=DEFAULT_IMAGE,
        help="The docker image to test with.")
    run.add_argument(
        "--ip", default=None,
        help="The IP address to publish ports on.")
    run.add_argument(
        "--http", default="random",
        help="Defines what local port should be mapped to the exported "
             "http port. Defaults to 'random' but an explict port may be "
             "provided instead.")
    run.add_argument(
        "--ssh", default="random",
        help="Defines what local port should be mapped to the exported "
             "ssh port. Defaults to 'random' but an explit port may be "
             "provided instead.")

    # subcommand: ports
    ports = subparsers.add_parser(
        "get-port",
        help="Returns the requested port for the provided container.")
    ports.add_argument(
        "port", choices=("http", "ssh"), help="The port to retrieve.")
    ports.add_argument(
        "container", help="The container to retrieve the port for.")
    ports.set_defaults(func=subcommand_get_port)

    # subcommand: ps
    ps = subparsers.add_parser(
        "ps",
        help="Returns a list of running containers gerrittest containers")
    ps.add_argument(
        "-a", "--all", default=False, action="store_true",
        help="If provided then show all gerrittest containers, regardless of "
             "their current state.")
    ps.set_defaults(func=subcommand_ps)

    # subcommand: kill-all
    kill = subparsers.add_parser(
        "kill",
        help="Kills and removes containers started by gerrittest")
    kill.add_argument(
        "-a", "--all", default=False, action="store_true",
        help="If provided then kill any container started by "
             "gerrittest. If this flag is not provided then one or more "
             "containers must be provided as input.")
    kill.add_argument(
        "containers", nargs="*",
        help="Optional list of containers to kill. If not provided, all "
             "gerritest containers will be killed.")
    kill.set_defaults(func=subcommand_kill)

    # subcommand: wait
    wait = subparsers.add_parser(
        "wait",
        help="Waits for ssh/http to become available.")
    wait.add_argument(
        "container", help="The container to wait on")
    wait.set_defaults(func=subcommand_wait)

    # subcommand: create-admin
    create_admin = subparsers.add_parser(
        "create-admin",
        help="Creates an admin account and ssh key.")
    create_admin.add_argument(
        "-f", "--key-file",
        help="A path to an explict key file to add. By default a random file "
             "will be generated for you.")
    create_admin.add_argument(
        "container", help="The container to create an admin in.")
    create_admin.set_defaults(func=subcommand_create_admin)
    return parser


def main():
    """The entrypoint for the `gerrittest` command."""
    parser = make_parser()
    args = parser.parse_args()

    level = logging.getLevelName(args.log_level.upper())
    logger.setLevel(level)

    try:
        check_output(["docker", "version"], logged=False)
    except CalledProcessError:
        raise RuntimeError(
            "`docker version` failed. Please make sure docker is running and "
            "that you can connect to it.")

    args.func(args)

