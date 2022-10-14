from abc import abstractmethod
import json
from typing import List, TypeVar, Iterable, Type, Tuple
from datetime import datetime

import typer
from sqlmodel import SQLModel, create_engine
import sqlalchemy as sql

from pydevlake.tasks import Task, Collector, Extractor, Convertor, SubstreamCollector
from pydevlake.model import ToolModel, DomainModel
from pydevlake.logging import logger


class Plugin:
    def __init__(self, **conf):
        self.conf = conf
        db_url = conf.get('db_url')
        if not db_url:
            raise Exception('Missing db_url conf parameter')

        self.engine = create_engine(db_url)
        self._streams = {s.name: s for s in self.streams}
        
        SQLModel.metadata.create_all(self.engine)

    @property
    def name(self) -> str:
        """
        The name of the plugin, defaults to the class name lowercased.
        """
        return type(self).__name__.lower().removesuffix('plugin')

    @property
    def tasks(self) -> List[Task]:
        return [task for stream in self.streams for task in stream.tasks]

    @property
    def streams(self) -> List['Stream']:
        pass

    def collect(self, stream: str):
        logger.info(f'start collecting {stream}')
        self.get_stream(stream).collector.run(self.engine)

    def extract(self, stream: str):
        logger.info(f'start extracting {stream}')
        self.get_stream(stream).extractor.run(self.engine)

    def convert(self, stream: str):
        logger.info(f'start converting {stream}')
        self.get_stream(stream).convertor.run(self.engine)

    def get_stream(self, stream: str):
        stream = self._streams.get(stream)
        if stream is None:
            raise Exception(f'Unkown stream {stream}')
        return stream

    def list_subtasks(self):
        for task in self.tasks:
            data = {
                'name': task.name,
                'description': task.description,
                'command': f'{task.verb} {task.stream.name}'
            }
            print(json.dumps(data))

    @classmethod
    def cli(cls):
        app = typer.Typer()

        @app.command()
        def collect(stream: str, conf: str=None):
            conf_dict: dict = json.loads(conf)
            cls(**conf_dict).collect(stream)
        
        @app.command()
        def extract(stream: str, conf: str=None):
            conf_dict: dict = json.loads(conf)
            cls(**conf_dict).extract(stream)

        @app.command()
        def convert(stream: str, conf: str=None):
            conf_dict: dict = json.loads(conf)
            cls(**conf_dict).convert(stream)

        @app.command()
        def list_subtasks():
            cls(**conf).list_subtasks()

        return app()


# TODO: rename to resource?
class Stream:
    def __init__(self, plugin: Plugin):
        self.plugin = plugin
        self.collector = Collector(self)
        self.extractor = Extractor(self)
        self.convertor = Convertor(self)
        
        metadata = sql.MetaData()

        self.raw_table = sql.Table(
            f'_raw_{self.qualified_name}',
            metadata,
            sql.Column('id', sql.Integer(), primary_key=True, autoincrement=True),
            sql.Column('params', sql.String()),
            sql.Column('data', sql.LargeBinary()),
            sql.Column('url', sql.String()),
            sql.Column('input', sql.LargeBinary(), default=bytes),
            sql.Column('created_at', sql.Date(), default=datetime.now)
        )

        metadata.create_all(self.plugin.engine)

    @property
    def name(self):
        return type(self).__name__.lower()

    @property
    def qualified_name(self):
        return f'{self.plugin.name}_{self.name}'

    @property
    def conf(self):
        return self.plugin.conf

    def tool_model(self) -> Type[ToolModel]:
        pass

    def domain_model(self) -> Type[DomainModel]:
        pass

    def collect(self, state) -> Iterable[Tuple[dict, dict]]:
        pass

    def extract(self, raw_data: dict) -> ToolModel:
        return self.tool_model(**raw_data) 

    def convert(self, tool_model: ToolModel) -> DomainModel:
        pass

    @property
    def tasks(self):
        return [self.collector, self.extractor, self.convertor]


class Substream(Stream):
    def __init__(self, plugin: Plugin):
        super().__init__(plugin)
        self.collector = SubstreamCollector(self)

    @property
    @abstractmethod
    def parent_stream(self):
        pass


# def get_return_type(func: function):
#     ret_type = func.__annotations__.get('return')
#     if not ret_type:
#         raise Exception(f'{function} has no return type annotation')
#     return ret_type


# def item_type(self):
#     ret_type = get_return_type(self.collect)
#     return ret_type


# def is_generic(type_):
#     return isinstance(type_, _GenericAlias)

# def get_type_arg(generic_type):
#     if not is_generic(generic_type):
#         raise Exception(f'{generic_type} is not a generic type')
    