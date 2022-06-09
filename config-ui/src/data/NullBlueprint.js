/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
const BlueprintMode = {
  NORMAL: 'NORMAL',
  ADVANCED: 'ADVANCED'
}

const BlueprintStatus = {
  ENABLED: true,
  DISABLED: false
}

const NullBlueprint = {
  id: null,
  createdAt: null,
  updatedAt: null,
  name: null,
  // Advanced mode uses tasks
  tasks: [
    []
  ],
  // Normal mode uses settings
  settings: {
    version: '1.0',
    connections: []
  },
  cronConfig: '0 0 * * *',
  description: '',
  interval: 'daily',
  enabled: BlueprintStatus.DISABLED,
  mode: BlueprintMode.NORMAL
}

export {
  NullBlueprint,
  BlueprintMode,
  BlueprintStatus
}
