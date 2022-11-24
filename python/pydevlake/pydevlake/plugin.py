from abc import abstractmethod
import json
from typing import List, Iterable, Type, Tuple
from datetime import datetime
from urllib.parse import urlparse, parse_qsl

import typer
from sqlmodel import SQLModel, create_engine
import sqlalchemy as sql

from pydevlake.tasks import Task, Collector, Extractor, Convertor, SubstreamCollector
from pydevlake.model import ToolModel, DomainModel
from pydevlake.logger import logger


class Plugin:
    def __init__(self, **conf):
        self.conf = conf
        self._engine = None
        self._streams = {stream.name:stream for stream in self.streams}

    @property
    def engine(self):
        if not self._engine:
            db_url = self.conf.get('db_url')
            if not db_url:
                raise Exception('Missing db_url conf parameter')

            # Extract query args if any
            connect_args = dict(parse_qsl(urlparse(db_url).query))
            db_url = db_url.split('?')[0]
            
            # `parseTime` parameter is not understood by MySQL driver
            if 'parseTime' in connect_args:
                del connect_args['parseTime']

            self._engine = create_engine(db_url, connect_args=connect_args)
            SQLModel.metadata.create_all(self._engine)
        return self._engine

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
        logger.info(f'finished collecting {stream}')

    def extract(self, stream: str):
        logger.info(f'start extracting {stream}')
        self.get_stream(stream).extractor.run(self.engine)
        logger.info(f'finished extracting {stream}')

    def convert(self, stream: str):
        logger.info(f'start converting {stream}')
        self.get_stream(stream).convertor.run(self.engine)
        logger.info(f'finished converting {stream}')

    def get_stream(self, stream: str):
        stream = self._streams.get(stream)
        if stream is None:
            raise Exception(f'Unkown stream {stream}')
        return stream

    def list_subtasks(self):
        for task in self.tasks:
            data = dict(
                name=task.name,
                description=task.description,
                command=[task.verb, task.stream.name],
                domain_model=task.stream.domain_model.__name__
            )
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
            cls().list_subtasks()

        return app()


# TODO: rename to resource?
class Stream:
    def __init__(self, plugin: Plugin):
        self.plugin = plugin
        self.collector = Collector(self)
        self.extractor = Extractor(self)
        self.convertor = Convertor(self)
        self._raw_table = None

    @property
    def raw_table(self):
        if self._raw_table is None:
            metadata = sql.MetaData()

            self._raw_table = sql.Table(
                f'_raw_{self.qualified_name}',
                metadata,
                sql.Column('id', sql.Integer(), primary_key=True, autoincrement=True),
                sql.Column('params', sql.String(256)),
                sql.Column('data', sql.LargeBinary()),
                sql.Column('url', sql.String(256)),
                sql.Column('input', sql.LargeBinary(), default=bytes),
                sql.Column('created_at', sql.Date(), default=datetime.now)
            )

            metadata.create_all(self.plugin.engine)
        return self._raw_table

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
