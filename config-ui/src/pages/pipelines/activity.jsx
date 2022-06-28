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
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { CSSTransition } from 'react-transition-group'
import { Link, useHistory, useParams } from 'react-router-dom'
import { GRAFANA_URL } from '@/utils/config'
import dayjs from '@/utils/time'
import {
  Button,
  Card,
  Classes,
  Colors,
  Elevation,
  Icon,
  Intent,
  Popover,
  Position,
  Spinner,
  Tag
} from '@blueprintjs/core'
import { ProviderLabels, Providers } from '@/data/Providers'
import usePipelineManager from '@/hooks/usePipelineManager'
import Nav from '@/components/Nav'
import Sidebar from '@/components/Sidebar'
import AppCrumbs from '@/components/Breadcrumbs'
import Content from '@/components/Content'
import StagePanel from '@/components/pipelines/StagePanel'
import TaskActivity from '@/components/pipelines/TaskActivity'
import CodeInspector from '@/components/pipelines/CodeInspector'
import { ReactComponent as GitlabProviderIcon } from '@/images/integrations/gitlab.svg'
import { ReactComponent as JenkinsProviderIcon } from '@/images/integrations/jenkins.svg'
import { ReactComponent as JiraProviderIcon } from '@/images/integrations/jira.svg'
import { ReactComponent as GitHubProviderIcon } from '@/images/integrations/github.svg'
import { ReactComponent as HelpIcon } from '@/images/help.svg'

// import GitExtractorIcon from '@/images/git.png'
// import RefDiffIcon from '@/images/git-diff.png'
// import AEIcon from '@/images/ae.png'
// import DBTIcon from '@/images/dbt.png'

