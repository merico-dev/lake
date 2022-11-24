from abc import abstractmethod
import json
from datetime import datetime
from typing import Tuple, Dict, Iterable, Optional


import sqlalchemy.sql as sql
from sqlmodel import Session, SQLModel, Field, select

from pydevlake.model import ToolModel, DomainEntity, generate_domain_id
from pydevlake import logger


class Task:
    def __init__(self, stream):
        self.stream = stream

    @property
    def plugin(self):
        return self.stream.plugin

    @property
    def conf(self):
        return self.plugin.conf

    @property
    def name(self):
        return f'{self.verb.lower()}{self.plugin.name.capitalize()}{self.stream.name.capitalize()}'

    @property
    def description(self):
        return f'{self.verb.capitalize()} {self.plugin.name} {self.stream.name.lower()}'

    @property
    def verb(self) -> str:
        pass

    def run(self, engine, sync_point_interval=100):
        with Session(engine) as session:
            task_run = self._start_task(session)

            state = self._get_last_state(session)
            
            try:
                for i, (data, state) in enumerate(self.fetch(state, session)):
                    if i % sync_point_interval == 0:
                        # Save current state
                        task_run.state = json.dumps(state)
                        session.merge(task_run)
                    self.process(data, session)
            except Exception as e:
                logger.error(e)
                return

            task_run.state = json.dumps(state)
            task_run.completed = datetime.now()
            session.merge(task_run)
            session.commit()
    
    def _start_task(self, session):
        task_run = TaskRun(
            task_name=self.name, 
            connection_id=self.conf['connection_id'], 
            started=datetime.now(),
            state=json.dumps({})
        )
        session.add(task_run)
        return task_run

    @abstractmethod
    def fetch(self, state: Dict, session: Session) -> Iterable[Tuple[object, Dict]]:
        """
        Queries the data source and returns an iterable of (data, state) tuples.
        The `data` can be any object.
        The `state` is a dict with str keys.
        `Fetch` is called with the last state of the last run of this task.
        """
        pass

    @abstractmethod
    def process(self, data: object, session: Session):
        """
        Called for all data entries returned by `fetch`.
        """
        pass

    def _get_last_state(self, session):
        stmt = (
            select(TaskRun)
            .where(TaskRun.task_name == self.name)
            .where(TaskRun.connection_id == self.conf["connection_id"])
            .where(TaskRun.completed != None)
            .order_by(TaskRun.started)
        )
        task_run = session.exec(stmt).first()
        if task_run is not None:
            return json.loads(task_run.state)
        return {}


class TaskRun(SQLModel, table=True):
    """
    Table storing information about the execution of tasks.

    #TODO: rework id uniqueness:
    # - see and unify with _devlake_tasks table on go side
    # - or sync with Keon about the table he created for Singer MR
    """
    id: Optional[int] = Field(primary_key=True) 
    task_name: str
    connection_id: int
    started: datetime
    completed: Optional[datetime]
    state: str # JSON encoded dict of atomic values
    

class Collector(Task):
    @property
    def verb(self):
        return 'collect'

    def fetch(self, state: Dict, _) -> Iterable[Tuple[object, Dict]]:
        return self.stream.collect(state)

    def process(self, data: object, session: Session):
        session.exec(
            sql.insert(self.stream.raw_table).values(
                params=b'',
                data=json.dumps(data).encode('utf8'),
                url=''
            )
        )


class CustomCollector(Collector):
    def fetch(self, state: Dict, _) -> Iterable[Tuple[object, Dict]]:
        return self.stream.collect(state)


class SubstreamCollector(Collector):
    def fetch(self, state: Dict, session):
        for parent in session.exec(sql.select(self.stream.parent_stream.tool_model)).scalars():
            yield from self.stream.collect(state, parent)


class Extractor(Task):
    @property
    def verb(self):
        return 'extract'

    def fetch(self, state: Dict, session: Session) -> Iterable[Tuple[object, dict]]:
        table = self.stream.raw_table
        for raw in session.execute(select(table.c.data)).scalars():
            yield json.loads(raw), state

    def process(self, data: dict, session: Session):
        tool_model = self.stream.extract(data)
        session.merge(tool_model)

        
class Convertor(Task):
    @property
    def verb(self):
        return 'convert'

    def fetch(self, state: Dict, session: Session) -> Iterable[Tuple[ToolModel, Dict]]:
        for item in session.exec(select(self.stream.tool_model)):
            yield item, state

    def process(self, tool_model: ToolModel, session: Session):
        for domain_model in self.stream.convert(tool_model):
            if isinstance(domain_model, DomainEntity):
                domain_model.id = generate_domain_id(tool_model)
            session.merge(domain_model)
