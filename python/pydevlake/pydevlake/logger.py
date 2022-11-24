import sys
import logging

logging.basicConfig(
    level=logging.DEBUG,
    format='%(levelname)s: %(message)s', 
    stream=sys.stdout
)

logger = logging.getLogger()
