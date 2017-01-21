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
    DEFAULT_IMAGE, DEFAULT_HTTP, DEFAULT_SSH, get_run_command, get_port)
from gerrittest.logger import logger


def subcommand_run(args):
    """Implements subcommand `run`"""
    command = get_run_command(
        image=args.image, ip=args.ip, http_port=args.http, ssh_port=args.ssh)
    print(check_output(command).strip())


def subcommand_get_port(args):
    """Implements subcommand `get-port`"""
    if args.port == "http":
        internal_port = DEFAULT_HTTP
    elif args.port == "ssh":
        internal_port = DEFAULT_SSH
    else:
        return

    port = get_port(internal_port, args.container)
    if port is None:
        sys.exit(1)
    print(port)


def add_container_argument(parser):
    parser.add_argument(
        "container", help="The container to retrieve the port for.")


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
    add_container_argument(ports)
    ports.set_defaults(func=subcommand_get_port)
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

