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
import React from 'react'
import {
  Menu,
  MenuItem,
  Checkbox,
  Intent
} from '@blueprintjs/core'

const RefDiffTasksMenu = (props) => {
  const {
    tasks = [
      { task: 'calculateCommitsDiff', label: 'Calculate Commits Diff' },
      { task: 'calculateIssuesDiff', label: 'Calculate Issues Diff' }],
    selected = [],
    onSelect = () => {}
  } = props
  return (
    <Menu className='tasks-menu refdiff-tasks-menu' minimal='true'>
      <label style={{
        fontSize: '10px',
        fontWeight: 800,
        textTransform: 'uppercase',
        padding: '6px 8px',
        display: 'block'
      }}
      >AVAILABLE PLUGIN TASKS
      </label>
      {tasks.map((t, tIdx) => (
        <MenuItem
          key={`refdiff-item-task-key-${tIdx}`}
          intent={Intent.DANGER}
          minimal='true'
          active={Boolean(selected.includes(t.task))}
          // onClick={(e) => e.preventDefault()}
          data-task={t}
          text={(
            <>
              <Checkbox
                intent={Intent.WARNING}
                checked={Boolean(selected.includes(t.task))}
                label={t.label}
                onChange={(e) => onSelect(e, t)}
              />
            </>
          )}
        />
      ))}
    </Menu>
  )
}

export default RefDiffTasksMenu
