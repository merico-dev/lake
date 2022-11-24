import json
import os
from argparse import Namespace
from typing import Generator, TextIO

import jsonpickle


def normalize(arg) -> object:
    return json.dumps(arg) if isinstance(arg, dict) else arg


def plugin_method(func):
    def open_send_channel() -> TextIO:
        # start_debugger()
        # TODO make this cross-platform (named-pipes?)
        # fd = os.open('.temp', os.O_CREAT | os.O_WRONLY)
        # assert fd == 3
        fd = 3
        return os.fdopen(fd, 'w')

    def send_output(send_ch: TextIO, obj: object):
        if obj is None:
            return
        encoded_ret = jsonpickle.encode(obj, unpicklable=False)
        send_ch.write(encoded_ret)
        send_ch.write('\n')
        send_ch.flush()

    def wrapper(*args, **kwargs):
        modified = [args[0]] + [json.loads(normalize(arg), object_hook=lambda d: Namespace(**d)) for arg in args[1:]]
        ret = func(*modified, **kwargs)
        if ret is not None:
            with open_send_channel() as send_ch:
                if isinstance(ret, Generator):
                    for each in ret:
                        send_output(send_ch, each)
                else:
                    send_output(send_ch, ret)
        return None

    wrapper.__doc__ = func.__doc__
    wrapper.__annotations__ = func.__annotations__
    return wrapper


def convert(arg):
    return json.loads(normalize(arg), object_hook=lambda d: Namespace(**d))
