from enum import Enum
from datetime import datetime

from sqlmodel import Field, Relationship
from pydevlake.model import ToolModel


class Pipeline(ToolModel, table=True):
    class State(Enum):
        CREATED = "created"
        ERRORED = "errored"
        SETUP_PENDING = "setup-pending"
        SETUP = "setup"
        PENDING = "pending"

    id: str = Field(primary_key=True)
    project_slug: str
    number: str
    state: str
    created_at: datetime
    updated_at: datetime
    # missing attrs:
    #   errors
    #   trigger_parameters
    #   trigger
    #   vcs
    workflows: list["Workflow"] = Relationship(back_populates="pipeline")
    

class Workflow(ToolModel, table=True):
    class Status(Enum):
        SUCCESS = "success"
        RUNNING = "running"
        NOT_RUN = "not_run"
        FAILED = "failed"
        ERROR = "error"
        FAILING = "failing"
        ON_HOLD = "on_hold"
        CANCELED = "canceled"
        UNAUTHORIZED = "unauthorized"

    id: str = Field(primary_key=True)
    pipeline_id: str = Field(foreign_key=Pipeline.id)
    name: str
    project_slug: str
    status: Status
    pipeline_number: int
    created_at: datetime
    stopped_at: datetime
    # canceled_by
    # errored_by
    # tag
    # started_by
    pipeline: Pipeline = Relationship(back_populates="workflows")
