from typing import Optional
from datetime import datetime
from enum import Enum

from sqlmodel import Field, Relationship

from pydevlake.model import DomainEntity, NoPKModel


class CICDPipeline(DomainEntity, table=True):
    __table_name__ = 'cicd_pipelines'

    class Result(Enum):
        SUCCESS = "SUCCESS"
        FAILURE = "FAILURE"
        ABORT = "ABORT"
        MANUAL = "MANUAL"

    class Status(Enum):
        IN_PROGRESS = "IN_PROGRESS"
        DONE = "DONE"
        MANUAL = "MANUAL"

    class Type(Enum):
        CI = "CI"
        CD = "CD"
        
    name: str
    status: Status
    created_date: datetime
    finished_date: Optional[datetime]
    result: Optional[Result]
    duration_sec: Optional[int]
    environment: Optional[str]
    type: Optional[Type] #Unused

    # parent_pipelines: list["CICDPipeline"] = Relationship(back_populates="child_pipelines", link_model="CICDPipelineRelationship")
    # child_pipelines: list["CICDPipeline"] = Relationship(back_populates="parent_pipelines", link_model="CICDPipelineRelationship")


class CICDPipelineRelationship(NoPKModel):
    __table_name__ = 'cicd_pipeline_relationships'
    parent_pipeline_id: str = Field(primary_key=True, foreign_key=CICDPipeline.id)
    child_pipeline_id: str = Field(primary_key=True, foreign_key=CICDPipeline.id)
