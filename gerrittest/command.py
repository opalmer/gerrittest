import subprocess
from gerrittest.logger import logger


def check_output(args, **kwargs):
    """Simple wrapper around subprocess.check_output so we can log commands."""
    if kwargs.pop("logged", True):
        keywords = ""
        if kwargs:
            keywords = "(%s)" % ", ".join(
                ["=".join(map(str, [k,v])) for k, v in kwargs.items()])
        logger.debug("%s %s", " ".join(args), keywords)
    return subprocess.check_output(args, **kwargs)
