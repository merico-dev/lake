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
        return f'{self.verb().lower()}{self.plugin.name.capitalize()}{self.stream.name.capitalize()}'

    @property
    def description(self):
        return f'{self.verb().capitalize()} {self.plugin.name} {self.stream.name.lower()}'

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
                raise e

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
            select([TaskRun.state, sql.func.max(TaskRun.completed)])
            .where(TaskRun.task_name == self.name)
            .where(TaskRun.connection_id == 1)
            # TODO: add connection_id and params
            .order_by(TaskRun.started)
        )
        task_run = session.exec(stmt).one()
        return json.loads(task_run)


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
    def verb(self):
        return 'collect'

    def fetch(self, state: Dict, _) -> Iterable[Tuple[object, Dict]]:
        return self.stream.collect(state)

    # def fetch(self, state: Dict, _) -> Iterable[Tuple[object, Dict]]:
    #     last_sync_point = state.get('last_sync_point')
    #     if last_sync_point:
    #         hook = IncrementalUpdateHook(self.apply_sync_point(last_sync_point), self.get_sync_point)
    #         with self.inc_hook(self.apply_sync_point(state[last_sync_point])):
                

    # def apply_sync_point(self, request: Request, last: object) -> Optional[Request]:
    #     """
    #     Modify a request, e.g. by adding a query parameter, 
    #     (or return a new request) to apply a filter so that the response
    #     contains only items after the last backup point.
    #     """
    #     pass


    # def get_sync_point(self, response: Response) -> object:
    #     """
    #     Extract the checkpoint from a response object
    #     """
    #     pass

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
    def verb(self):
        return 'extract'

    def fetch(self, state: Dict, session: Session) -> Iterable[Tuple[object, dict]]:
        table = self.stream.raw_table
        for raw in session.execute(select(table.c.data)).scalars():
            yield json.loads(raw), state

    def process(self, data: dict, session: Session):
        tool_model = self.stream.extract(data)
        session.add(tool_model)

        
class Convertor(Task):
    def verb(self):
        return 'convert'

    def fetch(self, state: Dict, session: Session) -> Iterable[Tuple[ToolModel, Dict]]:
        for item in session.exec(select(self.stream.tool_model)):
            yield item, state

    def process(self, tool_model: ToolModel, session: Session):
        for domain_model in self.stream.convert(tool_model):
            if isinstance(domain_model, DomainEntity):
                domain_model.id = generate_domain_id(tool_model)
            session.add(domain_model)
