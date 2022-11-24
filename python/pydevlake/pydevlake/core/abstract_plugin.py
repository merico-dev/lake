import signal
import sys
from abc import ABC, abstractmethod

from .api import *
from .doc import normalize_doc_types
from .registry import Registry
from .swagger import docgen


class AbstractPlugin(PluginAPI, PluginMethods, ABC):

    def __init__(self):
        self.endpoint = "not_set"
        self.is_cancelled = False
        signal.signal(signal.SIGTERM, self.__handle_cancellation__)
        self.plugin_info = self.get_plugin_info()
        self.plugin_path = self.plugin_info.plugin_path
        self.set_all_default_docs()

    def startup(self, endpoint: str):
        from ..test.debugger import start_debugger
        start_debugger()
        self.endpoint = endpoint
        if self.plugin_path != "":
            self.plugin_info.plugin_path = self.plugin_path
        registry = Registry(endpoint)
        plugins: list[AbstractPlugin]
        registry.register_plugin(self.plugin_info)
        print("python plugin={} completed initialization.".format(self.plugin_info.name))

    @abstractmethod
    def get_plugin_info(self) -> PluginInfo:
        pass

    def __handle_cancellation__(self, signum, frame):
        print("received cancellation request")
        # impl code has to use this variable
        self.is_cancelled = True

    def __terminate__(self):
        sys.exit(0)

    @plugin_method
    def post_connection(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @plugin_method
    def patch_connection(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @plugin_method
    def get_connection(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        """
        Get Connection.
        ---
        get:
          responses:
            200:
              content:
                application/json:
                  schema: {}
        """
        pass

    @plugin_method
    def list_connections(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        """
        Get all Connections.
        ---
        get:
          responses:
            200:
              content:
                application/json:
                  schema: "{}"
        """
        pass

    @plugin_method
    def delete_connection(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    def set_default_docs(self, func, path, *args):
        normalized_args = normalize_doc_types(*args)
        doc_string = func.__doc__.format(*normalized_args)

        def temp_func():
            pass

        temp_func.__doc__ = doc_string
        docgen.generate_doc("/plugins/{}/{}".format(self.plugin_info.name, path), temp_func)

    # default APIs have to be registered manually like this
    def set_all_default_docs(self):
        self.set_default_docs(self.get_connection, "connection", type(self.plugin_info.connection))
        self.set_default_docs(self.list_connections, "connections", [type(self.plugin_info.connection)])
