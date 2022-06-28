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
import React, { useEffect, useRef } from 'react'
import {
  Providers,
  ProviderLabels,
  ProviderIcons
} from '@/data/Providers'
import {
  Icon,
  Colors,
  Position,
  Popover,
  TextArea,
  Button,
  H3,
  Classes
} from '@blueprintjs/core'
import dayjs from '@/utils/time'

const StageTaskName = (props) => {
  const {
    task,
    showDetails = null,
    onClose = () => {}
  } = props

  const popoverTriggerRef = useRef()

  useEffect(() => {
    if (showDetails !== null && popoverTriggerRef.current) {
      popoverTriggerRef.current.click()
    }
  }, [showDetails])

  return (
    <>
      <Popover
        className='trigger-pipeline-activity-help'
        popoverClassName='popover-help-pipeline-activity'
        // isOpen={showDetails && showDetails.ID === task.ID}
        onClosed={onClose}
        position={Position.RIGHT}
        autoFocus={false}
        enforceFocus={false}
        usePortal={true}
      >
        <span className='task-plugin-text' ref={popoverTriggerRef}>{task.plugin}</span>
        <>
          <div style={{ textShadow: 'none', fontSize: '12px', padding: '12px', maxWidth: '400px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <div style={{
                marginBottom: '10px',
                color: Colors.GRAY2,
                fontWeight: 700,
                fontSize: '14px',
                maxWidth: '60%'
              }}
              >
                <H3 style={{
                  margin: 0,
                  fontSize: '18px',
                  color: Colors.BLACK,
                  textOverflow: 'ellipsis',
                  overflow: 'hidden',
                  whiteSpace: 'nowrap',
                }}
                >
                  {task.plugin === Providers.REFDIFF && (<>{ProviderLabels.REFDIFF}</>)}
                  {task.plugin === Providers.GITEXTRACTOR && (<>{ProviderLabels.GITEXTRACTOR}</>)}
                  {task.plugin === Providers.FEISHU && (<>{ProviderLabels.FEISHU}</>)}
                  {task.plugin === Providers.JENKINS && (<>{ProviderLabels.JENKINS}</>)}
                  {(task.plugin === Providers.GITLAB || task.plugin === Providers.JIRA) && (<>ID {task.options.projectId || task.options.boardId}</>)}
                  {task.plugin === Providers.GITHUB && task.plugin !== Providers.JENKINS && (<>@{task.options.owner}/{task.options.repositoryName}</>)}
                </H3>
                {![Providers.JENKINS, Providers.REFDIFF, Providers.GITEXTRACTOR].includes(task.plugin) && (
                  <>{ProviderLabels[task.plugin.toUpperCase()] || 'System Task'}<br /></>
                )}
              </div>
              <div style={{
                fontWeight: 800,
                displays: 'flex',
                alignItems: 'center',
                justifyContent: 'flex-start',
                alignSelf: 'flex-start',
                padding: '0 0 0 40px',
                fontSize: '18px',
                marginLeft: 'auto'
              }}
              >
                {Number(task.status === 'TASK_COMPLETED' ? 100 : (task.progress / 1) * 100).toFixed(0)}%
              </div>
              <div style={{ padding: '0 0 10px 20px' }}>
                {ProviderIcons[task.plugin.toLowerCase()](32, 32)}
              </div>
            </div>
            {task.status === 'TASK_CREATED' && (
              <div style={{ fontSize: '10px' }}>
                <p style={{ fontSize: '14px' }}>
                  This task (ID #{task.ID}) is <strong>PENDING</strong> and has not yet started.
                </p>
                <strong>Created Date &mdash;</strong> <span>{dayjs(task.CreatedAt).format('L LT')}</span>
              </div>
            )}
            {task.status !== 'TASK_CREATED' && (
              <div style={{ fontSize: '10px' }}>
                <div style={{ display: 'flex', flexDirection: 'row', justifyContent: 'space-between' }}>
                  <div>
                    <label style={{ color: Colors.GRAY2 }}>ID</label><br />
                    <span>{task.ID}</span>
                  </div>
                  <div style={{ marginLeft: '20px' }}>
                    <label style={{ color: Colors.GRAY2 }}>Created</label><br />
                    <span>{dayjs(task.CreatedAt).format('L LT')}</span>
                  </div>
                  {task.finishedAt && (
                    <div style={{ marginLeft: '20px' }}>
                      <label style={{ color: Colors.GRAY2 }}>Finished</label><br />
                      <span>{dayjs(task.finishedAt).format('L LT')}</span>
                    </div>
                  )}
                </div>
                <div style={{ marginTop: '6px' }}>
                  <label style={{ color: Colors.GRAY2 }}>Status</label><br />
                  <strong>{task.status}</strong>{' '}
                  <strong
                    className='bp3-tag'
                    style={{ minHeight: '16px', fontSize: '10px', padding: '2px 6px', borderRadius: '6px' }}
                  >PROGRESS {task.progress * 100}%
                  </strong>
                </div>
                <div style={{ marginTop: '6px' }}>
                  <label style={{ color: Colors.GRAY2 }}>Options</label><br />
                  <TextArea
                    readOnly
                    fill value={JSON.stringify(task.options)} style={{ fontSize: '10px', backgroundColor: '#f8f8f8', resize: 'none' }}
                  />
                  {/* <span>
                        <pre style={{ margin: 0 }}>
                          <code>
                            {JSON.stringify(task.options)}
                          </code>
                        </pre>
                      </span> */}
                </div>
                {task.message !== '' && (
                  <div style={{ marginTop: '6px' }}>
                    <label style={{ color: Colors.DARK_GRAY1 }}>Message</label><br />
                    <span style={{ color: task.status === 'TASK_FAILED' ? Colors.RED3 : Colors.BLACK }}>
                      {task.status === 'TASK_FAILED' && (
                        <Icon
                          icon='warning-sign'
                          color={Colors.RED5}
                          size={10}
                          style={{ marginRight: '3px' }}
                        />
                      )}
                      {task.message}
                    </span>
                  </div>
                )}
                <div style={{ marginTop: '6px' }}>
                  <label style={{ color: Colors.GRAY2 }}>Updated</label><br />
                  <span>{dayjs(task.UpdatedAt).format('L LT')}</span>
                </div>
              </div>
            )}
            <div style={{ marginTop: '10px', display: 'flex', justifyContent: 'flex-end' }}>
              <Button className={Classes.POPOVER_DISMISS} text='OK' intent='primary' small />
            </div>
          </div>
        </>
      </Popover>
    </>
  )
}

export default StageTaskName