const PipelineActivity = (props) => {
  const history = useHistory()
  const { pId } = useParams()
  const pollInterval = useRef()

  const [pipelineId, setPipelineId] = useState()
  // const [pipelineName, setPipelineName] = useState()
  const pollTimer = 5000
  const [autoRefresh, setAutoRefresh] = useState(false)

  const [showInspector, setShowInspector] = useState(false)
  const [pipelineReady, setPipelineReady] = useState(false)
  const [stages, setStages] = useState()
  // const [activeStageId, setActiveStageId] = useState(1)

  const {
    // runPipeline,
    cancelPipeline,
    fetchPipeline,
    activePipeline,
    // pipelineRun,
    // isRunning,
    isFetching,
    // errors: pipelineErrors,
    // setSettings: setPipelineSettings,
    // setAutoStart: setPipelineAutoStart,
    lastRunId,
  } = usePipelineManager()

  const pipelineHasProvider = (providerId) => {
    return activePipeline.tasks.some(t => t.plugin === providerId)
  }

  const buildPipelineStages = useCallback((tasks = []) => {
    let stages = {}
    tasks?.forEach(tS => {
      stages = {
        ...stages,
        [tS.pipelineRow]: tasks?.filter(t => t.pipelineRow === tS.pipelineRow)
      }
    })
    console.log('>>> BUILDING PIPELINE STAGES...', stages)
    return stages
  }, [])

  const findActiveStageId = (tasks = []) => {
    const activeTask = tasks.find(t => t.status === 'TASK_RUNNING')
    const failedTask = tasks.find(t => t.status === 'TASK_FAILED')
    const completedTask = tasks.find(t => t.status === 'TASK_COMPLETED')
    return activeTask?.pipelineRow || completedTask?.pipelineRow || failedTask?.pipelineRow || 0
  }

  const restartPipeline = useCallback((tasks = []) => {
    const existingTasksConfiguration = tasks.map(t => {
      return {
        plugin: t.plugin,
        options: t.options,
        pipelineRow: t.pipelineRow,
        pipelineCol: t.pipelineCol
      }
    })
    console.log('>>> RESTARTING PIPELINE WITH EXISTING CONFIGURATION!!', existingTasksConfiguration)
    history.push({
      pathname: '/pipelines/create',
      state: {
        existingTasks: existingTasksConfiguration
      }
    })
  }, [history])

  useEffect(() => {
    setPipelineId(pId)
    console.log('>>> REQUESTED PIPELINE ID ===', pId)
  }, [pId])

  useEffect(() => {
    if (pipelineId) {
      fetchPipeline(pipelineId)
      setAutoRefresh(activePipeline.status === 'TASK_RUNNING')
    }

    return () => {
      clearInterval(pollInterval.current)
    }
  }, [pipelineId, activePipeline.status, fetchPipeline])

  useEffect(() => {
    setPipelineReady(activePipeline.ID !== null && !isFetching)
  }, [activePipeline.ID, isFetching])

  useEffect(() => {
    if (autoRefresh) {
      console.log('>> ACTIVITY POLLING ENABLED!')
      pollInterval.current = setInterval(() => {
        fetchPipeline(pipelineId)
      }, pollTimer)
    } else {
      console.log('>> ACTIVITY POLLING DISABLED!')
      clearInterval(pollInterval.current)
    }
  }, [autoRefresh, fetchPipeline, pipelineId, pollTimer])

  useEffect(() => {
    if (activePipeline.ID) {
      setStages(buildPipelineStages(activePipeline.tasks))
    }
  }, [activePipeline, buildPipelineStages])

  useEffect(() => {

  }, [stages])

  useEffect(() => {
    console.log('>>> GOT LAST RUN ID!', lastRunId)
  }, [lastRunId])

  return (
    <>
      <div className='container container-pipeline-activity'>
        <Nav />
        <Sidebar />
        <Content>
          <main className='main'>
            <AppCrumbs
              items={[
                { href: '/', icon: false, text: 'Dashboard' },
                { href: '/pipelines', icon: false, text: 'Pipelines' },
                { href: `/pipelines/activity/${pipelineId}`, icon: false, text: 'Pipeline Activity & Details', current: true },
              ]}
            />
            <div className='headlineContainer'>
              <Link style={{ display: 'flex', fontSize: '14px', float: 'right', marginLeft: '10px', color: '#777777' }} to='/pipelines'>
                <Icon
                  icon='undo'
                  size={16}
                  style={{ marginRight: '5px', opacity: 0.6 }}
                /> Go Back
              </Link>
              <div style={{ display: 'flex' }}>
                <div>
                  <h1 style={{ margin: 0 }}>
                    Pipeline Activity
                    {activePipeline?.ID && (
                      <>
                        <span style={{ paddingLeft: '10px' }}>&rarr;
                          <strong style={{ paddingLeft: '10px' }}>
                            #{pipelineId}
                          </strong>
                        </span>
                      </>
                    )}
                    <Popover
                      className='trigger-pipeline-activity-help'
                      popoverClassName='popover-help-pipeline-activity'
                      position={Position.RIGHT}
                      autoFocus={false}
                      enforceFocus={false}
                      usePortal={false}
                    >
                      <a href='#' rel='noreferrer'><HelpIcon width={19} height={19} style={{ marginLeft: '10px' }} /></a>
                      <>
                        <div style={{ textShadow: 'none', fontSize: '12px', padding: '12px', maxWidth: '300px' }}>
                          <div style={{ marginBottom: '10px', fontWeight: 700, fontSize: '14px' }}>
                            <Icon icon='help' size={16} /> Pipeline RUN Activity
                          </div>
                          <p>Need Help? &mdash; For better accuracy, ensure that all of your Data Integrations
                            successfully pass the <strong>Connection Test</strong>.
                          </p>
                        </div>
                      </>
                    </Popover>
                  </h1>
                  <p className='page-description mb-0'>View the collection stages for a Pipeline  Run.</p>
                  <p style={{ margin: '0 0 36px 0', padding: 0 }}>
                    You may <strong>Cancel</strong> a running pipeline before it completes.
                  </p>
                </div>
              </div>
            </div>
            {/* (using native loader instead...) */}
            {/* {!autoRefresh && isFetching && (
              <ContentLoader title='Loading Pipeline Run ...' message='Please wait while pipeline activity is loaded.' />
            )} */}
            {activePipeline?.ID && (
              <>
                <StagePanel
                  activePipeline={activePipeline} pipelineReady={pipelineReady}
                  stages={buildPipelineStages(activePipeline.tasks)}
                  activeStageId={findActiveStageId(activePipeline.tasks)}
                  isLoading={isFetching}
                />
                <div style={{ marginBottom: '24px', width: '100%', minWidth: '630px' }}>
                  <CSSTransition
                    in={pipelineReady}
                    timeout={300}
                    classNames='activity-panel'
                    // unmountOnExit
                  >
                    <Card
                      className='pipeline-activity-card'
                      elevation={Elevation.TWO}
                      style={{ padding: 0, width: '100%', display: 'flex', flexDirection: 'column' }}
                    >
                      <div
                        className='pipeline-activity' style={{
                          display: 'flex',
                          width: '100%',
                          justifyContent: 'space-between',
                          borderBottom: '1.0px solid #f0f0f0',
                          padding: '20px'
                        }}
                      >
                        <div
                          className='pipeline-info' style={{
                            paddingRight: '12px',
                            maxWidth: '50%',
                            textOverflow: 'ellipsis',
                            flexGrow: 1
                          }}
                        >
                          <h2 className='headline' style={{ marginTop: '0' }}>
                            <span
                              className='pipeline-name'
                              style={{
                                textOverflow: 'ellipsis',
                                overflow: 'hidden',
                                // whiteSpace: 'nowrap',
                                display: 'block',
                                // maxWidth: '430px',
                                color: activePipeline.status === 'TASK_FAILED' ? Colors.RED4 : ''
                              }}
                            >{activePipeline.name || 'Unamed Collection'}
                            </span>
                          </h2>
                          <div className='pipeline-timestamp'>
                            {dayjs(activePipeline.createdAt).utc().format()} (UTC)
                          </div>
                          {Object.keys(buildPipelineStages(activePipeline.tasks)).length > 1 && (
                            <div className='pipeline-multistage-tag' style={{ padding: '5px 0 0 0' }}>
                              <Icon icon='layers' color={Colors.GRAY4} size={14} style={{ marginRight: '5px' }} />
                              <span style={{
                                fontStyle: 'normal',
                                fontWeight: 900,
                                letterSpacing: '1px',
                                color: '#333',
                                fontSize: '11px'
                              }}
                              >
                                MULTI-STAGE
                              </span>
                            </div>
                          )}

                        </div>
                        <div style={{
                          display: 'flex',
                          justifySelf: 'flex-start',
                          justifyContent: 'space-between',
                          alignItems: 'flex-start',
                          flexGrow: 1
                        }}
                        >
                          <div className='pipeline-status' style={{ paddingRight: '12px' }}>
                            <label style={{ color: Colors.GRAY3 }}>Status</label>
                            <div style={{ fontSize: '15px', display: 'flex' }}>
                              <span style={{ marginRight: '4px', color: activePipeline.status === 'TASK_RUNNING' ? '#0066FF' : '' }}>
                                {activePipeline.status.replace('TASK_', '')}
                              </span>
                              {activePipeline.status === 'TASK_FAILED' && (
                                <Icon
                                  icon='warning-sign' size={16}
                                  color={Colors.RED5} style={{ alignSelf: 'flex-start', marginLeft: '5px', marginBottom: '2px' }}
                                />
                              )}
                              {activePipeline.status === 'TASK_COMPLETED' && (
                                <Icon
                                  icon='tick' size={16}
                                  color={Colors.GREEN5} style={{ alignSelf: 'flex-start', marginLeft: '5px', marginBottom: '2px' }}
                                />
                              )}
                              {activePipeline.status === 'TASK_RUNNING' && (
                                <Spinner
                                  className='pipeline-status-spinner'
                                  size={14}
                                  intent={activePipeline.status === 'TASK_COMPLETED' ? 'success' : 'danger'}
                                  value={activePipeline.status === 'TASK_COMPLETED' ? 1 : null}
                                />
                              )}
                            </div>
                          </div>
                          <div className='pipeline-duration' style={{ paddingRight: '12px' }}>
                            <label style={{ color: Colors.GRAY3 }}>Duration</label>
                            <div style={{ fontSize: '15px', whiteSpace: 'nowrap' }}>
                              {activePipeline.status === 'TASK_RUNNING'
                                ? dayjs(activePipeline.beganAt).toNow(true)
                                : activePipeline.finishedAt == null ? 'N/A' : dayjs(activePipeline.finishedAt).from(activePipeline.beganAt, true)}
                            </div>
                          </div>
                          <div className='pipeline-actions' style={{ display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
                            {activePipeline.status === 'TASK_COMPLETED' && (
                              <a
                                className='bp3-button bp3-intent-primary'
                                href={GRAFANA_URL}
                                target='_blank'
                                rel='noreferrer'
                                style={{ backgroundColor: '#3bd477', color: '#ffffff' }}
                              >
                                <Icon icon='doughnut-chart' size={13} /> <span className='bp3-button-text'>Grafana</span>
                              </a>
                            )}
                            {activePipeline.status === 'TASK_RUNNING' && (
                              <Popover
                                key='popover-help-key-cancel-run'
                                className='trigger-pipeline-cancel'
                                popoverClassName='popover-pipeline-cancel'
                                position={Position.BOTTOM}
                                autoFocus={false}
                                enforceFocus={false}
                                usePortal={true}
                                disabled={activePipeline.status !== 'TASK_RUNNING'}
                              >
                                <Button
                                  className={`btn-cancel-pipeline${activePipeline.status !== 'TASK_RUNNING' ? '-disabled' : ''}`}
                                  icon='stop' text='CANCEL' intent={activePipeline.status !== 'TASK_RUNNING' ? '' : 'primary'}
                                  disabled={activePipeline.status !== 'TASK_RUNNING'}
                                />
                                <>
                                  <div style={{ fontSize: '12px', padding: '12px', maxWidth: '200px' }}>
                                    <p>Are you Sure you want to cancel this <strong>Run</strong>?</p>
                                    <div style={{ display: 'flex', width: '100%', justifyContent: 'flex-end' }}>
                                      <Button
                                        text='NO' minimal
                                        small className={Classes.POPOVER_DISMISS}
                                        style={{ marginLeft: 'auto', marginRight: '3px' }}
                                      />
                                      <Button
                                        className={Classes.POPOVER_DISMISS}
                                        text='YES' icon='small-tick' intent={Intent.DANGER} small
                                        onClick={() => cancelPipeline(activePipeline.ID)}
                                      />
                                    </div>
                                  </div>
                                </>
                              </Popover>
                            )}
                            {activePipeline.status === 'TASK_FAILED' && (
                              <Button
                                className='btn-restart-pipeline'
                                icon='reset'
                                text='RESTART'
                                intent='warning'
                                onClick={() => restartPipeline(activePipeline.tasks)}
                                disabled={activePipeline.status !== 'TASK_FAILED'}
                              />
                            )}
                            {activePipeline.status !== 'TASK_FAILED' && (
                              <Button
                                icon='refresh'
                                style={{ marginLeft: '5px' }}
                                onClick={() => restartPipeline(activePipeline.tasks)}
                                disabled={activePipeline.status === 'TASK_RUNNING' || activePipeline.status === 'TASK_FAILED'}
                                minimal
                              />
                            )}
                          </div>
                        </div>
                      </div>
                      <TaskActivity activePipeline={activePipeline} stages={buildPipelineStages(activePipeline.tasks)} />
                    </Card>
                  </CSSTransition>
                  <div style={{ display: 'flex', padding: '5px 3px', fontSize: '10px', color: '#777777', justifyContent: 'space-between' }}>
                    <div>
                      <Popover
                        className='trigger-pipeline-stage-help'
                        popoverClassName='popover-help-stage-activity'
                        position={Position.BOTTOM}
                        autoFocus={false}
                        enforceFocus={false}
                        usePortal={false}
                      >
                        <a href='#' rel='noreferrer'>
                          <Icon icon='help' color={Colors.GRAY3} size={12} style={{ marginTop: '0', marginRight: '5px' }} />
                        </a>
                        <>
                          <div style={{ textShadow: 'none', fontSize: '12px', padding: '12px', maxWidth: '315px' }}>
                            <div style={{
                              marginBottom: '10px',
                              fontWeight: 700,
                              fontSize: '14px',
                            }}
                            >
                              <Icon icon='help' size={16} /> Stages and Tasks
                            </div>
                            <p>
                              Monitor <strong>Duration</strong> and <strong>Progress</strong> completion for all tasks.
                              <strong>Grafana</strong> access will be  enabled when the pipeline completes.
                            </p>

                          </div>
                        </>
                      </Popover>

                      {activePipeline.finishedTasks}/{activePipeline.totalTasks} Tasks Completed
                      <span style={{ padding: '0 2px' }}>
                        <Icon icon='dot' size={10} color={Colors.GRAY3} style={{ marginBottom: '3px' }} />
                      </span>
                      <strong>Created </strong> {activePipeline.createdAt}
                    </div>
                    <div>
                      {isFetching && (
                        <span style={{ color: Colors.GRAY3 }}>
                          <Icon icon='updated' size={11} style={{ marginBottom: '2px' }} /> Refreshing Activity...
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                {/* <div className='stage-activity' style={{ alignSelf: 'flex-start', width: '100%' }}>
                </div> */}

                <div className='run-settings' style={{ alignSelf: 'flex-start', width: '100%' }}>
                  <div style={{ display: 'flex' }}>
                    <div style={{ display: 'flex', alignItems: 'flex-start', padding: '2px 8px 0 0' }}>
                      <Icon icon='cog' height={16} size={16} color='rgba(0,0,0,0.5)' />
                    </div>
                    <div>
                      <h2 className='headline' style={{ marginTop: 0 }}>
                        Run Settings
                      </h2>
                      <p>Data Provider settings configured for this pipeline execution.</p>
                    </div>
                  </div>

                  <div style={{ padding: '0 10px', display: 'flex', marginTop: '24px', justifyContent: 'space-between', width: '100%' }}>
                    {pipelineHasProvider(Providers.JENKINS) && (
                      <div className='jenkins-settings' style={{ display: 'flex' }}>
                        <div style={{ display: 'flex', padding: '2px 6px' }}>
                          <JenkinsProviderIcon width={24} height={24} />
                        </div>
                        <div>
                          <label style={{ lineHeight: '100%', display: 'block', fontSize: '10px', marginTop: '2px', marginBottom: '10px' }}>
                            <strong style={{
                              fontSize: '16px',
                              fontWeight: 800
                            }}
                            >{ProviderLabels.JENKINS}
                            </strong><br />Auto-configured
                          </label>
                          <span style={{ color: Colors.GRAY3 }}>(No Settings)</span>
                        </div>
                      </div>
                    )}
                    {pipelineHasProvider(Providers.JIRA) && (
                      <div className='jira-settings' style={{ display: 'flex', paddingLeft: '20px' }}>
                        <div style={{ display: 'flex', padding: '2px 6px' }}>
                          <JiraProviderIcon width={24} height={24} />
                        </div>
                        <div>
                          <label style={{ lineHeight: '100%', display: 'block', fontSize: '10px', marginTop: '2px', marginBottom: '10px' }}>
                            <strong style={{
                              fontSize: '16px',
                              fontWeight: 800
                            }}
                            >{ProviderLabels.JIRA}
                            </strong><br />Board IDs
                          </label>
                          {activePipeline.tasks.filter(t => t.plugin === Providers.JIRA).map((t, tIdx) => (
                            <div key={`board-id-key-${tIdx}`}>
                              <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '6px' }} />
                              <span>
                                {t.options[Object.keys(t.options)[0]]} on Server #{t.options[Object.keys(t.options)[1]]}<br />
                              </span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                    {pipelineHasProvider(Providers.GITLAB) && (
                      <div className='gitlab-settings' style={{ display: 'flex', paddingLeft: '20px' }}>
                        <div style={{ display: 'flex', padding: '2px 6px' }}>
                          <GitlabProviderIcon width={24} height={24} />
                        </div>
                        <div>
                          <label style={{ lineHeight: '100%', display: 'block', fontSize: '10px', marginTop: '2px', marginBottom: '10px' }}>
                            <strong style={{
                              fontSize: '16px',
                              fontWeight: 800
                            }}
                            >{ProviderLabels.GITLAB}
                            </strong><br />Project ID
                          </label>
                          {activePipeline.tasks.filter(t => t.plugin === 'gitlab').map((t, tIdx) => (
                            <div key={`project-id-key-${tIdx}`}>
                              <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '6px' }} />
                              <span>
                                {t.options[Object.keys(t.options)[0]]}<br />
                              </span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                    {pipelineHasProvider(Providers.GITHUB) && (
                      <div className='github-settings' style={{ display: 'flex', paddingLeft: '20px', justifySelf: 'flex-start' }}>
                        <div style={{ display: 'flex', padding: '2px 6px' }}>
                          <GitHubProviderIcon width={24} height={24} />
                        </div>
                        <div>
                          <label style={{ lineHeight: '100%', display: 'block', fontSize: '10px', marginTop: '2px', marginBottom: '10px' }}>
                            <strong style={{
                              fontSize: '16px',
                              fontWeight: 800
                            }}
                            >{ProviderLabels.GITHUB}
                            </strong><br />Repositories
                          </label>
                          {activePipeline.tasks.filter(t => t.plugin === Providers.GITHUB).map((t, tIdx) => (
                            <div key={`repostitory-id-key-${tIdx}`}>
                              <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '6px' }} />
                              <span>
                                <strong>{t.options.owner}</strong>
                                <span style={{ color: Colors.GRAY5, padding: '0 1px' }}>/</span>
                                <strong>{t.options.repositoryName || t.options.repo || '(Repository)'}</strong>
                              </span>
                              {t.options.tasks && (
                                <>
                                  <Tag style={{ fontSize: '9px', marginLeft: '5px', backgroundColor: '#eee', color: '#777' }}>+TASKS</Tag>
                                  <ul style={{ fontSize: '9px' }}>
                                    {t.options.tasks.map((oT, otId) => (
                                      <li key={`option-subtask-key${otId}`}>{oT}</li>
                                    ))}
                                  </ul>
                                </>
                              )}
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                    {pipelineHasProvider('gitextractor') && (
                      <div className='gitextractor-settings' style={{ display: 'flex', paddingLeft: '20px', justifySelf: 'flex-start' }}>
                        <div style={{ display: 'flex', padding: '2px 6px' }}>
                          <Icon icon='box' size={24} />
                        </div>
                        <div>
                          <label style={{ lineHeight: '100%', display: 'block', fontSize: '10px', marginTop: '2px', marginBottom: '0px' }}>
                            <strong style={{ fontSize: '16px', fontWeight: 800 }}>GitExtractor</strong><br />Extract Commits &amp; Refs
                          </label>
                          <div style={{ paddingTop: '15px' }}>
                            {activePipeline.tasks.filter(t => t.plugin === 'gitextractor').map((t, tIdx) => (
                              <div key={`gitextractor-opts-key-${tIdx}`}>
                                <div>
                                  <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '6px' }} />
                                  <strong>URL</strong>
                                  <span style={{ color: Colors.GRAY5, padding: '0 1px' }}>: </span>
                                  <span>{t.options.url}</span>
                                </div>
                                <div>
                                  <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '6px' }} />
                                  <strong>RepoId</strong>
                                  <span style={{ color: Colors.GRAY5, padding: '0 1px' }}>: </span>
                                  <span>{t.options.repoId}</span>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    )}
                    {pipelineHasProvider('refdiff') && (
                      <div className='refdiff-settings' style={{ display: 'flex', paddingLeft: '20px', justifySelf: 'flex-start' }}>
                        <div style={{ display: 'flex', padding: '2px 6px' }}>
                          <Icon icon='box' size={24} />
                        </div>
                        <div>
                          <label style={{ lineHeight: '100%', display: 'block', fontSize: '10px', marginTop: '2px', marginBottom: '0px' }}>
                            <strong style={{ fontSize: '16px', fontWeight: 800 }}>RefDiff</strong><br />Release Tag Diffs
                          </label>
                          <div style={{ paddingTop: '15px' }}>
                            {activePipeline.tasks.filter(t => t.plugin === 'refdiff').map((t, tIdx) => (
                              <div key={`gitextractor-opts-key-${tIdx}`}>
                                <div>
                                  <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '6px' }} />
                                  <strong>RepoId</strong>
                                  <span style={{ color: Colors.GRAY5, padding: '0 1px' }}>: </span>
                                  <span>{t.options.repoId}</span>
                                </div>
                                <div>
                                  {t.options.pairs && (
                                    <div>
                                      <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '0' }} /> <Tag style={{ fontSize: '9px', marginLeft: '0', backgroundColor: '#eee', color: '#777' }}>TAG PAIRS</Tag>
                                      <ul style={{ fontSize: '9px' }}>
                                        {t.options.pairs.map((ref, refIdx) => (
                                          <li key={`option-subtask-key${refIdx}`}><strong>old</strong> {ref.oldRef} &nbsp; <strong>new</strong> {ref.newRef} </li>
                                        ))}
                                      </ul>
                                    </div>
                                  )}
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    )}
                    {pipelineHasProvider('ae') && (
                      <div className='ae-settings' style={{ display: 'flex', paddingLeft: '20px', justifySelf: 'flex-start' }}>
                        <div style={{ display: 'flex', padding: '2px 6px' }}>
                          <Icon icon='box' size={24} />
                        </div>
                        <div>
                          <label style={{ lineHeight: '100%', display: 'block', fontSize: '10px', marginTop: '2px', marginBottom: '0px' }}>
                            <strong style={{ fontSize: '16px', fontWeight: 800 }}>Analysis Engine (AE)</strong><br />
                            Merico Enterprise Analysis
                          </label>
                          <div style={{ paddingTop: '15px' }}>
                            {activePipeline.tasks.filter(t => t.plugin === 'ae').map((t, tIdx) => (
                              <div key={`ae-opts-key-${tIdx}`}>
                                <div>
                                  <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '6px' }} />
                                  <strong>Project ID</strong>
                                  <span style={{ color: Colors.GRAY5, padding: '0 1px' }}>: </span>
                                  <span>{t.options.projectId}</span>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    )}
                    {pipelineHasProvider('dbt') && (
                      <div className='dbt-settings' style={{ display: 'flex', paddingLeft: '20px', justifySelf: 'flex-start' }}>
                        <div style={{ display: 'flex', padding: '2px 6px' }}>
                          <Icon icon='box' size={24} />
                        </div>
                        <div>
                          <label style={{ lineHeight: '100%', display: 'block', fontSize: '10px', marginTop: '2px', marginBottom: '0px' }}>
                            <strong style={{ fontSize: '16px', fontWeight: 800 }}>Data Build Tool (DBT)</strong><br />Transform Data with SQL
                          </label>
                          <div style={{ paddingTop: '15px' }}>
                            {activePipeline.tasks.filter(t => t.plugin === 'dbt').map((t, tIdx) => (
                              <div key={`dbt-opts-key-${tIdx}`}>
                                <div>
                                  <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '6px' }} />
                                  <strong>Project Path</strong>
                                  <span style={{ color: Colors.GRAY5, padding: '0 1px' }}>: </span>
                                  <span>{t.options.projectPath}</span>
                                </div>
                                <div>
                                  <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '6px' }} />
                                  <strong>Project Name</strong>
                                  <span style={{ color: Colors.GRAY5, padding: '0 1px' }}>: </span>
                                  <span>{t.options.projectName}</span>
                                </div>
                                <div>
                                  <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '6px' }} />
                                  <strong>Project Target</strong>
                                  <span style={{ color: Colors.GRAY5, padding: '0 1px' }}>: </span>
                                  <span>{t.options.projectTarget}</span>
                                </div>
                                <div>
                                  <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '6px' }} />
                                  <strong>Selected Models</strong>
                                  <span style={{ color: Colors.GRAY5, padding: '0 1px' }}>: </span>
                                  <span>[{t.options.selectedModels?.toString().split(', ')}]</span>
                                </div>
                                <div>
                                  <Icon icon='nest' size={12} color={Colors.GRAY4} style={{ marginRight: '6px' }} />
                                  <strong>Project Vars</strong>
                                  <span style={{ color: Colors.GRAY5, padding: '0 1px' }}>: </span>
                                  <span>{JSON.stringify(t.options.projectVars)}</span>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    )}
                    <div style={{ marginRight: 'auto' }} />
                  </div>
                </div>
              </>
            )}

            {!pipelineId && (
              <Card elevation={Elevation.TWO} style={{ display: 'flex', alignSelf: 'flex-start' }}>
                <div style={{ display: 'flex', alignSelf: 'flex-start', flexDirection: 'column' }}>
                  <h2 style={{ margin: '0 0 12px 0' }}>
                    <Icon
                      icon='warning-sign'
                      color={Colors.RED4} size={16} style={{ marginBottom: '4px' }}
                    /> Pipeline Run ID <strong>Missing</strong>...
                  </h2>
                  <p>Please provide a Pipeline ID to load Run activity and details.
                    <br /> Check the Address URL in your Browser and try again.
                  </p>
                </div>
              </Card>
            )}

            <div style={{
              marginTop: '40px',
              borderTop: '1px solid #f0f0f0',
              display: 'flex',
              width: '100%',
              justifyContent: 'flex-start'
            }}
            >

              <div style={{ padding: '8px', display: 'flex', width: '100%', justifyContent: 'space-between' }}>
                <span>See <strong style={{ textDecoration: 'underline' }}>All Jobs</strong> to monitor all pipeline activity.</span>
                <div>
                  <Button
                    className='btn-inspect-pipeline'
                    onClick={() => setShowInspector((iS) => !iS)}
                    icon='code' text='Inspect JSON' small minimal
                    style={{ marginRight: '3px', color: Colors.GRAY3 }}
                  />
                  <Button
                    onClick={() => history.push('/pipelines/create')}
                    icon='add' text='Create New Run' small minimal
                  />
                </div>
              </div>
            </div>

          </main>

        </Content>

      </div>
      <CodeInspector isOpen={showInspector} activePipeline={activePipeline} onClose={setShowInspector} />
    </>
  )
}

export default PipelineActivity
