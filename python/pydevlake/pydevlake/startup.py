import inspect

import fire

from .core.plugin_init import PluginInit
from .core.default_api import DefaultPluginAPI
from .core.models import PluginInfo
from .core.registry import registered_types


class Entry(object):
    def __init__(self, *objs):
        for obj in objs:
            for [name, f] in inspect.getmembers(obj, predicate=inspect.ismethod):
                if dict(f.__annotations__).get("callable") is not None:
                    setattr(self, name, f)


def init_cmd(script_path: str, plugin_info: PluginInfo):
    fire.Fire(Entry(
        PluginInit(plugin_info, script_path),
        DefaultPluginAPI(plugin_info),
        *registered_types))
