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


from typing import List, Literal, Union
from datetime import datetime

from pydantic import BaseModel, Field, Literal

from pydevlake.message import DynamicModelInfo
from pydevlake.model import ToolTable


class Execute(BaseModel):
    type: Literal["execute"]
    sql: str


class AutoMigrate(BaseModel):
    type: Literal["auto_migrate"]
    dynamic_model_info: DynamicModelInfo


Operation = Union[Execute, AutoMigrate]


class MigrationScript:
    script: List[Operation] = Field(..., discriminator="type")
    version: int
    name: str


class MigrationScriptBuilder:
    def __init__(self):
        self.operations = []

    def execute(self, sql: str):
        self.operations.append(Execute(sql=sql))

    def auto_migrate(self, table: ToolTable):
        dynamic_model_info = DynamicModelInfo.from_model(table)
        self.operations.append(AutoMigrate(dynamic_model_info=dynamic_model_info))


def migration(version: int):
    """
    Builds a migration script from a function.

    Usage:

    @migration(20230511)
    def change_description_type(op: MigrationScriptBuilder):
        op.exec('ALTER TABLE my_table ALTER COLUMN description TYPE text')
    """
    _validate_version(version)

    def wrapper(fn):
        builder = MigrationScriptBuilder()
        fn(builder)
        return MigrationScript(builder.operations, version, fn.__name__)
    return wrapper


def _validate_version(version: int):
    str_version = str(version)
    err = ValueError(f"Invalid version {version}, must be in YYYYMMDD format")
    if len(str_version) != 8:
        raise err
    try:
        datetime.strptime(str_version, "%Y%m%d")
    except ValueError:
        raise  err
