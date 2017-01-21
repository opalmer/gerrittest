"""
This module provides the `gerritest` command line entrypoint and other
utilities for handling command line input.
"""

from __future__ import print_function
import argparse
import subprocess
import sys

from gerrittest.docker import (
    DEFAULT_IMAGE, DEFAULT_HTTP, DEFAULT_SSH, run, get_port)


def subcommand_run(args):
    """Implements subcommand `run`"""
    container_id, http_port, ssh_port = run(
        image=args.image, ip=args.ip, http_port=args.http, ssh_port=args.ssh)
    if args.verbose:
        print(container_id, http_port, ssh_port)
        return
    print(container_id)


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


def make_parser():
    """Creates and returns a parser for handling command line input"""
    parser = argparse.ArgumentParser(
        description="Wraps the the `docker` command to run gerrittests")
    subparsers = parser.add_subparsers(title="Subcommands")

    # subcommand: run
    run = subparsers.add_parser(
        "run", help="Runs Gerrit in the docker container.")
    run.set_defaults(func=subcommand_run)
    run.add_argument(
        "-v", "--verbose", default=False, action="store_true",
        help="If provided then display the mapped ports too. Without this "
             "only the spawned container's ID will be shown.")
    run.add_argument(
        "--image", default=DEFAULT_IMAGE,
        help="The docker image to test with.")
    run.add_argument(
        "--ip", default=None,
        help="The IP address to publish ports on.")
    run.add_argument(
        "--http", default=DEFAULT_HTTP, type=int,
        help="Defines which local port will be mapped to the exported "
             "http port. 8080 by default, 0 will assign a random port.")
    run.add_argument(
        "--ssh", default=DEFAULT_SSH, type=int,
        help="Defines which local port will be mapped to the exported "
             "ssh port. 29418 by default, 0 will assign a random port.")

    # subcommand: ports
    run = subparsers.add_parser(
        "get-port",
        help="Returns the requested port for the provided container.")
    run.add_argument(
        "port", choices=("http", "ssh"), help="The port to retrieve.")
    run.add_argument(
        "container", help="The container to retrieve the port for.")
    run.set_defaults(func=subcommand_get_port)
    return parser


def main():
    """The entrypoint for the `gerrittest` command."""
    parser = make_parser()
    args = parser.parse_args()

    try:
        subprocess.check_output(["docker", "version"])
    except subprocess.CalledProcessError:
        raise RuntimeError(
            "`docker version` failed. Please make sure docker is running and "
            "that you can connect to it.")

    args.func(args)

