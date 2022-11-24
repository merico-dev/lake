# helper to reduce instantiation boilerplate
from dataclasses import dataclass
from datetime import time
from typing import TypeVar, Generic

from .doc_models import *

T = TypeVar('T')


# no-op decorator for labelling purposes
def shared(cls):
    def wrap(c):
        return c

    return wrap(cls)


@dataclass
class Field(Generic[T]):
    value: T
    tags: str = ""


@dataclass
@shared
class BaseConnection(DocSchema):
    table: str = ""
    Name: str = ""
    ID: int | Field[int] = Field(0, 'gotype:"uint64"')
    CreatedAt: time | Field[time] = Field(time(), 'gotype:"time"')
    UpdatedAt: time | Field[time] = Field(time(), 'gotype:"time"')

    def get_doc_schema(self):
        return BaseConnectionSchema()


@dataclass
@shared
class ApiParamsInput(object):
    Params: dict = None
    Query: dict = None
    Body: object = None
    Request: object = None


@dataclass
@shared
class ApiParamsOutput(object):
    Body: object = None
    Status: int = 200
    File: object = None
    ContentType: str = "application/content-json"


@dataclass
@shared
class ApiEndpoint(object):
    resource: str
    handler: str
    method: str


@dataclass
@shared
class SubtaskMeta(object):
    name: str
    entry_point_name: str
    required: bool
    enabled_by_default: bool
    description: str
    domain_types: list[str]
    arguments: list[str] = None


@dataclass
@shared
class RemoteProgress(object):
    increment: int = 0
    current: int = 0
    total: int = 0


@dataclass
@shared
class SwaggerDoc(object):
    name: str
    resource: str
    schema: bytes


@dataclass
@shared
class PluginInfo(object):
    name: str
    connection: BaseConnection
    plugin_path: str = ""
    api_endpoints: list[ApiEndpoint] = None
    subtask_metas: list[SubtaskMeta] = None
    description: str = ""
    extension: str = "None"
    type: str = "python-poetry"


@dataclass
@shared
class PluginDetails(object):
    plugin_info: PluginInfo
    swagger: SwaggerDoc = None
