import tempfile
import os
from os.path import join

import requests
from requests.cookies import RequestsCookieJar
from requests.auth import HTTPDigestAuth

from gerrittest.logger import logger
from gerrittest.command import check_output

AUTH_ADMIN = HTTPDigestAuth("admin", "secret")


def create_admin(address, port):
    """Creates the admin account and returns (username, secret)"""
    cookies = RequestsCookieJar()
    base_url = "http://{address}:{port}".format(address=address, port=port)

    url = "{}/login/%23%2F?account_id=1000000".format(base_url)
    response = requests.get(url, cookies=cookies)
    logger.debug("GET %s (response: %s)", url, response.status_code)
    response.raise_for_status()

    # Try to login with the newly created admin account.
    url = "{}/a/accounts/self".format(base_url)
    response = requests.get(url, auth=AUTH_ADMIN)
    logger.debug("GET %s (response: %s)", url, response.status_code)
    response.raise_for_status()

    return "admin", "secret"


def generate_rsa_key():
    """Generates an RSA key for ssh. Returns the generated key."""
    dirname = tempfile.mkdtemp()
    path = join(dirname, "id_rsa")

    # TODO Figure out why this only works with os.system. With check_output
    # ssh-keygen basically ignore the -q/-N flags even with shell=True
    command = 'ssh-keygen -b 2048 -t rsa -f %s -q -N ""' % path
    logger.debug(command)
    os.system(command)

    return path
