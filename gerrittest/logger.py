
import logging
import sys

HANDLER = logging.StreamHandler(sys.stderr)
HANDLER.setFormatter(
    logging.Formatter("%(asctime)s %(levelname)s %(message)s"))
logger = logging.getLogger("gerrittest")
logger.setLevel(logging.INFO)
logger.addHandler(HANDLER)
