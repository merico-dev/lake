import os
from typing import Optional
from inspect import getmodule

from datetime import datetime
from sqlalchemy import Column, DateTime, func
from sqlalchemy.orm import declared_attr
from sqlalchemy.inspection import inspect
from sqlmodel import SQLModel, Field
import inflect

inflect_engine = inflect.engine()


def get_plugin_name(cls):
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


class RawModel(SQLModel):
    @declared_attr
    def __tablename__(cls) -> str:
        plugin_name = get_plugin_name(cls)
        plural_entity = inflect_engine.plural_noun(cls.__name__.lower())
        return f'_raw_{plugin_name}_{plural_entity}'


class ToolModel(SQLModel):
    @declared_attr
    def __tablename__(cls) -> str:
        plugin_name = get_plugin_name(cls)
        plural_entity = inflect_engine.plural_noun(cls.__name__.lower())
        return f'_tool_{plugin_name}_{plural_entity}'


class DomainModel(SQLModel):
    pass


class RawDataOrigin(SQLModel):
    _raw_data_params: str
    _raw_data_table: str
    _raw_data_id: Optional[str]
    _raw_data_remark: Optional[str]
    

class NoPKModel(RawDataOrigin):
    created_at: datetime = Field(
        sa_column=Column(DateTime(), default=func.now())
    )
    updated_at: datetime = Field(
        sa_column=Column(DateTime(), default=func.now(), onupdate=func.now())
    )


class DomainEntity(NoPKModel):
    id: str = Field(primary_key=True)


def generate_domain_id(tool_model: ToolModel):
    """
    Generate an identifier for a domain entity
    from the tool entity it originates from.
    """
    model_type = type(tool_model)
    segments = [get_plugin_name(model_type), model_type.__name__]
    mapper = inspect(model_type)
    for primary_key_column in mapper.primary_key:
        prop = mapper.get_property_by_column(primary_key_column)
        attr_val = getattr(tool_model, prop.key)
        segments.append(str(attr_val))
    return ':'.join(segments)
