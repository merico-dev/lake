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


from typing import Optional

from pydantic import BaseModel, Field
import jsonref

from pydevlake.model import ToolScope


class Message(BaseModel):
    class Config:
        allow_population_by_field_name = True


class SubtaskMeta(BaseModel):
    name: str
    entry_point_name: str
    required: bool
    enabled_by_default: bool
    description: str
    domain_types: list[str]
    arguments: list[str] = None


class DynamicModelInfo(Message):
    json_schema: dict
    table_name: str

    @staticmethod
    def from_model(model_class):
        schema = model_class.schema()
        if 'definitions' in schema:
            # Replace $ref with actual schema
            schema = jsonref.replace_refs(schema, proxies=False)
            del schema['definitions']
        return DynamicModelInfo(
            json_schema=schema,
            table_name=model_class.__tablename__
        )


class PluginInfo(Message):
    name: str
    description: str
    connection_model_info: DynamicModelInfo
    transformation_rule_model_info: Optional[DynamicModelInfo]
    scope_model_info: DynamicModelInfo
    plugin_path: str
    subtask_metas: list[SubtaskMeta]
    extension: str = "datasource"
    type: str = "python-poetry"
    tables: list[str]


class RemoteProgress(Message):
    increment: int = 0
    current: int = 0
    total: int = 0


class PipelineTask(Message):
    plugin: str
    skip_on_fail: bool = Field(default=False, alias="skipOnFail")
    subtasks: list[str] = Field(default_factory=list)
    options: dict[str, object] = Field(default_factory=dict)


class DynamicDomainScope(Message):
	type_name: str
	data: bytes


class PipelineData(Message):
    plan: list[list[PipelineTask]]
    scopes: list[DynamicDomainScope]


class RemoteScopeTreeNode(Message):
    id: str
    name: str


class RemoteScopeGroup(RemoteScopeTreeNode):
    type: str = Field("group", const=True)


class RemoteScope(RemoteScopeTreeNode):
    type: str = Field("scope", const=True)
    parent_id: str = Field(..., alias="parentId")
    data: ToolScope


class RemoteScopes(Message):
    __root__: list[RemoteScopeTreeNode]
