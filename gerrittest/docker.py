
from __future__ import print_function

import json
import sys

from gerrittest.logger import logger
from gerrittest.command import check_output

DEFAULT_IMAGE = "opalmer/gerrittest:latest"
DEFAULT_HTTP = 8080
DEFAULT_SSH = 29418
DEFAULT_RANDOM = "random"


def get_run_command(
        image=DEFAULT_IMAGE, ip=None,
        http_port=DEFAULT_HTTP, ssh_port=DEFAULT_SSH):
    """
    Constructs and returns the full `docker run` command based
    on the provided inputs.

    :param str image:
        The name of the docker image to return a command for

    :param str ip:
        An explicit address which ports should be published to.

    :param int http_port:
        The port to publish the http service on. Supply `random` to bind
        the service to a random port.

    :param int ssh_port:
    The port to publish the ssh service on. Supply `random` to bind
        the service to a random port.

    """
    command = ["docker", "run", "--detach"]

    try:
        http_port = int(http_port)
    except ValueError:
        if http_port != "random":
            raise ValueError("http port must be 'random' or an integer.")

    try:
        ssh_port = int(ssh_port)
    except ValueError:
        if ssh_port != "random":
            raise ValueError("ssh port must be 'random' or an integer.")

    publish_prefix = ""
    if ip is not None:
        publish_prefix += "%s:" % ip

    publish_components = []
    if ip is not None:
        publish_components += [ip]

    if http_port == DEFAULT_RANDOM:
        http = publish_components + [str(DEFAULT_HTTP)]
    else:
        http = publish_components + [str(http_port), str(DEFAULT_HTTP)]

    if ssh_port == DEFAULT_RANDOM:
        ssh = publish_components + [str(DEFAULT_SSH)]
    else:
        ssh = publish_components + [str(ssh_port), str(DEFAULT_SSH)]

    return command + [
        "--label", "gerrittest=1",
        "--publish", ":".join(http),
        "--publish", ":".join(ssh),
        image # must always be last
    ]


def inspect(container_id, required_status="running"):
    """
    Returns the results of the `docker inspect` command as
    json.

    :param str container_id:
        The container id to return status for.

    :param str required_status:
        The status of the container must match this value otherwise
        ValueError will be raised.
    """
    results = json.loads(check_output(
        ["docker", "inspect", "--type", "container", container_id]))

    if len(results) != 1:
        logger.error(
            "Expected exactly one container but found %s for %s",
            len(results), container_id)
        sys.exit(1)

    data = results[0]
    if required_status and data["State"]["Status"] != required_status:
        logger.error(
            "Container %s (state: %s, expected: %s)",
            container_id, data["State"]["Status"], required_status)
        raise ValueError("Unxpected container status.")

    return results[0]


def get_port(internal_port, container):
    """
    This function will return the exposed port for the requested internal port
    and container.
    """
    if internal_port == "http":
        internal_port = DEFAULT_HTTP
    if internal_port == "ssh":
        internal_port = DEFAULT_SSH
    assert isinstance(internal_port, int)

    data = inspect(container)
    for port, info in data["NetworkSettings"]["Ports"].items():
        if port.split("/")[0] == str(internal_port):
            if len(info) != 1:
                raise ValueError("Expected exactly one entry for the port")
            return int(info[0]["HostPort"])


def list_containers(show_all=False):
    """
    This function prints out a list of all gerritest containers. Any
    container run with `gerritest run` should be returned.
    """
    command = ["docker", "ps", "--format", '"{{.ID}}"']
    if show_all:
        command += ["-a"]

    for container_id in check_output(command).splitlines():
        container_id = container_id.replace('"', "")
        data = inspect(container_id, required_status="")

        if data["Config"]["Labels"].get("gerrittest") == "1":
            yield container_id.strip()


def remove_container(container_id):
    """Removes the requested container. Ignores any errors"""
    return check_output(["docker", "rm", "-f", container_id])


def get_network_gateway(container_id):
    return inspect(container_id)["NetworkSettings"]["Gateway"]
