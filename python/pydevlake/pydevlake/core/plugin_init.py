import pathlib

from .api import *
from .ipc import plugin_method
from .registry import Registry


class PluginInit(object):
    def __init__(self, plugin_info: PluginInfo, script_path: str):
        plugin_info.plugin_path = str(pathlib.Path(script_path).resolve())
        self.endpoint = "not_set"
        self.plugin_info = plugin_info
        self.plugin_path = self.plugin_info.plugin_path

    @plugin_method(json_serialized=False)
    def startup(self, endpoint: str):
        self.endpoint = endpoint
        registry = Registry(endpoint)
        registry.register_plugin(self.plugin_info)
        print("python plugin={} completed initialization.".format(self.plugin_info.name))
