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
import { useHistory } from 'react-router-dom'
import { GRAFANA_URL } from '@/utils/config'
import dayjs from '@/utils/time'
import {
  Button,
  ButtonGroup,
  Card,
  Classes,
  Colors,
  Elevation,
  Icon,
  Intent,
  Menu,
  MenuItem,
  Popover,
  Position,
  Spinner,
  Tag,
  Tooltip
} from '@blueprintjs/core'
import usePipelineManager from '@/hooks/usePipelineManager'
import Nav from '@/components/Nav'
import Sidebar from '@/components/Sidebar'
import AppCrumbs from '@/components/Breadcrumbs'
import Content from '@/components/Content'
import ContentLoader from '@/components/loaders/ContentLoader'
import PipelineIndicator from '@/components/widgets/PipelineIndicator'
import CodeInspector from '@/components/pipelines/CodeInspector'
import { ReactComponent as HelpIcon } from '@/images/help.svg'

const Pipelines = (props) => {
  const history = useHistory()
  // const { providerId } = useParams()
  // const [activeProvider, setActiveProvider] = useState(integrationsData[0])

  const [isProcessing, setIsProcessing] = useState(false)
  const [refresh, setRefresh] = useState(false)
  const [activeStatus, setActiveStatus] = useState('all')
  // const [latestPipeline, setLatestPipeline] = useState()
  const [showInspector, setShowInspector] = useState(false)
  const [inspectPipeline, setInspectPipeline] = useState(null)

  const {
    // runPipeline,
    cancelPipeline,
    fetchPipeline,
    pipelines,
    pipelineCount,
    fetchAllPipelines,
    activePipeline,
    // pipelineRun,
    // isFetching,
    isFetchingAll,
    // errors: pipelineErrors,
    // setSettings: setPipelineSettings,
    // lastRunId,
  } = usePipelineManager()

  const [filteredPipelines, setFilteredPipelines] = useState([])
  const [pagedPipelines, setPagedPipelines] = useState([])
  const pageOptions = [10, 25, 50, 75, 100]
  const currentPage = useRef(1)
  const [perPage, setPerPage] = useState(pageOptions[0])
  const [maxPage, setMaxPage] = useState(Math.ceil(filteredPipelines.length / perPage))
  // @todo: generate dynamically from $pageOptions
  const pagingOptionsMenu = (
    <Menu>
      <MenuItem active={perPage === 10} icon='key-option' text='10 Records' onClick={() => setPerPage(10)} />
      <MenuItem active={perPage === 25} icon='key-option' text='25 Records' onClick={() => setPerPage(25)} />
      <MenuItem active={perPage === 50} icon='key-option' text='50 Records' onClick={() => setPerPage(50)} />
      <MenuItem active={perPage === 75} icon='key-option' text='75 Records' onClick={() => setPerPage(75)} />
      <MenuItem active={perPage === 100} icon='key-option' text='100 Records' onClick={() => setPerPage(100)} />
    </Menu>
  )
  const nextPage = () => {
    currentPage.current = Math.min(maxPage, currentPage.current + 1)
    setRefresh(r => !r)
    console.log('>>>> NEXT PAGE', currentPage.current)
  }

  const prevPage = () => {
    currentPage.current = Math.max(1, currentPage.current - 1)
    setRefresh(r => !r)
    console.log('>>>> PREV PAGE', currentPage.current)
  }

  const resetPage = () => {
    currentPage.current = 1
  }

  const filterPipelines = useCallback((status) => {
    console.log('>>> GOT PIPELINE COUNT = ', pipelines.length, pipelines.length / perPage)
    resetPage()
    setFilteredPipelines(status === 'all' ? pipelines : pipelines.filter((p) => p.status === status))
    // setMaxPage(pipelines.length <= perPage ? 1 : Math.floor(pipelines.length / perPage))
    setTimeout(() => {
      setIsProcessing(false)
    }, 300)
  }, [pipelines, perPage])

  const paginatePipelines = useCallback(() => {
    const sliceOffset = currentPage.current >= 2 ? -1 : 0
    const sliceBegin = currentPage.current === 1
      ? 0
      : (currentPage.current + sliceOffset) * perPage
    const sliceEnd = currentPage.current === 1
      ? perPage
      : ((currentPage.current + sliceOffset) * perPage) + perPage
    console.log('>> CURRENT PAGE = ', currentPage.current)
    console.log('>> START RECORD INDEX ====', sliceBegin)
    console.log('>> END RECORD INDEX ====', sliceEnd)
    setPagedPipelines(filteredPipelines.slice(sliceBegin, sliceEnd))
  }, [filteredPipelines, perPage])

  const getPipelineCountByStatus = useCallback((status) => {
    return status === 'all' ? pipelines.length : pipelines.filter((p) => p.status === status).length
  }, [pipelines])

  const handleInspectorClose = () => {
    setShowInspector(false)
    setInspectPipeline(null)
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
    fetchAllPipelines()
    return () => {
      currentPage.current = 1
    }
  }, [fetchAllPipelines])

  useEffect(() => {
    console.log('>>> Pipelines', filteredPipelines)
    console.log('>> CURRENT PAGE = ', currentPage.current)
    console.log('>> MAX PAGE = ', maxPage)
    if (pipelines.length > 0) {
      const latestPipelineRun = pipelines[0]
      fetchPipeline(latestPipelineRun.ID)
    }
  }, [pipelines, filteredPipelines, fetchPipeline, currentPage, maxPage, perPage])

  useEffect(() => {

  }, [pipelineCount])

  useEffect(() => {
    console.log('>> FILTER STATUS CHANGED ===> ', activeStatus)
    setIsProcessing(true)
    filterPipelines(activeStatus)
  }, [activeStatus, filterPipelines])

  useEffect(() => {
    console.log('>> PAGINATING PIPELINES...')
    paginatePipelines()
  }, [refresh, perPage, filteredPipelines, paginatePipelines])

  // useEffect(() => {
  //   console.log('>>> LATEST PIPELINE!', latestPipeline)
  // }, [latestPipeline])

  useEffect(() => {
    console.log('>>> FILTERED PIPELINES!', filteredPipelines)
    setMaxPage(filteredPipelines.length <= perPage ? 1 : Math.ceil(filteredPipelines.length / perPage))
    // @todo: JC -- check if pagination call is needed here!
  }, [filteredPipelines, perPage])

  useEffect(() => {
    console.log('>>> PAGED PIPELINES...', pagedPipelines)
  }, [pagedPipelines])

  return (
    <>
      <div className='container'>
        <Nav />
        <Sidebar />
        <Content>
          <main className='main'>
            <AppCrumbs
              items={[
                { href: '/', icon: false, text: 'Dashboard' },
                { href: '/pipelines', icon: false, text: 'Pipelines' },
                { href: '/pipelines', icon: false, text: 'All Pipeline Runs', current: true },
              ]}
            />
            <div className='headlineContainer'>
              {/* <Link style={{ float: 'right', marginLeft: '10px', color: '#777777' }} to='/integrations'>
                <Icon icon='fast-backward' size={16} /> Go Back
              </Link> */}
              <div style={{ display: 'flex' }}>
                <div>
                  <h1 style={{ margin: 0 }}>
                    All Pipeline Runs
                    <Popover
                      className='trigger-manage-pipelines-help'
                      popoverClassName='popover-help-manage-pipelines'
                      position={Position.RIGHT}
                      autoFocus={false}
                      enforceFocus={false}
                      usePortal={false}
                    >
                      <a href='#' rel='noreferrer'><HelpIcon width={19} height={19} style={{ marginLeft: '10px' }} /></a>
                      <>
                        <div style={{ textShadow: 'none', fontSize: '12px', padding: '12px', maxWidth: '300px' }}>
                          <div style={{ marginBottom: '10px', fontWeight: 700, fontSize: '14px' }}>
                            <Icon icon='help' size={16} /> Manage Pipeline Runs
                          </div>
                          <p>Need Help? &mdash; Manage, Stop running and Restart failed pipelines.
                            Access <strong>Task Progress</strong> and Activity for all your pipelines.
                          </p>
                        </div>
                      </>
                    </Popover>
                  </h1>
                  <p className='page-description mb-0'>Manage Job Activity and see duration for all your pipeline runs.</p>
                  <p className=''>The most recent runs are shown first, filter by key status types.</p>
                </div>
                <div style={{ marginLeft: 'auto' }}>
                  <Button icon='add' intent={Intent.PRIMARY} text='Create Pipeline Run' onClick={() => history.push('/pipelines/create')} />
                </div>
              </div>
            </div>

            {(isFetchingAll || !isFetchingAll) && (
              <>
                <div style={{ display: 'flex', marginTop: '30px', minHeight: '36px', width: '100%', justifyContent: 'space-between' }}>

                  <ButtonGroup
                    disabled={isFetchingAll || isProcessing}
                    className='filter-status-group'
                    round='true'
                    style={{ fontSize: '12px', zIndex: 0 }}
                  >
                    <Button
                      className='btn-pipeline-filter'
                      intent={activeStatus === 'all' ? 'primary' : null} onClick={() => setActiveStatus('all')}
                    >
                      <span style={{ marginRight: '10x', letterSpacing: '0', fontWeight: 900 }}>All&nbsp;</span>
                      <Tag className='tag-data-count'>{getPipelineCountByStatus('all')}</Tag>
                    </Button>
                    <Button
                      className='btn-pipeline-filter'
                      intent={activeStatus === 'TASK_RUNNING' ? 'primary' : null} onClick={() => setActiveStatus('TASK_RUNNING')}
                    >
                      <span style={{ marginRight: '10x', letterSpacing: '0', fontWeight: 700 }}>Running&nbsp;</span>
                      <Tag className='tag-data-count'>{getPipelineCountByStatus('TASK_RUNNING')}</Tag>
                    </Button>
                    <Button
                      className='btn-pipeline-filter'
                      intent={activeStatus === 'TASK_COMPLETED' ? 'primary' : null} onClick={() => setActiveStatus('TASK_COMPLETED')}
                    >
                      <span style={{ marginRight: '10x', letterSpacing: '0', fontWeight: 700 }}>Complete&nbsp;</span>
                      <Tag className='tag-data-count'>{getPipelineCountByStatus('TASK_COMPLETED')}</Tag>
                    </Button>
                    <Button
                      className='btn-pipeline-filter'
                      intent={activeStatus === 'TASK_FAILED' ? 'primary' : null} onClick={() => setActiveStatus('TASK_FAILED')}
                    >
                      <span style={{
                        display: 'flex',
                        justifyContent: 'center',
                        alignItems: 'center',
                        marginRight: '10x',
                        letterSpacing: '0',
                        fontWeight: 700
                      }}
                      >
                        <Icon icon='warning-sign' size={14} style={{ justifySelf: 'center', marginRight: '10px' }} />
                        Failed&nbsp;
                        <Tag className='tag-data-count'>{getPipelineCountByStatus('TASK_FAILED')}</Tag>

                      </span>
                    </Button>
                  </ButtonGroup>
                  {isProcessing && (
                    <Button minimal style={{ marginRight: 'auto' }}>
                      <Spinner size={18} />
                    </Button>
                  )}
                  {/* @todo: Reactivate search input & enable feature */}
                  {/* <InputGroup
                    leftElement={<Icon icon='search' />}
                    placeholder='Search Pipelines'
                    rightElement={<Button text='GO' intent='primary' />}
                    round
                  /> */}
                </div>

                <div style={{ display: 'flex', width: '100%', justifySelf: 'flex-start', marginTop: '8px' }}>
                  <Card
                    className='manage-pipelines-card'
                    interactive={false}
                    elevation={Elevation.TWO}
                    style={{ width: '100%', minWidth: '836px', padding: '2px' }}
                  >
                    <table className='bp3-html-table bp3-html-table-bordered pipelines-table' style={{ width: '100%' }}>
                      <thead>
                        <tr>
                          <th style={{ minWidth: '80px', maxWidth: '80px', whiteSpace: 'nowrap' }}>
                            <Icon icon='sort-desc' color='#aaa' size={10} style={{ marginRight: '3px', marginBottom: '3px' }} /> ID
                          </th>
                          <th style={{ width: '100%' }}>Pipeline Name</th>
                          <th style={{ minWidth: '94px', whiteSpace: 'nowrap' }}>Duration</th>
                          <th style={{ minWidth: '104px', whiteSpace: 'nowrap' }}>Status</th>
                          <th style={{ minWidth: '92px', whiteSpace: 'nowrap' }} />
                        </tr>
                      </thead>
                      <tbody>
                        {!isFetchingAll && pagedPipelines.length > 0 && pagedPipelines.map((pipeline, pIdx) => (
                          <tr
                            key={`pipeline-row-${pIdx}`}
                            className={pipeline?.status === 'TASK_FAILED' ? 'pipeline-row pipeline-failed' : 'pipeline-row'}
                            style={{ verticalAlign: 'middle' }}
                          >
                            <td
                              style={{ cursor: 'pointer' }}
                              className='cell-id'
                            >
                              <Tooltip content={`Pipeline Run ID #${pipeline.ID}`} position={Position.TOP}>
                                <a
                                  href='#'
                                  style={{ fontWeight: inspectPipeline?.ID === pipeline.ID ? 800 : 'normal' }}
                                  onClick={() => history.push(`/pipelines/activity/${pipeline.ID}`)}
                                >
                                  {pipeline.ID}
                                </a>
                              </Tooltip>
                              {inspectPipeline?.ID === pipeline.ID && (
                                <Icon icon='menu-open' color='#E8471C' size={12} style={{ margin: '0 5px 0 0', float: 'left' }} />
                              )}
                            </td>

                            <td
                              onClick={(e) => history.push(`/pipelines/activity/${pipeline.ID}`)}
                              style={{ cursor: 'pointer' }}
                              className='cell-name'
                            >
                              <span style={{
                                display: 'inline-block',
                                float: 'right',
                                color: '#999999',
                                marginLeft: '15px'
                              }}
                              >{dayjs(pipeline.createdAt).format('L LTS')}
                              </span>

                              <strong style={{
                                lineHeight: '100%',
                                fontSize: '12px',
                                fontWeight: 800,
                                textOverflow: 'ellipsis',
                                overflow: 'hidden',
                                display: 'block',
                                whiteSpace: 'nowrap',
                                maxWidth: '100%'
                              }}
                              >

                                {pipeline.name}
                                {parseInt(pipeline.blueprintId, 10) > 0 && (
                                  <Tooltip content={`Blueprint ID ${pipeline.blueprintId}`}>
                                    <Icon icon='bold' color={Colors.BLUE4} size={12} style={{ margin: '3px' }} />
                                  </Tooltip>
                                )}
                                {pipeline.status === 'TASK_COMPLETED' && (
                                  <Icon
                                    icon='tick' size={10} color={Colors.GREEN5}
                                    style={{ margin: '0 10px', float: 'right', marginBottom: '2px' }}
                                  />
                                )}
                              </strong>

                            </td>
                            <td
                              className='cell-duration no-user-select'
                              style={{ cursor: 'pointer', whiteSpace: 'nowrap' }}
                            >

                              {/* {dayjs(pipeline.CreatedAt).toNow(pipeline.CreatedAt)} */}
                              {pipeline.status === 'TASK_RUNNING' ? dayjs(pipeline.beganAt).toNow(true) : pipeline.finishedAt == null ? 'N/A' : dayjs(pipeline.finishedAt).from(pipeline.beganAt, true)}
                            </td>
                            <td className='cell-status no-user-select' style={{ textTransform: 'uppercase', whiteSpace: 'nowrap' }}>
                              <span style={{ display: 'inline-block', float: 'left', marginRight: '10px' }}>
                                <Tooltip content={`Progress ${pipeline.finishedTasks}/${pipeline.totalTasks} Tasks`}>
                                  {pipeline.status === 'TASK_RUNNING' &&
                                  (
                                    <Spinner
                                      style={{ margin: 0 }}
                                      className='mini-task-spinner' size={14} intent='warning'
                                      value={Number(pipeline.finishedTasks / pipeline.totalTasks).toFixed(1)}
                                    />
                                  )}
                                </Tooltip>
                                {pipeline.status === 'TASK_COMPLETED' &&
                                 (
                                   <Tooltip
                                     intent={Intent.SUCCESS}
                                     content={`Finished ${pipeline.finishedTasks}/${pipeline.totalTasks} Tasks`}
                                   >
                                     <Spinner
                                       style={{ margin: 0 }}
                                       className='mini-task-spinner' size={14} intent='success' value={1}
                                     />
                                   </Tooltip>
                                 )}
                                {pipeline.status === 'TASK_FAILED' &&
                                 (
                                   <Tooltip
                                     intent={Intent.PRIMARY}
                                     content={`Failed ${pipeline.finishedTasks}/${pipeline.totalTasks} Tasks`}
                                   >
                                     <Spinner
                                       style={{ margin: 0 }} className='mini-task-spinner'
                                       size={14} intent='info' value={Number(pipeline.finishedTasks / pipeline.totalTasks).toFixed(1)}
                                     />
                                   </Tooltip>
                                 )}
                              </span>
                              {pipeline.status === 'TASK_FAILED' && (
                                <strong style={{ color: Colors.RED5 }}>Failed</strong>
                              )}
                              {pipeline.status === 'TASK_COMPLETED' && (
                                <strong style={{ color: Colors.GREEN5 }}>Complete</strong>
                              )}
                              {pipeline.status === 'TASK_RUNNING' && (
                                <strong style={{ color: Colors.BLUE5 }}>Running</strong>
                              )}
                              {pipeline.status === 'TASK_CREATED' && (
                                <strong style={{ color: Colors.GRAY3 }}>
                                  <Icon icon='time' size={12} color={Colors.GRAY2} style={{ marginBottom: '3px' }} /> Pending...
                                </strong>
                              )}
                            </td>
                            <td className='cell-actions no-user-select' style={{ padding: '0 10px', verticalAlign: 'middle' }}>
                              <div style={{
                                display: 'flex',
                                justifySelf: 'center',
                                gap: '3px',
                                alignSelf: 'center',
                                alignContent: 'center',
                                alignItems: 'center'
                              }}
                              >
                                <a
                                  href='#'
                                  onClick={() => history.push(`/pipelines/activity/${pipeline.ID}`)}
                                  data-provider={pipeline.id}
                                  className='bp3-button bp3-small bp3-minimal'
                                >
                                  <Icon icon='eye-open' size={16} />
                                </a>
                                {['TASK_FAILED', 'TASK_COMPLETED'].includes(pipeline.status) && (
                                  <a
                                    href='#'
                                    onClick={() => restartPipeline(pipeline.tasks.flat())}
                                    data-provider={pipeline.id}
                                    className='bp3-button bp3-small bp3-minimal'
                                  >
                                    <Icon icon='refresh' size={12} />
                                  </a>)}
                                {pipeline.status === 'TASK_RUNNING' && (

                                  <Popover
                                    key={`popover-help-key-cancel-run-${pipeline.ID}`}
                                    className='trigger-pipeline-cancel'
                                    popoverClassName='popover-pipeline-cancel'
                                    position={Position.BOTTOM}
                                    autoFocus={false}
                                    enforceFocus={false}
                                    usePortal={true}
                                    disabled={pipeline.status !== 'TASK_RUNNING'}
                                  >
                                    <a
                                      href='#'
                                      data-provider={pipeline.id}
                                      className='bp3-button bp3-small bp3-minimal'
                                    >
                                      <Icon icon='stop' size={16} style={{ color: Colors.RED5 }} />
                                    </a>
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
                                            onClick={() => cancelPipeline(pipeline.ID)}
                                          />
                                        </div>
                                      </div>
                                    </>
                                  </Popover>

                                )}
                                <a
                                  href='#'
                                  onClick={() => setInspectPipeline(pipeline) | setShowInspector(true)}
                                  data-provider={pipeline.id}
                                  className='bp3-button bp3-small bp3-minimal'
                                >
                                  <Icon icon='code' size={16} />
                                </a>
                              </div>
                            </td>
                          </tr>
                        ))}
                        {isFetchingAll && (
                          <tr>
                            <td className='loading-cell' colSpan='5' style={{ backgroundColor: '#f8f8f8' }}>
                              <ContentLoader
                                elevation={Elevation.ZERO}
                                cardStyle={{ border: '0 !important', boxShadow: 'none', backgroundColor: 'transparent' }}
                                title='Loading Pipeline Runs ...' message='Please wait while the data records are loaded.'
                              />
                            </td>
                          </tr>
                        )}

                        {!isFetchingAll && filteredPipelines.length === 0 && (
                          <tr>
                            <td className='no-data-message-cell' colSpan='5' style={{ backgroundColor: '#fffcf0' }}>
                              <h3 style={{
                                fontWeight: 800,
                                letterSpacing: '2px',
                                textTransform: 'uppercase',
                                margin: 0,
                              }}
                              >0 Pipeline Runs
                              </h3>
                              <p style={{ margin: 0 }}>There are no pipeline logs for the current status
                                {' '}<strong>{activeStatus}</strong>.
                              </p>
                            </td>
                          </tr>
                        )}
                      </tbody>
                    </table>
                  </Card>

                </div>
                <div
                  className='operations panel no-user-select'
                  style={{
                    marginTop: '10px',
                    display: 'flex',
                    width: '100%',
                    justifySelf: 'flex-start',
                    whiteSpace: 'nowrap',
                  }}
                >

                  <div className='no-user-elect' style={{ display: 'flex', width: '50%', fontSize: '11px', color: '#555555' }}>
                    <Icon icon='user' size={14} style={{ marginRight: '8px' }} />
                    <div>
                      <span>by {' '} <strong>Administrator</strong></span><br />

                      <span style={{ color: '#888888' }}>Displaying{' '}
                        {filteredPipelines.length === 0 && (<>0</>)}
                        {filteredPipelines.length > 0 && (
                          <>
                            {currentPage.current === 1 && filteredPipelines.length <= perPage && (
                              <>1 - {filteredPipelines.length}</>
                            )}
                            {currentPage.current === 1 && filteredPipelines.length > perPage && (
                              <>1 - {perPage}</>
                            )}
                            {currentPage.current > 1 && filteredPipelines.length > perPage && (
                              <>
                                {(perPage * currentPage.current) - perPage} - {currentPage.current === 1
                                  ? perPage
                                  : (Math.min(filteredPipelines.length, perPage * currentPage.current))}
                              </>
                            )}
                            {' '} of {' '}
                            <Tooltip
                              content={`Page ${currentPage.current.toString()} of ${maxPage}`}
                            >
                              <strong>{filteredPipelines.length}</strong>
                            </Tooltip>

                          </>
                        )}
                        {' '}pipeline runs from API.
                      </span>

                    </div>
                  </div>

                  <div style={{ display: 'flex', marginLeft: 'auto', marginRight: '30px' }}>

                    <Button small icon='add' style={{ marginRight: '5px' }} minimal onClick={() => history.push('/pipelines/create')} />
                    <Button
                      icon='refresh'
                      text='Refresh'
                      onClick={() => fetchAllPipelines()}
                      minimal
                      small
                    />
                  </div>
                  <div className='pagination-controls' style={{ display: 'flex', whiteSpace: 'nowrap' }}>
                    <Popover placement='bottom'>
                      <Button
                        className='btn-select-page-size'
                        style={{ whiteSpace: 'nowrap' }}
                        icon='numbered-list'
                        text={`Rows: ${perPage}`}
                        disabled={isFetchingAll}
                        outlined
                        minimal
                      />
                      <>
                        {pagingOptionsMenu}
                      </>
                    </Popover>
                    <Button
                      onClick={prevPage}
                      className='pagination-btn btn-prev-page'
                      icon='step-backward' small text='PREV'
                      style={{ marginLeft: '5px', marginRight: '5px', whiteSpace: 'nowrap' }}
                      disabled={currentPage.current === 1 || isFetchingAll}
                    />
                    <Button
                      style={{ whiteSpace: 'nowrap' }}
                      disabled={currentPage.current === maxPage || isFetchingAll}
                      onClick={nextPage}
                      className='pagination-btn btn-next-page'
                      rightIcon='step-forward'
                      text='NEXT'
                      small
                    />
                  </div>
                </div>
                <div style={{ height: '50px' }} />
              </>
            )}
          </main>
        </Content>
      </div>
      {!isFetchingAll &&
      activePipeline &&
      (
        <PipelineIndicator
          isVisible={!showInspector}
          pipeline={activePipeline}
          graphsUrl={GRAFANA_URL}
          onFetch={fetchPipeline}
          onCancel={cancelPipeline}
          onView={() => history.push(`/pipelines/activity/${activePipeline.ID}`)}
          onRetry={() => restartPipeline(activePipeline.tasks)}
        />)}
      {!isFetchingAll &&
      inspectPipeline &&
      (
        <CodeInspector
          isOpen={showInspector}
          activePipeline={inspectPipeline}
          onClose={handleInspectorClose}
          hasBackdrop={false}
        />)}

    </>
  )
}

export default Pipelines
