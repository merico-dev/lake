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
import React, { useEffect, useCallback } from 'react'
import dayjs from '@/utils/time'
import {
  ButtonGroup,
  Button, Icon, Intent,
  Card, Elevation,
  PopoverInteractionKind,
  Popover,
  Tooltip,
  Position,
  Spinner,
  Colors,
  Collapse,
  Divider,
  Switch,
  Tag
} from '@blueprintjs/core'
import DeletePopover from '@/components/blueprints/DeletePopover'

const BlueprintsGrid = (props) => {
  const {
    blueprints = [],
    filteredBlueprints = [],
    pipelines = [],
    activeBlueprint,
    blueprintSchedule,
    isActiveBlueprint = (b) => {},
    expandBlueprint = (b) => {},
    deleteBlueprint = (b) => {},
    // eslint-disable-next-line no-unused-vars
    createCron = () => {},
    handleBlueprintActivation = (b) => {},
    configureBlueprint = (b) => {},
    getNextRunDate = (b) => {},
    isDeleting = false,
    isLoading = false,
    expandDetails = false,
    cronPresets,
    activeFilterStatus,
    onFilter = () => {},
    onViewPipeline = () => {}
  } = props

  // eslint-disable-next-line no-unused-vars
  const getCronPreset = useCallback((presetName) => {
    return cronPresets.find(p => p.name === presetName)
  }, [cronPresets])

  useEffect(() => {
    console.log('>> FILTER BLUEPRINTS BY INTERVAL....', activeFilterStatus)
  }, [activeFilterStatus])

  return (
    <>
      <div
        className='blueprints-filter-panel'
        style={{
          marginTop: '30px',
          display: 'flex',
          alignSelf: 'flex-start',
          alignContent: 'flex-start'
        }}
      >
        <ButtonGroup className='filter-status-group blueprints-filter-group' style={{ zIndex: 0 }}>
          <Button
            intent={!activeFilterStatus ? Intent.PRIMARY : Intent.NONE}
            active={!activeFilterStatus}
            onClick={() => onFilter(null)}
          >All
          </Button>
          <Button
            intent={activeFilterStatus === 'hourly' ? Intent.PRIMARY : Intent.NONE}
            active={activeFilterStatus === 'hourly'}
            onClick={() => onFilter('hourly')}
          >
            Hourly
          </Button>
          <Button
            intent={activeFilterStatus === 'daily' ? Intent.PRIMARY : Intent.NONE}
            active={activeFilterStatus === 'daily'}
            onClick={() => onFilter('daily')}
          >
            Daily
          </Button>
          <Button
            intent={activeFilterStatus === 'weekly' ? Intent.PRIMARY : Intent.NONE}
            active={activeFilterStatus === 'weekly'}
            onClick={() => onFilter('weekly')}
          >
            Weekly
          </Button>
          <Button
            intent={activeFilterStatus === 'monthly' ? Intent.PRIMARY : Intent.NONE}
            active={activeFilterStatus === 'monthly'}
            onClick={() => onFilter('monthly')}
          >
            Monthly
          </Button>
          <Button
            intent={activeFilterStatus === 'custom' ? Intent.PRIMARY : Intent.NONE}
            active={activeFilterStatus === 'custom'}
            onClick={() => onFilter('custom')}
          >
            Custom
          </Button>
        </ButtonGroup>
      </div>
      <Card
        className='blueprints-grid-card'
        elevation={expandDetails ? Elevation.ZERO : Elevation.TWO}
        style={{
          display: 'flex',
          marginTop: '8px',
          minHeight: '36px',
          width: '100%',
          padding: 0,
          justifyContent: 'flex-start',
          boxShadow: expandDetails ? 'none' : '',
          overflow: 'hidden'
        }}
      >
        <div
          className='blueprints-list-grid' style={{
            display: 'flex',
            flexDirection: 'column',
            width: '100%',
            minWidth: '830px'
          }}
        >
          {(activeFilterStatus ? filteredBlueprints : blueprints).map((b, bIdx) => (
            <div key={`blueprint-row-key-${bIdx}`}>
              <div
                style={{
                  display: 'flex',
                  width: '100%',
                  minHeight: '48px',
                  borderBottom: isActiveBlueprint(b.id) && expandDetails ? 'none' : '1px solid #eee',
                  backgroundColor: !b.enable ? '#f8f8f8' : 'inherit',
                  color: !b.enable ? '#555555' : 'inherit',
                }}
              >
                <div
                  className='blueprint-row' style={{
                    display: 'flex',
                    width: '100%',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '8px 5px',
                    paddingBottom: '16px',
                    position: 'relative'
                  }}
                >
                  <div className='blueprint-id' style={{ flex: 1, maxWidth: '100px' }}>
                    <div style={{ height: '24px', lineHeight: '24px' }}>
                      <label style={{
                        marginLeft: '25px',
                        fontSize: '9px',
                        fontWeight: '400',
                        color: '#777777'
                      }}
                      >
                        ID
                      </label>
                    </div>
                    <Button
                      className='bp-row-expand-trigger'
                      onClick={() => expandBlueprint(b)}
                      small minimal style={{
                        minHeight: '20px',
                        minWidth: '20px',
                        marginTop: '-3px',
                        padding: 0,
                        marginRight: '5px',
                        float: 'left'
                      }}
                    >
                      <Icon
                        size={12} color={isActiveBlueprint(b.id) && expandDetails ? Colors.BLUE3 : Colors.GRAY2}
                        icon={isActiveBlueprint(b.id) && expandDetails ? 'collapse-all' : 'expand-all'}
                        style={{ margin: '0' }}
                      />
                    </Button>
                    {b.id}
                  </div>
                  <div
                    className='blueprint-name'
                    style={{ flex: 2, minWidth: '176px', fontWeight: 800, cursor: 'pointer' }}
                    onClick={() => expandBlueprint(b)}
                  >
                    <div style={{ height: '24px', lineHeight: '24px' }}>
                      <label style={{
                        fontSize: '9px',
                        fontWeight: '400',
                        color: '#777777'
                      }}
                      >
                        Blueprint Name
                      </label>
                    </div>
                    <Icon
                      size={16}
                      icon={b.enable ? 'calendar' : 'disable'}
                    />
                    {b.name}
                  </div>
                  <div className='blueprint-interval' style={{ flex: 1, minWidth: '60px' }}>
                    <div style={{ height: '24px', lineHeight: '24px' }}>
                      <label style={{
                        fontSize: '9px',
                        fontWeight: '400',
                        color: '#777777'
                      }}
                      >
                        Frequency
                      </label>
                    </div>
                    {b.interval}
                  </div>
                  <div className='blueprint-next-rundate' style={{ flex: 1, whiteSpace: 'nowrap' }}>
                    <div style={{ height: '24px', lineHeight: '24px' }}>
                      <label style={{
                        fontSize: '9px',
                        fontWeight: '400',
                        color: '#777777'
                      }}
                      >
                        Next Run Date
                      </label>
                    </div>
                    <div>
                      {dayjs(getNextRunDate(b.cronConfig)).format('L LTS')}
                    </div>
                    <div>
                      <span style={{ color: b.enable ? Colors.GREEN5 : Colors.GRAY3, position: 'absolute', bottom: '4px' }}>
                        {b.cronConfig}
                      </span>
                    </div>
                  </div>
                  <div className='blueprint-actions' style={{ flex: 1, textAlign: 'right' }}>
                    <div style={{ height: '24px', lineHeight: '24px' }}>
                      <label style={{
                        fontSize: '9px',
                        fontWeight: '400',
                        color: '#777777'
                      }}
                      >
                   &nbsp;
                      </label>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', justifySelf: 'flex-end' }}>
                      <Button small minimal style={{ marginLeft: 'auto', marginRight: '5px' }} onClick={() => configureBlueprint(b)}>
                        <Tooltip
                          content='Blueprint Settings'
                          interactionKind={PopoverInteractionKind.HOVER}
                          openOnTargetFocus={false}
                          enforceFocus={false}
                          autoFocus={false}
                        >
                          <Icon icon='cog' size={16} color={Colors.GRAY3} />
                        </Tooltip>
                      </Button>
                      <Popover position={Position.LEFT}>
                        <Button small minimal style={{ marginRight: '10px' }}>
                          <Icon icon='trash' color={Colors.GRAY3} size={15} />
                        </Button>
                        <DeletePopover
                          activeBlueprint={b}
                          onCancel={() => {}}
                          onConfirm={deleteBlueprint}
                          isRunning={isDeleting}
                        />
                      </Popover>

                      <Switch
                        checked={b.enable}
                        label={false}
                        onChange={() => handleBlueprintActivation(b)}
                        style={{ marginBottom: '0' }}
                      />
                    </div>
                  </div>
                </div>
              </div>
              <Collapse isOpen={expandDetails && activeBlueprint.id === b.id}>
                <Card
                  elevation={Elevation.TWO} style={{
                    padding: '0',
                    margin: '30px 30px',
                    minWidth: '860px',
                    backgroundColor: !b.enable ? '#f8f8f8' : 'initial'
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', margin: '0', padding: '10px' }}>
                    <div style={{ letterSpacing: '1px', fontWeight: 800 }}>
                      <Icon icon='bold' color={Colors.BLUE4} size={14} style={{ marginRight: '5px' }} /> BLUEPRINT ID {activeBlueprint?.id}
                      {isLoading && (
                        <span style={{ paddingLeft: '20px', fontWeight: 700, color: '#777777', fontSize: '11px' }}>
                          <span style={{ display: 'inline-block', marginRight: '10px', marginTop: '2px' }}>
                            <Spinner size={12} />
                          </span>
                          LOADING ASSOCIATED PIPELINES...
                        </span>
                      )}
                    </div>
                    <div>
                      <Tag
                        style={{
                          backgroundColor: b.enable
                            ? Colors.GREEN3
                            : Colors.GRAY3
                        }} round='true'
                      >
                        {b.enable ? 'ACTIVE' : 'INACTIVE'}
                      </Tag>
                    </div>
                  </div>
                  <Divider style={{ marginRight: 0, marginLeft: 0 }} />
                  <div style={{ padding: '20px', display: 'flex' }}>
                    <div style={{ flex: 2, paddingRight: '30px' }}>
                      <h3 style={{ margin: 0, textTransform: 'uppercase' }}>Pipeline Run Schedule</h3>
                      <p style={{ margin: 0 }}>
                        Based on the current CRON settings, here are next{' '}
                        <strong>5</strong> expected pipeline collection dates.
                      </p>
                      <div style={{ margin: '10px 0' }}>
                        {activeBlueprint?.id && blueprintSchedule.map((s, sIdx) => (
                          <div key={`run-schedule-event-key${sIdx}`} style={{ padding: '6px 4px', opacity: b.enable ? 1 : 0.5 }}>
                            <Icon
                              icon='calendar' size={14}
                              color={b.enable ? Colors.BLUE4 : Colors.GRAY4}
                              style={{ marginRight: '10px' }}
                            />
                            {dayjs(s).format('L LTS')}
                          </div>
                        ))}
                      </div>
                      <div className='related-pipelines-list' style={{ marginBottom: '20px' }}>
                        {!isLoading && (
                          <h3 style={{ margin: '0 0 5px 0', textTransform: 'uppercase' }}>
                            Pipeline Runs <small style={{ color: Colors.GRAY5 }}>(last 5)</small>
                          </h3>
                        )}
                        {!isLoading && pipelines.length === 0 && (<p>No Pipelines have been found for this blueprint.</p>)}
                        {!isLoading && pipelines.slice(0, 5).map((p, pIdx) => (
                          <div
                            key={`pipeline-run-key-${pIdx}`}
                            className='pipeline-run-entry'
                            style={{ display: 'flex', justifyContent: 'flex-start', alignItems: 'center' }}
                          >
                            <div
                              className='pipeline-id'
                              style={{ minWidth: '60px', paddingRight: '15px', fontWeight: 'bold' }}
                            >
                              #{p.id}
                            </div>
                            <div className='pipeline-created' style={{ minWidth: '180px', paddingRight: '15px' }}>
                              {dayjs(p.createdAt).format('L LTS')}
                            </div>
                            <div className='pipeline-name' style={{ flex: 1, paddingRight: '15px' }}>
                              <span
                                style={{
                                  display: 'block',
                                  whiteSpace: 'wrap',
                                  maxWidth: '300px',
                                  textOverflow: 'ellipsis',
                                  overflow: 'hidden'
                                }}
                              >{p.name}
                              </span>
                            </div>
                            <div style={{ paddingRight: '15px', color: Colors.GRAY2, whiteSpace: 'nowrap' }}>
                              {p.status === 'TASK_RUNNING'
                                ? dayjs(p.createdAt).toNow(true)
                                : dayjs(p.updatedAt).from(p.createdAt, true)}
                            </div>
                            <div>{p.status?.replace('TASK_', '')}</div>
                            <div style={{ padding: '0 15px' }}>
                              <Button
                                onClick={() => onViewPipeline(p.id || p.ID)}
                                icon='eye-open'
                                size={14}
                                color={Colors.GRAY3}
                                small minimal
                              />
                            </div>
                          </div>
                        ))}
                      </div>

                      {!b.enable && (
                        <p style={{ marginTop: '10px 0 0 0', fontSize: '9px' }}>
                          <Icon icon='warning-sign' size={11} color={Colors.ORANGE5} style={{ float: 'left', marginRight: '5px' }} />
                          Blueprint is NOT Enabled / Active this schedule will not run.
                        </p>
                      )}
                    </div>
                    <div style={{ flex: 1 }}>
                      <label style={{ color: Colors.GRAY1 }}>Blueprint</label>
                      <h3 style={{ marginTop: 0, fontSize: '18px', fontWeight: 800 }}>
                        {b.name}
                      </h3>
                      <label style={{ color: Colors.GRAY1 }}>Crontab Configuration</label>
                      <h3 style={{ margin: '0 0 20px 0', fontSize: '18px' }}>{b.cronConfig}</h3>

                      <label style={{ color: Colors.GRAY1 }}>Next Run</label>
                      <h3 style={{ margin: '0 0 20px 0', fontSize: '18px' }}>
                        {dayjs(getNextRunDate(b.cronConfig)).fromNow()}
                      </h3>

                      <label style={{ color: Colors.GRAY3 }}>Operations</label>
                      <div style={{
                        marginTop: '5px',
                        display: 'flex',
                        justifySelf: 'flex-start',
                        alignItems: 'center',
                        justifyContent: 'left',
                        fontSize: '10px'
                      }}
                      >
                        <Button
                          intent={Intent.PRIMARY}
                          icon='cog'
                          text='Settings'
                          small
                          style={{ marginRight: '8px' }}
                          onClick={() => configureBlueprint(b)}
                        />
                        <Popover>
                          <Button icon='trash' text='Delete' small minimal style={{ marginRight: '8px' }} />
                          <DeletePopover
                            activeBlueprint={activeBlueprint}
                            onCancel={() => {}}
                            onConfirm={deleteBlueprint}
                            isRunning={isDeleting}
                          />
                        </Popover>
                        <Switch
                          checked={b.enable}
                          label={b.enable ? 'Disable' : 'Enable'}
                          onChange={() => handleBlueprintActivation(b)}
                          style={{ marginBottom: '0', fontSize: '11px' }}
                        />
                      </div>
                    </div>

                  </div>
                </Card>

              </Collapse>
            </div>
          ))}
          {(activeFilterStatus ? filteredBlueprints.length === 0 : blueprints.length === 0) && (
            <div style={{ padding: '12px' }}>
              <h3 style={{
                fontWeight: 800,
                letterSpacing: '2px',
                textTransform: 'uppercase',
                margin: 0,
              }}
              >0 Blueprints
              </h3>
              <p style={{ margin: 0 }}>There are no blueprints for the current status
                {' '}<strong>{activeFilterStatus}</strong>.
              </p>
            </div>
          )}
        </div>
      </Card>
      <div style={{
        display: 'flex',
        margin: '20px 10px',
        alignSelf: 'flex-start',
        width: '50%',
        fontSize: '11px',
        color: '#555555'
      }}
      >
        <Icon icon='user' size={14} style={{ marginRight: '8px' }} />
        <div>
          <span>by {' '} <strong>Administrator</strong></span><br />
          Displaying {activeFilterStatus ? filteredBlueprints.length : blueprints.length} Blueprints from API.
        </div>
      </div>
    </>
  )
}

export default BlueprintsGrid
