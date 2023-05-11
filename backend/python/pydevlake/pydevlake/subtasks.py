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


from abc import abstractmethod
import json
from datetime import datetime
from typing import Tuple, Dict, Iterable, Optional, Generator


import sqlalchemy.sql as sql
from sqlmodel import Session, SQLModel, Field, select

from pydevlake.model import RawModel, ToolModel, DomainModel
from pydevlake.context import Context
from pydevlake.message import RemoteProgress
from pydevlake import logger


class Subtask:
    def __init__(self, stream):
        self.stream = stream

    @property
    def name(self):
        return f'{self.verb.lower()}{self.stream.plugin_name.capitalize()}{self.stream.name.capitalize()}'

    @property
    def description(self):
        return f'{self.verb.capitalize()} {self.stream.plugin_name} {self.stream.name.lower()}'

    @property
    def verb(self) -> str:
        pass

    def run(self, ctx: Context, sync_point_interval=100):
        with Session(ctx.engine) as session:
            subtask_run = self._start_subtask(session, ctx.connection.id)
            if ctx.incremental:
                state = self._get_last_state(session, ctx.connection.id)
            else:
                self.delete(session, ctx)
                state = dict()

            try:
                records = self.fetch(state, session, ctx)
                progress = last_progress = 0
                for data, state in records:
                    progress += 1
                    self.process(data, session, ctx)
                    if progress % sync_point_interval == 0:
                        # Save current state
                        subtask_run.state = json.dumps(state)
                        session.merge(subtask_run)
                        session.commit()
                        # Send progress
                        yield RemoteProgress(
                            increment=sync_point_interval,
                            current=progress
                        )
                        last_progress = progress
                # Send final progress
                if progress != last_progress:
                    yield RemoteProgress(
                        increment=progress-last_progress,
                        current=progress
                    )
            except Exception as e:
                logger.error(f'{type(e).__name__}: {e}')
                raise e

            subtask_run.state = json.dumps(state)
            subtask_run.completed = datetime.now()
            session.merge(subtask_run)
            session.commit()

    def _start_subtask(self, session, connection_id):
        subtask_run = SubtaskRun(
            subtask_name=self.name,
            connection_id=connection_id,
            started=datetime.now(),
            state=json.dumps({})
        )
        session.add(subtask_run)
        return subtask_run

    @abstractmethod
    def fetch(self, state: Dict, session: Session, ctx: Context) -> Iterable[Tuple[object, Dict]]:
        """
        Queries the data source and returns an iterable of (data, state) tuples.
        The `data` can be any object.
        The `state` is a dict with str keys.
        `Fetch` is called with the last state of the last run of this subtask.
        """
        pass

    @abstractmethod
    def process(self, data: object, session: Session, ctx: Context):
        """
        Called for all data entries returned by `fetch`.
        """
        pass

    def _get_last_state(self, session, connection_id):
        stmt = (
            select(SubtaskRun)
            .where(SubtaskRun.subtask_name == self.name)
            .where(SubtaskRun.connection_id == connection_id)
            .where(SubtaskRun.completed != None)
            .order_by(sql.desc(SubtaskRun.started))
        )
        subtask_run = session.exec(stmt).first()
        if subtask_run is not None:
            return json.loads(subtask_run.state)
        return {}

    def _params(self, ctx: Context) -> str:
        return json.dumps({
            "connection_id": ctx.connection.id,
            "scope_id": ctx.scope.id
        }, separators=(',', ':'))

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


class Collector(Subtask):
    @property
    def verb(self):
        return 'collect'

    def fetch(self, state: Dict, _, ctx: Context) -> Iterable[Tuple[object, Dict]]:
        return self.stream.collect(state, ctx)

    def process(self, data: object, session: Session, ctx: Context):
        raw_model_class = self.stream.raw_model(session)
        raw_model = raw_model_class(
            params=self._params(ctx),
            data=json.dumps(data).encode('utf8')
        )
        session.add(raw_model)

    def delete(self, session, ctx):
        raw_model = self.stream.raw_model(session)
        stmt = sql.delete(raw_model).where(raw_model.params == self._params(ctx))
        session.execute(stmt)


class SubstreamCollector(Collector):
    def fetch(self, state: Dict, session, ctx: Context):
        for parent in session.exec(sql.select(self.stream.parent_stream.tool_model)).scalars():
            yield from self.stream.collect(state, ctx, parent)


class Extractor(Subtask):
    @property
    def verb(self):
        return 'extract'

    def fetch(self, state: Dict, session: Session, ctx: Context) -> Iterable[Tuple[object, dict]]:
        raw_model = self.stream.raw_model(session)
        query = session.query(raw_model).where(raw_model.params == self._params(ctx))
        for raw in query.all():
            yield raw, state

    def process(self, raw: RawModel, session: Session, ctx: Context):
        tool_model = self.stream.extract(json.loads(raw.data))
        tool_model.set_raw_origin(raw)
        tool_model.connection_id = ctx.connection.id
        session.merge(tool_model)

    def delete(self, session, ctx):
        pass

class Convertor(Subtask):
    @property
    def verb(self):
        return 'convert'

    def fetch(self, state: Dict, session: Session, ctx: Context) -> Iterable[Tuple[ToolModel, Dict]]:
        model = self.stream.tool_model
        query = session.query(model).where(model.raw_data_params == self._params(ctx))
        for item in query.all():
            yield item, state

    def process(self, tool_model: ToolModel, session: Session, ctx: Context):
        res = self.stream.convert(tool_model, ctx)
        if isinstance(res, Generator):
            for each in res:
                self._save(tool_model, each, session, ctx.connection.id)
        else:
            self._save(tool_model, res, session, ctx.connection.id)

    def _save(self, tool_model: ToolModel, domain_model: DomainModel, session: Session, connection_id: int):
        domain_model.set_tool_origin(tool_model)
        if isinstance(domain_model, DomainModel):
            domain_model.id = tool_model.domain_id()
        session.merge(domain_model)

    def delete(self, session, ctx):
        pass
