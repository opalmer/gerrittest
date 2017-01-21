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
    DEFAULT_IMAGE, DEFAULT_HTTP, DEFAULT_SSH, DEFAULT_RANDOM,
    get_run_command, get_port, list_containers, remove_container)
from gerrittest.logger import logger


def subcommand_run(args, silent=False):
    """Implements subcommand `run`"""
    command = get_run_command(
        image=args.image, ip=args.ip, http_port=args.http, ssh_port=args.ssh)
    container = check_output(command).strip()
    if not silent:
        print(container)
    return container


def subcommand_get_port(args, silent=False):
    """Implements subcommand `get-port`"""
    internal_port = 0
    if args.port == "http":
        internal_port = DEFAULT_HTTP
    if args.port == "ssh":
        internal_port = DEFAULT_SSH

    port = get_port(internal_port, args.container)
    if port is None:
        sys.exit(1)

    if not silent:
        print(port)
    return port


def subcommand_ps(args, silent=False):
    """Implements the subcommand: `ps`"""
    outputs = []
    for container_id in list_containers(show_all=args.all):
        outputs.append(container_id)
        if not silent:
            print(container_id)
    return outputs


def subcommand_kill(args, silent=False):
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
        if not silent:
            print(output)

    return outputs


def subcommand_self_test(args):
    """Implements the subcommand: `self-test`"""
    # Add additional arguments that are not set
    # by the 'self-test' sub-command.
    args.image = DEFAULT_IMAGE
    args.http = DEFAULT_RANDOM
    args.ssh = DEFAULT_RANDOM
    args.ip = None

    # subcomman: run
    container_id = subcommand_run(args, silent=True)
    args.container = container_id
    logger.info("Created container %s", container_id)

    # subcommand: get-port
    for port in ("http", "ssh"):
        args.port = port
        port = subcommand_get_port(args, silent=True)
        logger.info("   %s port: %s", args.port, port)

    # subcommand: ps
    args.all = True
    for found_container_id in subcommand_ps(args, silent=True):
        logger.info("Found container: %s", found_container_id)

    # subcommand: kill
    args.containers = [container_id]

    for container in subcommand_kill(args, silent=True):
        logger.info("Killed %s", container)


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

    # subcommand: self-check
    self_test = subparsers.add_parser(
        "self-test",
        help="Runs a sequence of sub-commands intended to 'self test' "
             "the gerrittest command.")
    self_test.set_defaults(func=subcommand_self_test)
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

