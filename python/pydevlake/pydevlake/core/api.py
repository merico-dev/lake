from .ipc import plugin_method
from .models import *


# Must correspond to Golang's default plugin APIs
class PluginAPI(metaclass=abc.ABCMeta):

    @abc.abstractmethod
    @plugin_method
    def test_connection(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @abc.abstractmethod
    @plugin_method
    def post_connection(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @abc.abstractmethod
    @plugin_method
    def patch_connection(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @abc.abstractmethod
    @plugin_method
    def get_connection(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @abc.abstractmethod
    @plugin_method
    def list_connections(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @abc.abstractmethod
    @plugin_method
    def delete_connection(self, ctx, input: ApiParamsInput) -> ApiParamsOutput:
        pass


# Must correspond to Golang's plugin interfaces
class PluginMethods(metaclass=abc.ABCMeta):

    @abc.abstractmethod
    @plugin_method
    def RunMigrations(self, ctx: dict, force: bool):
        pass

    @abc.abstractmethod
    @plugin_method
    def PrepareTaskData(self, ctx: dict, opts: dict):
        pass
