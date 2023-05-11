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


from typing import Iterable, Type
from abc import abstractmethod
from enum import Enum

from pydevlake.context import Context
from pydevlake.subtasks import Collector, Extractor, Convertor, SubstreamCollector
from pydevlake.model import RawModel, ToolModel, ToolScope, DomainModel
from pydevlake.extractor import autoextract


class DomainType(Enum):
    CODE = "CODE"
    TICKET = "TICKET"
    CODE_REVIEW = "CODEREVIEW"
    CROSS = "CROSS"
    CICD = "CICD"
    CODE_QUALITY = "CODEQUALITY"


class Stream:
    def __init__(self, plugin_name: str):
        self.plugin_name = plugin_name
        self.collector = Collector(self)
        self.extractor = Extractor(self)
        self.convertor = Convertor(self)
        self._raw_model = None

    @property
    def subtasks(self):
        return [self.collector, self.extractor, self.convertor]

    @property
    def name(self):
        return type(self).__name__.lower()

    @property
    def qualified_name(self):
        return f'{self.plugin_name}_{self.name}'

    @property
    def tool_model(self) -> Type[ToolModel]:
        pass

    @property
    @abstractmethod
    def domain_types(self) -> list[DomainType]:
        pass

    def raw_model(self, session) -> Type[RawModel]:
        if self._raw_model is not None:
            return self._raw_model

        table_name = self.raw_model_table

        # Look for existing raw model
        for mapper in RawModel._sa_registry.mappers:
            model = mapper.class_
            if model.__tablename__ == table_name:
                self._raw_model = model
                return self._raw_model

        # Create raw model
        class StreamRawModel(RawModel, table=True):
            __tablename__ = table_name

        self._raw_model = StreamRawModel
        RawModel.metadata.create_all(session.get_bind())
        return self._raw_model

    @property
    def raw_model_table(self):
        return f'_raw_{self.plugin_name}_{self.name}'

    def collect(self, state, context) -> Iterable[tuple[object, dict]]:
        pass

    def extract(self, raw_data: dict) -> ToolModel:
        return autoextract(raw_data, self.tool_model)

    def convert(self, tool_model: ToolModel, context: Context) -> DomainModel:
        pass

    def should_run_on(self, scope: ToolScope) -> bool:
        return True


class Substream(Stream):
    def __init__(self, plugin_name: str):
        super().__init__(plugin_name)
        self.collector = SubstreamCollector(self)

    @property
    @abstractmethod
    def parent_stream(self):
        pass

    def collect(self, state, context, parent) -> Iterable[tuple[object, dict]]:
        pass
