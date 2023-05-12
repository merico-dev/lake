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

from typing import Any, Optional

import pytest
pytest.register_assert_rewrite('pydevlake.testing')

from sqlmodel import Field as _Field


def Field(*args, primary_key: bool=False, source: Optional[str]=None, **kwargs):
    """
    A wrapper around sqlmodel.Field that adds a source parameter.
    """
    schema_extra = kwargs.get('schema_extra', {})
    if source:
        schema_extra['source'] = source
    if primary_key:
        schema_extra['primaryKey'] = True
    return _Field(*args, **kwargs, primary_key=primary_key, schema_extra=schema_extra)


from .model import ToolModel, ToolScope, DomainScope, Connection, TransformationRule, domain_id
from .logger import logger
from .message import RemoteScopeGroup
from .plugin import Plugin, ScopeTxRulePair
from .stream import DomainType, Stream, Substream
from .context import Context

# the debugger hangs on startup during plugin registration (reason unknown), hence this workaround
import sys
if not sys.argv.__contains__('startup'):
    from pydevlake.helpers import debugger
    debugger.init()
