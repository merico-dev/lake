# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


import os
from typing import Iterable, Optional
from inspect import getmodule
from datetime import datetime

import inflect
from pydantic import AnyUrl, SecretStr, validator
from sqlalchemy import Column, DateTime
from sqlalchemy.orm import declared_attr, Session
from sqlalchemy.inspection import inspect
from sqlmodel import SQLModel, Field


inflect_engine = inflect.engine()


class Model(SQLModel):
    id: Optional[int] = Field(primary_key=True)
    created_at: Optional[datetime] = Field(
        sa_column=Column(DateTime(), default=datetime.utcnow)
    )
    updated_at: Optional[datetime] = Field(
        sa_column=Column(DateTime(), default=datetime.utcnow, onupdate=datetime.utcnow)
    )

class ToolTable(SQLModel):
    @declared_attr
    def __tablename__(cls) -> str:
        plugin_name = _get_plugin_name(cls)
        plural_entity = inflect_engine.plural_noun(cls.__name__.lower())
        return f'_tool_{plugin_name}_{plural_entity}'

    class Config:
        allow_population_by_field_name = True
        json_encoders = {
            SecretStr: lambda v: v.get_secret_value() if v else None
        }

        @classmethod
        def alias_generator(cls, attr_name: str) -> str:
            # Allow to set snake_cased attributes with camelCased keyword args.
            # Useful for extractors dealing with raw data that has camelCased attributes.
            parts = attr_name.split('_')
            return parts[0] + ''.join(word.capitalize() for word in parts[1:])

    @classmethod
    def migrate(cls, session: Session):
        """
        Redefine this method to perform migration on this tool model.
        """
        pass


class Connection(ToolTable, Model):
    name: str
    proxy: Optional[AnyUrl]

    @validator('proxy', pre=True)
    def allow_empty_proxy(cls, proxy):
        if proxy == "":
            return None
        return proxy


class TransformationRule(ToolTable, Model):
    name: str


class RawModel(SQLModel):
    id: int = Field(primary_key=True)
    params: str = b''
    data: bytes
    url: str = b''
    input: bytes = b''
    created_at: datetime = Field(default_factory=datetime.now)


class RawDataOrigin(SQLModel):
    # SQLModel doesn't like attributes starting with _
    # so we change the names of the columns.
    raw_data_params: Optional[str] = Field(sa_column_kwargs={'name':'_raw_data_params'})
    raw_data_table: Optional[str] = Field(sa_column_kwargs={'name':'_raw_data_table'})
    raw_data_id: Optional[str] = Field(sa_column_kwargs={'name':'_raw_data_id'})
    raw_data_remark: Optional[str] = Field(sa_column_kwargs={'name':'_raw_data_remark'})

    def set_raw_origin(self, raw: RawModel):
        self.raw_data_id = raw.id
        self.raw_data_params = raw.params
        self.raw_data_table = raw.__tablename__

    def set_tool_origin(self, tool_model: 'ToolModel'):
        self.raw_data_id = tool_model.raw_data_id
        self.raw_data_params = tool_model.raw_data_params
        self.raw_data_table = tool_model.raw_data_table


class NoPKModel(RawDataOrigin):
    created_at: Optional[datetime] = Field(
        sa_column=Column(DateTime(), default=datetime.utcnow)
    )
    updated_at: Optional[datetime] = Field(
        sa_column=Column(DateTime(), default=datetime.utcnow, onupdate=datetime.utcnow)
    )


class ToolModel(ToolTable, NoPKModel):
    connection_id: Optional[int] = Field(primary_key=True)

    def domain_id(self):
        """
        Generate an identifier for domain entities
        originates from self.
        """
        return domain_id(type(self), self.connection_id, *self.primary_keys())

    def primary_keys(self) -> Iterable[object]:
        model_type = type(self)
        mapper = inspect(model_type)
        for primary_key_column in mapper.primary_key:
            prop = mapper.get_property_by_column(primary_key_column)
            if prop.key == 'connection_id':
                continue
            yield getattr(self, prop.key)


class DomainModel(NoPKModel):
    id: Optional[str] = Field(primary_key=True)


class ToolScope(ToolModel):
    id: str = Field(primary_key=True)
    name: str


class DomainScope(DomainModel):
    pass


def domain_id(model_type, connection_id, *args):
    """
    Generate an identifier for domain entities
    originates from a model of type model_type.
    """
    segments = [_get_plugin_name(model_type), model_type.__name__, str(connection_id)]
    segments.extend(str(arg) for arg in args)
    return ':'.join(segments)


def _get_plugin_name(cls):
    """
    Get the plugin name from a class by looking into
    the file path of its module.
    """
    module = getmodule(cls)
    path_segments = module.__file__.split(os.sep)
    # Finds the name of the first enclosing folder
    # that is not a python module
    depth = len(module.__name__.split('.')) + 1
    return path_segments[-depth]


class SubtaskRun(SQLModel, table=True):
    """
    Table storing information about the execution of subtasks.
    """
    id: Optional[int] = Field(primary_key=True)
    subtask_name: str
    connection_id: int
    started: datetime
    completed: Optional[datetime]
    state: str # JSON encoded dict of atomic values
