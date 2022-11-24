import pathlib
import distutils
import fire

from .core.abstract_plugin import AbstractPlugin


def init_cmd(script_path: str, plugin: AbstractPlugin):
    plugin_path = str(pathlib.Path(script_path).resolve())
    plugin.plugin_path = plugin_path
    fire.Fire(plugin)
