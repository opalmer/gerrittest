
from __future__ import print_function

import json
import subprocess

DEFAULT_IMAGE = "opalmer/gerrittest:latest"
DEFAULT_HTTP = 8080
DEFAULT_SSH = 29418


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
        The port to publish the http service on. Use `0` to bind
        the service to a random port.

    :param int ssh_port:
    The port to publish the ssh service on. Use `0` to bind
        the service to a random port.

    """
    command = ["docker", "run", "--detach"]

    publish_prefix = ""
    if ip is not None:
        publish_prefix += "%s:" % ip

    publish_components = []
    if ip is not None:
        publish_components += [ip]

    if http_port == DEFAULT_HTTP:
        http = publish_components + [str(DEFAULT_HTTP), str(DEFAULT_HTTP)]
    elif http_port == 0:
        http = publish_components + [str(DEFAULT_HTTP)]
    else:
        http = publish_components + [str(http_port), str(DEFAULT_HTTP)]

    if ssh_port == DEFAULT_SSH:
        ssh = publish_components + [str(DEFAULT_SSH), str(DEFAULT_SSH)]
    elif ssh_port == 0:
        ssh = publish_components + [str(DEFAULT_SSH)]
    else:
        ssh = publish_components + [str(ssh_port), str(DEFAULT_SSH)]

    return command + [
        "--publish", ":".join(http), "--publish", ":".join(ssh), image]


def inspect(container_id, expected_status="running"):
    results = json.loads(
        subprocess.check_output(
            ["docker", "inspect", "--type", "container", container_id]))
    if len(results) != 1:
        raise ValueError(
            "Found %s containers for %r. Expected exactly "
            "one." % (len(results), container_id))

    data = results[0]
    if expected_status is not None and data["State"]["Status"] != expected_status:
        raise ValueError(
            "Container %r state: %r != %r" % (
                container_id, data["State"]["Status"], expected_status))
    return data


def get_port(internal_port, container):
    assert isinstance(internal_port, int)

    data = inspect(container)
    for port, info in data["NetworkSettings"]["Ports"].items():
        if port.split("/")[0] == str(internal_port):
            if len(info) != 1:
                raise ValueError("Expected exactly one entry for the port")
            return int(info[0]["HostPort"])


def run(**kwargs):
    command = get_run_command(**kwargs)
    container_id = subprocess.check_output(command).strip()
    http_port = get_port(DEFAULT_HTTP, container_id)
    ssh_port = get_port(DEFAULT_SSH, container_id)
    print(http_port, ssh_port)
