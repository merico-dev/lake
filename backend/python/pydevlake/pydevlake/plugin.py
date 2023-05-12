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


from typing import Type, Union, Iterable, Optional
from abc import ABC, abstractmethod
from pathlib import Path
import os
import sys

import fire
from sqlmodel import SQLModel, Session
from sqlalchemy.inspection import inspect
from sqlalchemy.engine import Engine

import pydevlake.message as msg
from pydevlake.subtasks import Subtask
from pydevlake.ipc import PluginCommands
from pydevlake.context import Context
from pydevlake.stream import Stream, DomainType
from pydevlake.model import ToolScope, DomainScope, Connection, TransformationRule, SubtaskRun


ScopeTxRulePair = tuple[ToolScope, Optional[TransformationRule]]


class Plugin(ABC):
    def __init__(self):
        self._streams = dict()
        for stream in self.streams:
            if isinstance(stream, type):
                stream = stream(self.name)
            self._streams[stream.name] = stream

    @property
    def name(self) -> str:
        """
        The name of the plugin, defaults to the class name lowercased.
        """
        return type(self).__name__.lower().removesuffix('plugin')

    @property
    def description(self) -> str:
        return f"{self.name} plugin"

    @property
    @abstractmethod
    def connection_type(self) -> Type[Connection]:
        pass

    @property
    @abstractmethod
    def tool_scope_type(self) -> Type[ToolScope]:
        pass

    @property
    def transformation_rule_type(self) -> Type[TransformationRule]:
        return None

    @abstractmethod
    def test_connection(self, connection: Connection):
        """
        Test if the the connection with the datasource can be established with the given connection.
        Must raise an exception if the connection can't be established.
        """
        pass

    @property
    def subtasks(self) -> list[Subtask]:
        return [subtask for stream in self._streams.values() for subtask in stream.subtasks]

    @abstractmethod
    def domain_scopes(self, tool_scope: ToolScope) -> Iterable[DomainScope]:
        pass

    @abstractmethod
    def remote_scopes(self, connection: Connection, group_id: str) -> list[ToolScope]:
        pass

    @abstractmethod
    def remote_scope_groups(self, connection: Connection) -> list[msg.RemoteScopeGroup]:
        pass

    @property
    def streams(self) -> list[Union[Stream, Type[Stream]]]:
        pass

    def collect(self, ctx: Context, stream: str):
        yield from self.get_stream(stream).collector.run(ctx)

    def extract(self, ctx: Context, stream: str):
        yield from self.get_stream(stream).extractor.run(ctx)

    def convert(self, ctx: Context, stream: str):
        yield from self.get_stream(stream).convertor.run(ctx)

    def run_migrations(self, engine: Engine, force: bool = False):
        # NOTE: Not sure what "force" is for
        # TODO: Support migration for transformation rule and connection tables
        # They are currently created on go-side.
        tool_models = [stream.tool_model for stream in self._streams.values()]
        tool_models.append(self.tool_scope_type)
        inspector = inspect(engine)
        tables = SQLModel.metadata.tables
        with Session(engine) as session:
            for model in tool_models:
                if inspector.has_table(model.__tablename__):
                    # TODO: Add version table and migrate if needed
                    model.migrate(session)
                else:
                    tables[model.__tablename__].create(engine)
            session.commit()
        tables[SubtaskRun.__tablename__].create(engine, checkfirst=True)

    def make_remote_scopes(self, connection: Connection, group_id: Optional[str] = None) -> msg.RemoteScopes:
        if group_id:
            remote_scopes = []
            for tool_scope in self.remote_scopes(connection, group_id):
                tool_scope.connection_id = connection.id
                remote_scopes.append(
                    msg.RemoteScope(
                        id=tool_scope.id,
                        parent_id=group_id,
                        name=tool_scope.name,
                        data=tool_scope
                    )
                )
        else:
            remote_scopes = self.remote_scope_groups(connection)
        return msg.RemoteScopes(__root__=remote_scopes)

    def make_pipeline(self, scope_tx_rule_pairs: list[ScopeTxRulePair],
                      entity_types: list[DomainType], connection: Connection):
        """
        Make a simple pipeline using the scopes declared by the plugin.
        """
        plan = self.make_pipeline_plan(scope_tx_rule_pairs, entity_types, connection)
        domain_scopes = []
        for tool_scope, _ in scope_tx_rule_pairs:
            for scope in self.domain_scopes(tool_scope):
                scope.id = tool_scope.domain_id()
                domain_scopes.append(
                    msg.DynamicDomainScope(
                        type_name=type(scope).__name__,
                        data=scope.json(exclude_unset=True)
                    )
                )
        return msg.PipelineData(
            plan=plan,
            scopes=domain_scopes
        )

    def make_pipeline_plan(self, scope_tx_rule_pairs: list[ScopeTxRulePair],
                           entity_types: list[DomainType], connection: Connection) -> list[list[msg.PipelineTask]]:
        """
        Generate a pipeline plan with one stage per scope, plus optional additional stages.
        Redefine `extra_stages` to add stages at the end of this pipeline.
        """
        return [
            *(self.make_pipeline_stage(scope, tx_rule, entity_types, connection) for scope, tx_rule in scope_tx_rule_pairs),
            *self.extra_stages(scope_tx_rule_pairs, entity_types, connection)
        ]

    def extra_stages(self, scope_tx_rule_pairs: list[ScopeTxRulePair],
                     entity_types: list[DomainType], connection: Connection) -> list[list[msg.PipelineTask]]:
        """Override this method to add extra stages to the pipeline plan"""
        return []

    def make_pipeline_stage(self, scope: ToolScope, tx_rule: Optional[TransformationRule],
                            entity_types: list[DomainType], connection: Connection) -> list[msg.PipelineTask]:
        """
        Generate a pipeline stage for the given scope, plus optional additional tasks.
        Subtasks are selected from `entity_types` via `select_subtasks`.
        Redefine `extra_tasks` to add tasks to this stage.
        """
        return [
            msg.PipelineTask(
                plugin=self.name,
                skip_on_fail=False,
                subtasks=self.select_subtasks(scope, entity_types),
                options={
                    "scopeId": scope.id,
                    "scopeName": scope.name,
                    "connectionId": connection.id
                }
            ),
            *self.extra_tasks(scope, tx_rule, entity_types, connection)
        ]

    def extra_tasks(self, scope: ToolScope, tx_rule: Optional[TransformationRule],
                    entity_types: list[DomainType], connection: Connection) -> list[msg.PipelineTask]:
        """Override this method to add tasks to the given scope stage"""
        return []

    def select_subtasks(self, scope: ToolScope, entity_types: list[DomainType]) -> list[str]:
        """
        Returns the list of subtasks names that should be run for given scope and entity types.
        """
        subtasks = []
        for stream in self._streams.values():
            if set(stream.domain_types).intersection(entity_types) and stream.should_run_on(scope):
                for subtask in stream.subtasks:
                    subtasks.append(subtask.name)
        return subtasks

    def get_stream(self, stream_name: str):
        stream = self._streams.get(stream_name)
        if stream is None:
            raise Exception(f'Unknown stream {stream_name}')
        return stream

    def plugin_info(self) -> msg.PluginInfo:
        subtask_metas = [
            msg.SubtaskMeta(
                name=subtask.name,
                entry_point_name=subtask.verb,
                arguments=[subtask.stream.name],
                required=True,
                enabled_by_default=True,
                description=subtask.description,
                domain_types=[dm.value for dm in subtask.stream.domain_types]
            )
            for subtask in self.subtasks
        ]
        if self.transformation_rule_type:
            tx_rule_model_info = msg.DynamicModelInfo.from_model(self.transformation_rule_type)
        else:
            tx_rule_model_info = None
        return msg.PluginInfo(
            name=self.name,
            description=self.description,
            plugin_path=self._plugin_path(),
            extension="datasource",
            connection_model_info=msg.DynamicModelInfo.from_model(self.connection_type),
            transformation_rule_model_info=tx_rule_model_info,
            scope_model_info=msg.DynamicModelInfo.from_model(self.tool_scope_type),
            tool_model_infos=[msg.DynamicModelInfo.from_model(stream.tool_model) for stream in self._streams.values()],
            subtask_metas=subtask_metas
        )

    def _plugin_path(self):
        module_name = type(self).__module__
        module = sys.modules[module_name]
        pluginMainPath = Path(module.__file__)
        run_sh_path = pluginMainPath.parent.parent / "run.sh"
        assert run_sh_path.exists(), f"run.sh not found at {run_sh_path.parent}"
        assert os.access(run_sh_path, os.X_OK), f"run.sh is not executable"
        return str(run_sh_path)

    @classmethod
    def start(cls):
        plugin = cls()
        fire.Fire(PluginCommands(plugin))
