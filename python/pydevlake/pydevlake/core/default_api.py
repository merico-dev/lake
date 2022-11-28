from pydevlake.core.models import PluginInfo, ApiParamsInput, ApiParamsOutput
from .api import PluginAPI
from .doc import normalize_doc_types
from .ipc import plugin_method
from .swagger import docgen


class DefaultPluginAPI(PluginAPI):

    def __init__(self, plugin_info: PluginInfo):
        self.plugin_info = plugin_info
        self.set_all_default_docs()

    @plugin_method()
    def post_connection(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @plugin_method()
    def patch_connection(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @plugin_method()
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

    @plugin_method()
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

    @plugin_method()
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
