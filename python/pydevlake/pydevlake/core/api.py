from .context import Context
from .models import *


class PluginState(metaclass=abc.ABCMeta):
    @abc.abstractmethod
    def is_cancelled(self) -> bool:
        pass

    @abc.abstractmethod
    def terminate(self) -> bool:
        pass


class PluginAPI(metaclass=abc.ABCMeta):

    @abc.abstractmethod
    def post_connection(self, ctx: Context, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @abc.abstractmethod
    def patch_connection(self, ctx: Context, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @abc.abstractmethod
    def get_connection(self, ctx: Context, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @abc.abstractmethod
    def list_connections(self, ctx: Context, input: ApiParamsInput) -> ApiParamsOutput:
        pass

    @abc.abstractmethod
    def delete_connection(self, ctx: Context, input: ApiParamsInput) -> ApiParamsOutput:
        pass


# Must correspond to Golang's plugin interfaces
class PluginTask(metaclass=abc.ABCMeta):

    @abc.abstractmethod
    def RunMigrations(self, ctx: Context, force: bool):
        pass

    @abc.abstractmethod
    def PrepareTaskData(self, ctx: Context, opts: any):
        pass