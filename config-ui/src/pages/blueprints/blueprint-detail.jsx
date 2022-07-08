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
import React, { Fragment, useEffect, useState, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { CSSTransition } from 'react-transition-group'
import dayjs from '@/utils/time'
import {
  Button,
  Elevation,
  Intent,
  Switch,
  Card,
  Tooltip,
  Icon,
  Tag,
  Colors
} from '@blueprintjs/core'
import { useHistory, useLocation, Link } from 'react-router-dom'
import { NullBlueprint, BlueprintMode } from '@/data/NullBlueprint'

import Nav from '@/components/Nav'
import Sidebar from '@/components/Sidebar'
import Content from '@/components/Content'

import useBlueprintManager from '@/hooks/useBlueprintManager'
import usePipelineManager from '@/hooks/usePipelineManager'
import useConnectionManager from '@/hooks/useConnectionManager'


const TEST_BLUEPRINT = { 
  ...NullBlueprint,
  id: 100,
  name: 'DevLake Daily Blueprint',
  createdAt: new Date().toLocaleString(),
  updatedAt: new Date().toLocaleString()
}

const BlueprintDetail = (props) => {
  const history = useHistory()
  const { bId } = useParams()

  const [blueprintId, setBlueprintId] = useState()
  // @todo: replace with live $blueprint from Hook
  const [activeBlueprint, setActiveBlueprint] = useState(TEST_BLUEPRINT)
  const [blueprintConnections, setBlueprintConnections] = useState([
    {id: 0, dataConnection: 'Merico GitHub', dataScope: 'merico-dev/ake, merico-dev/lake-website', dataEntities: ['code', 'ticket', 'user']},
    {id: 0, dataConnection: 'Merico JIRA', dataScope: 'Sprint Dev Board, DevLake Sync Board ', dataEntities: ['ticket']}
  ])
  const [currentRun, setCurrentRun] = useState({
    status: 'Running',
    startedAt: '7/7/2022, 5:31:33 PM',
    duration: '1 min',
    stage: 'Stage 1',
    tasksCompleted: 5,
    tasksPending: 8
  })
  const [historicalRuns, setHistoricalRuns] = useState([
    {id: 0, status: 'TASK_COMPLETED', statusLabel: 'Completed', statusIcon: <Icon icon='tick-circle' size={14} color={Colors.GREEN5} />, startedAt: '05/25/2022 0:00 AM', completedAt: '05/25/2022 0:15 AM', duration: '15 min' },
    {id: 1, status: 'TASK_COMPLETED', statusLabel: 'Completed', statusIcon: <Icon icon='tick-circle' size={14} color={Colors.GREEN5} />, startedAt: '05/25/2022 0:00 AM', completedAt: '05/25/2022 0:15 AM', duration: '15 min' },
    {id: 2, status: 'TASK_FAILED', statusLabel: 'Failed', statusIcon: <Icon icon='delete' size={14} color={Colors.RED5} />, startedAt: '05/25/2022 0:00 AM', completedAt: '05/25/2022 0:00 AM', duration: '0 min' },
    {id: 3, status: 'TASK_COMPLETED', statusLabel: 'Completed', statusIcon: <Icon icon='tick-circle' size={14} color={Colors.GREEN5} />, startedAt: '05/25/2022 0:00 AM', completedAt: '05/25/2022 0:15 AM', duration: '15 min' },
    {id: 4, status: 'TASK_COMPLETED', statusLabel: 'Completed', statusIcon: <Icon icon='tick-circle' size={14} color={Colors.GREEN5} />, startedAt: '05/25/2022 0:00 AM', completedAt: '05/25/2022 0:15 AM', duration: '15 min' },
    {id: 5, status: 'TASK_FAILED', statusLabel: 'Failed', statusIcon: <Icon icon='delete' size={14} color={Colors.RED5} />,startedAt: '05/25/2022 0:00 AM', completedAt: '05/25/2022 0:00 AM', duration: '0 min' },
  ])

  const {
    // eslint-disable-next-line no-unused-vars
    blueprint,
    blueprints,
    name,
    cronConfig,
    customCronConfig,
    cronPresets,
    tasks,
    detectedProviderTasks,
    enable,
    setName: setBlueprintName,
    setCronConfig,
    setCustomCronConfig,
    setTasks: setBlueprintTasks,
    setDetectedProviderTasks,
    setEnable: setEnableBlueprint,
    isFetching: isFetchingBlueprints,
    isSaving,
    isDeleting,
    createCronExpression: createCron,
    getCronSchedule: getSchedule,
    getCronPreset,
    getCronPresetByConfig,
    getNextRunDate,
    activateBlueprint,
    deactivateBlueprint,
    // eslint-disable-next-line no-unused-vars
    fetchBlueprint,
    fetchAllBlueprints,
    saveBlueprint,
    deleteBlueprint,
    saveComplete,
    deleteComplete
  } = useBlueprintManager()


  useEffect(() => {
    setBlueprintId(bId)
    console.log('>>> REQUESTED BLUEPRINT ID ===', bId)
  }, [bId])

  useEffect(() => {
    if (blueprintId) {
      // @todo: enable blueprint data fetch
      // fetchBlueprint(blueprintId)
    }

  }, [blueprintId, fetchBlueprint])

  const runBlueprint = () => {

  }

  const cancelRun = () => {

  }

  return (
    <>
      <div className="container">
        <Nav />
        <Sidebar />
        <Content>
          <main className="main">
            
            <div className='blueprint-header' style={{ display: 'flex', width: '100%', justifyContent: 'space-between', marginBottom: '10px' }}>
              <div className='blueprint-name' style={{}}>
                <h2 style={{ fontWeight: 'bold' }}>{activeBlueprint?.name}</h2>
              </div>
              <div className='blueprint-info' style={{ display: 'flex', alignItems: 'center' }}>
                <div className='blueprint-schedule'>
                  <span className='blueprint-schedule-interval' style={{ textTransform: 'capitalize', padding: '0 10px'  }}>{activeBlueprint?.interval}</span> &nbsp; {' '}
                  <span className='blueprint-schedule-nextrun'>Next Run in 6 Hours</span>
                </div>
                <div className='blueprint-actions' style={{ padding: '0 10px' }}>
                  <Button intent={Intent.PRIMARY} small text='Run Now' onClick={runBlueprint} />
                </div>
                <div className='blueprint-enabled'>
                  <Switch
                    id='blueprint-enable'
                    name='blueprint-enable'
                    checked={activeBlueprint?.enable}
                    label={activeBlueprint?.enable ? 'Blueprint Enabled' : 'Blueprint Disabled'}
                    // onChange={(e) => toggleBlueprintStatus()}
                    style={{ marginBottom: 0, marginTop: 0 }}
                  />
                </div>
              </div>
            </div>

            <div className='blueprint-connections' style={{ width: '100%', alignSelf: 'flex-start' }}>
              <h3>Overview</h3>
              <Card elevation={Elevation.TWO} style={{ padding: '2px' }}>
              <table className='bp3-html-table bp3-html-table-bordered connections-overview-table' style={{ width: '100%' }}>
                <thead>
                  <tr>
                    <th style={{ minWidth: '200px' }}>Data Connection</th>
                    <th style={{ width: '100%' }}>Data Scope</th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td>
                      Merico GitHub
                    </td>
                    <td>
                      merico-dev/ake, merico-dev/lake-website{' '}
                      <Tag minimal intent={Intent.PRIMARY}>Issue Tracking</Tag>{' '}
                      <Tag minimal intent={Intent.PRIMARY}>Source Code Management</Tag>
                    </td>
                  </tr>
                  <tr>
                    <td>
                      Merico JIRA
                    </td>
                    <td>
                      Sprint Dev Board, DevLake Sync Board{' '}
                      <Tag minimal intent={Intent.PRIMARY}>Issue Tracking</Tag>{' '}
                    </td>
                  </tr>                  
                </tbody>
              </table>
              </Card>
            </div>

            <div className='blueprint-run' style={{ width: '100%', alignSelf: 'flex-start', minWidth: '750px'  }}>
              <h3>Current Run</h3>
              <Card elevation={Elevation.TWO} style={{ padding: '12px', marginBottom: '8px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <div>
                    <label style={{ color: '#7497F7' }}>Status</label>
                    <h4 style={{ margin: 0, padding: 0 }}>{currentRun?.status}</h4>
                  </div>
                  <div>
                    <label style={{ color: '#7497F7' }}>Started at</label>
                    <h4 style={{ margin: 0, padding: 0 }}>{currentRun?.startedAt}</h4>
                  </div>
                  <div>
                    <label style={{ color: '#7497F7' }}>Duration</label>
                    <h4 style={{ margin: 0, padding: 0 }}>{currentRun?.duration}</h4>
                  </div>
                  <div>
                    <label style={{ color: '#7497F7' }}>Current Stage</label>
                    <h4 style={{ margin: 0, padding: 0 }}>{currentRun?.stage}</h4>
                  </div>
                  <div>
                    <label style={{ color: '#7497F7' }}>Tasks Completed</label>
                    <h4 style={{ margin: 0, padding: 0 }}>{currentRun.tasksCompleted} / {currentRun.tasksPending}</h4>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
                    <div  style={{ display: 'block' }}>
                      <Button intent={Intent.PRIMARY} outlined text='Cancel' onClick={cancelRun} />
                    </div>
                  </div>
                </div>
              </Card>
              <Card elevation={Elevation.TWO} style={{ padding: '12px', marginBottom: '8px' }}>
                <div className='blueprint-run-activity' style={{ display: 'flex', width: '100%' }}>
                    <div className='run-stage' style={{ flex: 1, marginRight: '4px' }}>
                      <h3 className='stage-header complete' style={{ margin: '0', padding: '7px', color: '#4DB764', backgroundColor: '#EDFBF0' }}>
                        Stage 1
                      </h3>
                      <div className='task-activity'>

                        
                      </div>
                    </div>
                    <div className='run-stage' style={{ flex: 1, marginLeft: '4px' }}>
                      <h3 className='stage-header active' style={{ margin: '0', padding: '7px', color: '#7497F7', backgroundColor: '#F0F4FE' }}>
                        Stage 2
                      </h3>
                    </div>
                    <div className='run-stage' style={{ flex: 1, marginLeft: '4px' }}>
                      <h3 className='stage-header failed' style={{ margin: '0', padding: '7px', color: '#E34040', backgroundColor: '#FEEFEF' }}>
                        Stage 3
                      </h3>
                    </div>
                    <div className='run-stage' style={{ flex: 1, marginLeft: '4px' }}>
                      <h3 className='stage-header pending' style={{ margin: '0', padding: '7px', color: '#94959F', backgroundColor: '#F9F9FA' }}>
                        Stage 4
                      </h3>
                    </div>
                </div>
              </Card>
            </div>

            <div className='blueprint-historical-runs' style={{ width: '100%', alignSelf: 'flex-start', minWidth: '750px' }}>
              <h3>Historical Runs</h3>
              <Card elevation={Elevation.TWO} style={{ padding: '0', marginBottom: '8px' }}>
                <table className='bp3-html-table bp3-html-table historical-runs-table' style={{ width: '100%' }}>
                  <thead>
                    <tr>
                      <th style={{ minWidth: '100px', whiteSpace: 'nowrap' }}>Status</th>
                      <th style={{ minWidth: '100px', whiteSpace: 'nowrap' }}>Started at</th>
                      <th style={{ minWidth: '100px', whiteSpace: 'nowrap' }}>Completed at</th>
                      <th style={{ minWidth: '100px', whiteSpace: 'nowrap' }}>Duration</th>
                      <th style={{ width: '100%', whiteSpace: 'nowrap' }}></th>
                    </tr>
                  </thead>
                  <tbody>
                    {historicalRuns.map((run, runIdx) => (
                      <tr key={`historical-run-key-${runIdx}`}>
                        <td style={{ width: '15%', whiteSpace: 'nowrap', borderBottom: '1px solid #f0f0f0' }}>{run.statusIcon} {run.statusLabel}</td>
                        <td style={{ width: '25%', whiteSpace: 'nowrap', borderBottom: '1px solid #f0f0f0' }}>{run.startedAt}</td>
                        <td style={{ width: '25%', whiteSpace: 'nowrap', borderBottom: '1px solid #f0f0f0' }}>{run.completedAt}</td>
                        <td style={{ width: '15%', whiteSpace: 'nowrap', borderBottom: '1px solid #f0f0f0' }}>{run.duration}</td>
                        <td style={{ textAlign: 'right', borderBottom: '1px solid #f0f0f0', whiteSpace: 'nowrap' }}>
                          <Tooltip intent={Intent.PRIMARY} content='View JSON'>
                            <Button intent={Intent.PRIMARY} minimal small icon='code' />
                          </Tooltip>
                          <Tooltip intent={Intent.PRIMARY} content='View Full Log'>
                            <Button intent={Intent.PRIMARY} minimal small icon='document' style={{ marginLeft: '10px' }} />
                          </Tooltip>
                          <Button intent={Intent.PRIMARY} minimal small icon='chevron-right' style={{ marginLeft: '10px' }}></Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </Card>
            </div>

          </main>
        </Content>
      </div>
    </>
  )
}

export default BlueprintDetail
