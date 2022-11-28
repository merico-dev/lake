import json
import os
from argparse import Namespace
from typing import Generator, TextIO
import jsonpickle

from pydevlake.core import ApiDoc
from pydevlake.core.api import PluginState
from pydevlake.core.doc import normalize_doc_types
from pydevlake.core.registry import registered_types
from pydevlake.core.swagger import docgen


def normalize(arg) -> object:
    return json.dumps(arg) if isinstance(arg, dict) else arg


def plugin_class(kls):
    registered_types.append(kls())
    return kls


def plugin_method(**opts):
    json_serialized = True
    api_doc: ApiDoc = None
    if opts.get("json_serialized") is not None:
        json_serialized = opts.get("json_serialized")
    if opts.get("api_doc") is not None:
        api_doc = opts.get("api_doc")

    def wrapped_plugin_method(func):
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
            modified = args
            if json_serialized:
                modified = [args[0]] + [convert(arg) for arg in args[1:]]
            ret = func(*modified, **kwargs)
            if ret is not None:
                with open_send_channel() as send_ch:
                    if isinstance(ret, Generator):
                        for each in ret:
                            send_output(send_ch, each)
                    else:
                        send_output(send_ch, ret)
            return None

        if api_doc is not None:
            if api_doc.types is not None:
                resolved = normalize_doc_types(*api_doc.types)
                func.__doc__ = api_doc.doc.format(*resolved)
            else:
                func.__doc__ = api_doc.doc
            docgen.generate_doc("/plugins/{}".format(api_doc.path), func)

        func.__annotations__["callable"] = True
        wrapper.__doc__ = func.__doc__
        wrapper.__annotations__ = func.__annotations__
        return wrapper

    return wrapped_plugin_method


def convert(arg):
    return json.loads(normalize(arg), object_hook=lambda d: Namespace(**d))


class State(PluginState):

    def __init__(self):
        import signal
        self.is_cancelled = False
        signal.signal(signal.SIGTERM, self.__handle_cancellation__)

    def is_cancelled(self) -> bool:
        return self.is_cancelled

    def terminate(self):
        import sys
        sys.exit(0)

    def __handle_cancellation__(self, signum, frame):
        print("received cancellation request")
        # impl code has to use this variable
        self.is_cancelled = True


state: PluginState = State()
